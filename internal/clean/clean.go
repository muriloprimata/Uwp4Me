package clean

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/muriloprimata/uwp4me/internal/ui"
)

func size(p string) int64 {
    var n int64
    filepath.WalkDir(p, func(_ string, d os.DirEntry, err error) error {
        if err != nil {
            return nil
        }
        if !d.IsDir() && d.Type()&os.ModeSymlink == 0 {
            if i, e := d.Info(); e == nil {
                n += i.Size()
            }
        }
        return nil
    })
    return n
}

// Run removes what UWP4Me built or downloaded.
// Default: compiled tools. --sdk: also the SDK and download cache.
// --all: everything, license included. --yes: do not ask.
func Run(home string, args []string) error {
    all, withSDK, yes := false, false, false
    for _, a := range args {
        switch a {
        case "--all":
            all = true
        case "--sdk":
            withSDK = true
        case "--yes", "-y":
            yes = true
        default:
            return fmt.Errorf("unknown option: %s", a)
        }
    }
    if filepath.Base(home) != ".uwp4me" {
        return fmt.Errorf("refusing to clean unexpected path: %s", home)
    }

    var names []string
    if all {
        names = []string{""}
    } else {
        names = []string{"src", "bin", "lib", "msix-build.log", "msix-termux.patch", ".setup-done", "doctor-tmp"}
        if withSDK {
            names = append(names, "sdk", "cache")
        }
    }

    ui.Step("Cleaning " + home)
    var targets []string
    var total int64
    for _, n := range names {
        t := filepath.Join(home, n)
        if _, err := os.Lstat(t); err != nil {
            continue
        }
        sz := size(t)
        total += sz
        targets = append(targets, t)
        ui.Info(fmt.Sprintf("%s (%.1f MB)", t, float64(sz)/1e6))
    }
    if len(targets) == 0 {
        ui.OK("Nothing to clean")
        return nil
    }

    if !yes {
        fmt.Print("\n  Remove these? [y/N] ")
        ans, _ := bufio.NewReader(os.Stdin).ReadString('\n')
        ans = strings.ToLower(strings.TrimSpace(ans))
        if ans != "y" && ans != "yes" {
            ui.Warn("Cancelled, nothing removed")
            return nil
        }
    }
    for _, t := range targets {
        if err := os.RemoveAll(t); err != nil {
            return err
        }
    }
    ui.OK(fmt.Sprintf("Freed about %.1f MB", float64(total)/1e6))
    return nil
}
