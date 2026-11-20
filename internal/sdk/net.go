package sdk

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/muriloprimata/uwp4me/internal/ui"
)

func fileSHA(path, want string) (bool, error) {
    f, err := os.Open(path)
    if err != nil {
        return false, err
    }
    defer f.Close()
    h := sha256.New()
    if _, err := io.Copy(h, f); err != nil {
        return false, err
    }
    return strings.EqualFold(hex.EncodeToString(h.Sum(nil)), want), nil
}

// dlTry downloads (or resumes) link into tmp. Integrity is checked by SHA later.
func dlTry(link, tmp string) error {
    var offset int64
    if st, err := os.Stat(tmp); err == nil {
        offset = st.Size()
    }
    req, err := http.NewRequest("GET", link, nil)
    if err != nil {
        return err
    }
    req.Header.Set("User-Agent", "uwp4me")
    if offset > 0 {
        req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
    }
    client := &http.Client{Timeout: 30 * time.Minute}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    flags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
    switch resp.StatusCode {
    case 200:
        flags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
    case 206:
    case 416:
        return nil // the .part is already complete; SHA decides
    default:
        return fmt.Errorf("HTTP %d", resp.StatusCode)
    }
    f, err := os.OpenFile(tmp, flags, 0o644)
    if err != nil {
        return err
    }
    _, err = io.Copy(f, resp.Body)
    f.Close()
    return err
}

func dl(link, dst, sha string, size int64) error {
    if st, err := os.Stat(dst); err == nil {
        if sha != "" {
            if ok, _ := fileSHA(dst, sha); ok {
                return nil
            }
            os.Remove(dst)
        } else if st.Size() == size {
            return nil
        }
    }
    if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
        return err
    }
    tmp := dst + ".part"
    var last error
    for attempt := 1; attempt <= 6; attempt++ {
        last = dlTry(link, tmp)
        if last == nil {
            if sha == "" {
                return os.Rename(tmp, dst)
            }
            ok, err := fileSHA(tmp, sha)
            if err != nil {
                return err
            }
            if ok {
                return os.Rename(tmp, dst)
            }
            os.Remove(tmp)
            last = fmt.Errorf("sha256 mismatch")
        }
        ui.Warn(fmt.Sprintf("%s: %v, retrying (%d/6)", filepath.Base(dst), last, attempt))
        time.Sleep(time.Duration(attempt) * 2 * time.Second)
    }
    return fmt.Errorf("%s: %w", filepath.Base(dst), last)
}
