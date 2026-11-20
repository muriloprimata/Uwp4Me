package sdk

import (
    "archive/zip"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/muriloprimata/uwp4me/internal/ui"
)

// Lowercase file names of the SDK installers we want (x64 UWP target).
var wantMSI = map[string]bool{
    "windows sdk desktop headers x86-x86_en-us.msi":                      true,
    "windows sdk desktop headers x64-x86_en-us.msi":                      true,
    "windows sdk desktop libs x64-x86_en-us.msi":                         true,
    "windows sdk for windows store apps headers-x86_en-us.msi":           true,
    "windows sdk for windows store apps libs-x86_en-us.msi":              true,
    "windows sdk for windows store apps headers onecoreuap-x86_en-us.msi": true,
    "windows sdk onecoreuap headers x64-x86_en-us.msi":                   true,
    "universal crt headers libraries and sources-x86_en-us.msi":          true,
}

// Run is the entry point of the "sdk" command.
func runOld(home string, args []string) error {
    for _, a := range args {
        if a == "plan" {
            return Plan(home)
        }
    }
    if err := Fetch(home); err != nil {
        return err
    }
    return Extract(home)
}

func safeJoin(base, rel string) (string, bool) {
    p := filepath.Join(base, filepath.FromSlash(rel))
    b := filepath.Clean(base) + string(os.PathSeparator)
    if !strings.HasPrefix(p, b) {
        return "", false
    }
    return p, true
}

func dlOld(link, dst, sha string, size int64) error {
    if st, err := os.Stat(dst); err == nil && st.Size() == size {
        return nil
    }
    if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
        return err
    }
    client := &http.Client{Timeout: 30 * time.Minute}
    req, err := http.NewRequest("GET", link, nil)
    if err != nil {
        return err
    }
    req.Header.Set("User-Agent", "uwp4me")
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return fmt.Errorf("GET %s: HTTP %d", link, resp.StatusCode)
    }
    tmp := dst + ".part"
    f, err := os.Create(tmp)
    if err != nil {
        return err
    }
    h := sha256.New()
    _, err = io.Copy(io.MultiWriter(f, h), resp.Body)
    f.Close()
    if err != nil {
        return err
    }
    if sha != "" && !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), sha) {
        os.Remove(tmp)
        return fmt.Errorf("sha256 mismatch for %s", filepath.Base(dst))
    }
    return os.Rename(tmp, dst)
}

func unzipVSIX(src, dstDir string) (int, error) {
    r, err := zip.OpenReader(src)
    if err != nil {
        return 0, err
    }
    defer r.Close()
    n := 0
    for _, f := range r.File {
        if !strings.HasPrefix(f.Name, "Contents/") {
            continue
        }
        rel := strings.TrimPrefix(f.Name, "Contents/")
        if dec, err := url.PathUnescape(rel); err == nil {
            rel = dec
        }
        if rel == "" {
            continue
        }
        dst, ok := safeJoin(dstDir, rel)
        if !ok {
            continue
        }
        if strings.HasSuffix(f.Name, "/") {
            os.MkdirAll(dst, 0o755)
            continue
        }
        if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
            return n, err
        }
        rc, err := f.Open()
        if err != nil {
            return n, err
        }
        out, err := os.Create(dst)
        if err != nil {
            rc.Close()
            return n, err
        }
        _, err = io.Copy(out, rc)
        out.Close()
        rc.Close()
        if err != nil {
            return n, err
        }
        n++
    }
    return n, nil
}

// Fetch downloads the CRT and the chosen SDK installers.
func Fetch(home string) error {
    cache := filepath.Join(home, "cache")
    vsPath := filepath.Join(cache, "vs.manifest.json")
    if _, err := os.Stat(vsPath); err != nil {
        if err := Plan(home); err != nil {
            return err
        }
    }
    var m vsman
    if err := readJSON(vsPath, &m); err != nil {
        return fmt.Errorf("package manifest json: %w", err)
    }
    best := map[string]pkg{}
    for _, p := range m.Packages {
        k := kind(p.ID)
        if k == "" {
            continue
        }
        if cur, ok := best[k]; !ok || verLess(cur.Version, p.Version) {
            best[k] = p
        }
    }

    ui.Step("Downloading the C runtime (CRT)")
    for _, k := range []string{"vc:crt.headers.base", "vc:crt.x64.desktop.base", "vc:crt.x64.store.base"} {
        p, ok := best[k]
        if !ok {
            return fmt.Errorf("package not found: %s", k)
        }
        tag := strings.TrimPrefix(k, "vc:")
        for _, pl := range p.Payloads {
            name := filepath.Base(filepath.FromSlash(pl.FileName))
            dst := filepath.Join(cache, "dl", "crt", tag, name)
            ui.Info(fmt.Sprintf("%s (%.1f MB)", name, float64(pl.Size)/1e6))
            if err := dl(pl.URL, dst, pl.SHA256, pl.Size); err != nil {
                return err
            }
            if strings.HasSuffix(strings.ToLower(name), ".vsix") {
                n, err := unzipVSIX(dst, filepath.Join(cache, "crt", tag))
                if err != nil {
                    return err
                }
                ui.OK(fmt.Sprintf("%d files extracted to cache/crt/%s", n, tag))
            }
        }
    }

    ui.Step("Downloading Windows SDK installers (.msi)")
    sp, ok := best["sdk"]
    if !ok {
        return fmt.Errorf("Windows SDK package not found")
    }
    found := map[string]bool{}
    for _, pl := range sp.Payloads {
        low := strings.ToLower(pl.FileName)
        base := low[strings.LastIndex(low, "/")+1:]
        if !wantMSI[base] {
            continue
        }
        found[base] = true
        dst, ok := safeJoin(filepath.Join(cache, "sdk"), pl.FileName)
        if !ok {
            continue
        }
        ui.Info(fmt.Sprintf("%s (%.1f MB)", filepath.Base(dst), float64(pl.Size)/1e6))
        if err := dl(pl.URL, dst, pl.SHA256, pl.Size); err != nil {
            return err
        }
    }
    for name := range wantMSI {
        if !found[name] {
            ui.Warn("not found in manifest: " + name)
        }
    }
    ui.OK("Download done. Next step: read the .msi files to find the .cab files they need")
    return nil
}
