package sdk

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "strings"

    "github.com/muriloprimata/uwp4me/internal/ui"
)

var cabRe = regexp.MustCompile(`(?i)^(.+\.cab)$`)

func loadBest(home string) (map[string]pkg, error) {
    var m vsman
    if err := readJSON(filepath.Join(home, "cache", "vs.manifest.json"), &m); err != nil {
        return nil, err
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
    return best, nil
}

// cabsOf lists the external .cab files an .msi refers to.
func cabsOf(msi string) ([]string, error) {
    out, err := exec.Command("msiinfo", "export", msi, "Media").Output()
    if err != nil {
        return nil, fmt.Errorf("msiinfo: %w", err)
    }
    var cabs []string
    seen := map[string]bool{}
    for _, line := range strings.Split(string(out), "\n") {
        for _, f := range strings.Split(strings.TrimRight(line, "\r"), "\t") {
            f = strings.TrimSpace(f)
            if strings.HasPrefix(f, "#") {
                continue // embedded in the msi itself
            }
            if m := cabRe.FindStringSubmatch(f); m != nil {
                k := strings.ToLower(m[1])
                if !seen[k] {
                    seen[k] = true
                    cabs = append(cabs, m[1])
                }
            }
        }
    }
    return cabs, nil
}

// Extract downloads the .cab files and unpacks the SDK installers.
func Extract(home string) error {
    for _, t := range []string{"msiextract", "msiinfo"} {
        if _, err := exec.LookPath(t); err != nil {
            return fmt.Errorf("%s not found (run: pkg install msitools)", t)
        }
    }
    best, err := loadBest(home)
    if err != nil {
        return err
    }
    sp, ok := best["sdk"]
    if !ok {
        return fmt.Errorf("Windows SDK package not found")
    }
    idx := map[string]payload{}
    for _, pl := range sp.Payloads {
        low := strings.ToLower(pl.FileName)
        idx[low[strings.LastIndex(low, "/")+1:]] = pl
    }

    cache := filepath.Join(home, "cache")
    outDir := filepath.Join(cache, "sdkx")
    if err := os.MkdirAll(outDir, 0o755); err != nil {
        return err
    }

    var msis []string
    filepath.Walk(filepath.Join(cache, "sdk"), func(p string, info os.FileInfo, err error) error {
        if err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(p), ".msi") {
            msis = append(msis, p)
        }
        return nil
    })

    ui.Step("Downloading .cab files and extracting")
    for _, msi := range msis {
        cabs, err := cabsOf(msi)
        if err != nil {
            ui.Warn(filepath.Base(msi) + ": " + err.Error())
            continue
        }
        var total int64
        for _, c := range cabs {
            pl, ok := idx[strings.ToLower(c)]
            if !ok {
                ui.Warn("cab not in manifest: " + c)
                continue
            }
            total += pl.Size
            if err := dl(pl.URL, filepath.Join(filepath.Dir(msi), c), pl.SHA256, pl.Size); err != nil {
                return err
            }
        }
        ui.Info(fmt.Sprintf("%s (%d cab, %.1f MB)", filepath.Base(msi), len(cabs), float64(total)/1e6))
        cmd := exec.Command("msiextract", "-C", outDir, msi)
        if out, err := cmd.CombinedOutput(); err != nil {
            ui.Warn("msiextract failed: " + strings.TrimSpace(string(out)))
        }
    }

    var files int
    var size int64
    filepath.Walk(outDir, func(p string, info os.FileInfo, err error) error {
        if err == nil && !info.IsDir() {
            files++
            size += info.Size()
        }
        return nil
    })
    ui.OK(fmt.Sprintf("%d files extracted (%.1f MB) to %s", files, float64(size)/1e6, outDir))
    return nil
}
