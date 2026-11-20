package sdk

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strings"

    "github.com/muriloprimata/uwp4me/internal/ui"
)

var includeRe = regexp.MustCompile(`^\s*#\s*include\s*[<"]([^>"]+)[>"]`)

type entry struct {
    dir  string
    name string
}

func firstDir(p string) (string, error) {
    ents, err := os.ReadDir(p)
    if err != nil {
        return "", err
    }
    for _, e := range ents {
        if e.IsDir() {
            return e.Name(), nil
        }
    }
    return "", fmt.Errorf("no subfolder in %s", p)
}

// lowerLinks creates a lowercase symlink next to every file or folder with capitals.
func lowerLinks(root string) int {
    var paths []string
    filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
        if err != nil || p == root {
            return nil
        }
        paths = append(paths, p)
        return nil
    })
    n := 0
    for _, p := range paths {
        base := filepath.Base(p)
        low := strings.ToLower(base)
        if base == low {
            continue
        }
        dst := filepath.Join(filepath.Dir(p), low)
        if _, err := os.Lstat(dst); err == nil {
            continue
        }
        if os.Symlink(base, dst) == nil {
            n++
        }
    }
    return n
}

// includeLinks reads every #include in the headers and creates links with the exact spelling used.
func includeLinks(roots []string) int {
    index := map[string][]entry{}
    var headers []string
    for _, root := range roots {
        filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
            if err != nil || d.IsDir() || d.Type()&os.ModeSymlink != 0 {
                return nil
            }
            name := d.Name()
            index[strings.ToLower(name)] = append(index[strings.ToLower(name)], entry{filepath.Dir(p), name})
            switch strings.ToLower(filepath.Ext(name)) {
            case ".h", ".hpp", ".inl", "":
                headers = append(headers, p)
            }
            return nil
        })
    }

    seen := map[string]bool{}
    for _, h := range headers {
        f, err := os.Open(h)
        if err != nil {
            continue
        }
        sc := bufio.NewScanner(f)
        sc.Buffer(make([]byte, 64*1024), 1024*1024)
        for sc.Scan() {
            line := sc.Text()
            if !strings.Contains(line, "include") {
                continue
            }
            m := includeRe.FindStringSubmatch(line)
            if m == nil {
                continue
            }
            name := m[1]
            if i := strings.LastIndexAny(name, "/\\"); i >= 0 {
                name = name[i+1:]
            }
            if name != "" && !strings.Contains(name, "..") {
                seen[name] = true
            }
        }
        f.Close()
    }

    n := 0
    for name := range seen {
        for _, e := range index[strings.ToLower(name)] {
            if e.name == name {
                continue
            }
            dst := filepath.Join(e.dir, name)
            if _, err := os.Lstat(dst); err == nil {
                continue
            }
            if os.Symlink(e.name, dst) == nil {
                n++
            }
        }
    }
    return n
}

// Layout builds ~/.uwp4me/sdk from the extracted SDK and CRT, fixing file name case.
func Layout(home string) error {
    cache := filepath.Join(home, "cache")
    k := filepath.Join(cache, "sdkx", "Program Files", "Windows Kits", "10")
    v, err := firstDir(filepath.Join(k, "Include"))
    if err != nil {
        return fmt.Errorf("SDK not extracted yet (run: uwp4me sdk): %w", err)
    }
    crt := filepath.Join(cache, "crt")
    msvc := filepath.Join(crt, "crt.headers.base", "VC", "Tools", "MSVC")
    m, err := firstDir(msvc)
    if err != nil {
        return fmt.Errorf("CRT not downloaded yet (run: uwp4me sdk): %w", err)
    }
    mi := filepath.Join(msvc, m, "include")
    ml := filepath.Join(crt, "crt.x64.desktop.base", "VC", "Tools", "MSVC", m, "lib", "x64")
    ms := filepath.Join(crt, "crt.x64.store.base", "VC", "Tools", "MSVC", m, "lib", "x64")
    inc := filepath.Join(k, "Include", v)
    libUm := filepath.Join(k, "Lib", v, "um", "x64")
    libUcrt := filepath.Join(k, "Lib", v, "ucrt", "x64")

    ui.Step("Fixing file name case (Windows ignores it, Android does not)")
    n := 0
    for _, d := range []string{inc, mi, libUm, libUcrt, ml, ms} {
        n += lowerLinks(d)
    }
    ui.Info(fmt.Sprintf("%d lowercase links", n))
    ui.Info("Scanning #include lines (can take a minute)...")
    ui.Info(fmt.Sprintf("%d include-name links", includeLinks([]string{inc, mi})))

    s := Root(home)
    os.RemoveAll(s)
    for _, d := range []string{"crt/lib/x64", "sdk/include", "sdk/lib/um", "sdk/lib/ucrt"} {
        if err := os.MkdirAll(filepath.Join(s, d), 0o755); err != nil {
            return err
        }
    }
    os.Symlink(mi, filepath.Join(s, "crt", "include"))
    for _, dir := range []string{ml, ms} {
        ents, _ := os.ReadDir(dir)
        for _, e := range ents {
            dst := filepath.Join(s, "crt", "lib", "x64", e.Name())
            os.Remove(dst)
            os.Symlink(filepath.Join(dir, e.Name()), dst)
        }
    }
    for _, d := range []string{"cppwinrt", "shared", "ucrt", "um", "winrt"} {
        os.Symlink(filepath.Join(inc, d), filepath.Join(s, "sdk", "include", d))
    }
    os.Symlink(libUm, filepath.Join(s, "sdk", "lib", "um", "x64"))
    os.Symlink(libUcrt, filepath.Join(s, "sdk", "lib", "ucrt", "x64"))

    ui.OK(fmt.Sprintf("SDK %s / MSVC %s ready at %s", v, m, s))
    return nil
}
