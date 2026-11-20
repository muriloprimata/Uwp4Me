package sdk

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "time"

    "github.com/muriloprimata/uwp4me/internal/ui"
)

const channelURL = "https://aka.ms/vs/17/release/channel"

type payload struct {
    FileName string `json:"fileName"`
    SHA256   string `json:"sha256"`
    Size     int64  `json:"size"`
    URL      string `json:"url"`
}

type pkg struct {
    ID       string    `json:"id"`
    Version  string    `json:"version"`
    Payloads []payload `json:"payloads"`
}

type vsman struct {
    Packages []pkg `json:"packages"`
}

type channelItem struct {
    ID       string    `json:"id"`
    Payloads []payload `json:"payloads"`
}

type channel struct {
    ChannelItems []channelItem `json:"channelItems"`
}

var vcRe = regexp.MustCompile(`(?i)^Microsoft\.VC\.[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+\.(.+)$`)

func get(url, dst string) error {
    client := &http.Client{Timeout: 10 * time.Minute}
    req, err := http.NewRequest("GET", url, nil)
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
        return fmt.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
    }
    f, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer f.Close()
    _, err = io.Copy(f, resp.Body)
    return err
}

func readJSON(path string, v interface{}) error {
    raw, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    raw = bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf"))
    return json.Unmarshal(raw, v)
}

func verLess(a, b string) bool {
    as, bs := strings.Split(a, "."), strings.Split(b, ".")
    for i := 0; i < len(as) && i < len(bs); i++ {
        x, _ := strconv.Atoi(as[i])
        y, _ := strconv.Atoi(bs[i])
        if x != y {
            return x < y
        }
    }
    return len(as) < len(bs)
}

func kind(id string) string {
    if m := vcRe.FindStringSubmatch(id); m != nil {
        k := strings.ToLower(m[1])
        if k == "crt.headers.base" || k == "crt.x64.desktop.base" || k == "crt.x64.store.base" {
            return "vc:" + k
        }
        return ""
    }
    l := strings.ToLower(id)
    if strings.HasPrefix(l, "win10sdk_") || strings.HasPrefix(l, "win11sdk_") {
        return "sdk"
    }
    return ""
}

// Plan downloads the VS package manifest (cached) and lists what we need.
func Plan(home string) error {
    cache := filepath.Join(home, "cache")
    if err := os.MkdirAll(cache, 0o755); err != nil {
        return err
    }
    vsPath := filepath.Join(cache, "vs.manifest.json")

    if _, err := os.Stat(vsPath); err != nil || os.Getenv("UWP4ME_REFRESH") != "" {
        ui.Step("Fetching Visual Studio manifest")
        chPath := filepath.Join(cache, "channel.json")
        ui.Info("Channel: " + channelURL)
        if err := get(channelURL, chPath); err != nil {
            return err
        }
        var ch channel
        if err := readJSON(chPath, &ch); err != nil {
            return fmt.Errorf("channel json: %w", err)
        }
        url := ""
        for _, it := range ch.ChannelItems {
            if it.ID == "Microsoft.VisualStudio.Manifests.VisualStudio" && len(it.Payloads) > 0 {
                url = it.Payloads[0].URL
            }
        }
        if url == "" {
            return fmt.Errorf("package manifest link not found in channel manifest")
        }
        ui.Info("Downloading package manifest (can take a minute)...")
        if err := get(url, vsPath); err != nil {
            return err
        }
        ui.OK("Manifest cached at " + vsPath)
    } else {
        ui.OK("Using cached manifest")
    }

    var m vsman
    if err := readJSON(vsPath, &m); err != nil {
        return fmt.Errorf("package manifest json: %w", err)
    }
    ui.Info(fmt.Sprintf("%d packages in manifest", len(m.Packages)))

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
    if len(best) == 0 {
        return fmt.Errorf("no matching packages found (names may have changed)")
    }

    keys := make([]string, 0, len(best))
    for k := range best {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    ui.Step("Newest packages we need")
    for _, k := range keys {
        p := best[k]
        var total int64
        for _, pl := range p.Payloads {
            total += pl.Size
        }
        ui.Info(fmt.Sprintf("%s  v%s  %d files  %.1f MB", p.ID, p.Version, len(p.Payloads), float64(total)/1e6))
        if k == "sdk" {
            cabs := 0
            for _, pl := range p.Payloads {
                n := strings.ToLower(pl.FileName)
                if strings.HasSuffix(n, ".msi") {
                    fmt.Println("      " + pl.FileName)
                } else if strings.HasSuffix(n, ".cab") {
                    cabs++
                }
            }
            fmt.Printf("      (+ %d .cab files)\n", cabs)
        }
    }
    return nil
}
