package deps

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "sort"
    "strings"

    "github.com/muriloprimata/uwp4me/internal/msix"
    "github.com/muriloprimata/uwp4me/internal/ui"
)

// Tool is a command UWP4Me needs and the Termux package that provides it.
type Tool struct {
    Bin string
    Pkg string
}

var Required = []Tool{
    {"git", "git"},
    {"cmake", "cmake"},
    {"make", "make"},
    {"clang", "clang"},
    {"clang++", "clang"},
    {"lld-link", "lld"},
    {"msiextract", "msitools"},
    {"msiinfo", "msitools"},
    {"osslsigncode", "osslsigncode"},
    {"openssl", "openssl-tool"},
}

func Home() string {
    h, _ := os.UserHomeDir()
    return filepath.Join(h, ".uwp4me")
}

func marker() string { return filepath.Join(Home(), ".setup-done") }

func FirstRun() bool {
    _, err := os.Stat(marker())
    return err != nil
}

func isTermux() bool {
    return strings.Contains(os.Getenv("PREFIX"), "com.termux")
}

func Setup() error {
    ui.Step("Checking dependencies")
    pkgs := map[string]bool{}
    for _, t := range Required {
        if _, err := exec.LookPath(t.Bin); err != nil {
            ui.Warn(t.Bin + " not found")
            pkgs[t.Pkg] = true
        } else {
            ui.OK(t.Bin + " found")
        }
    }

    if len(pkgs) > 0 {
        var list []string
        for p := range pkgs {
            list = append(list, p)
        }
        sort.Strings(list)
        if !isTermux() {
            return fmt.Errorf("please install: %s", strings.Join(list, ", "))
        }
        ui.Info("Installing with pkg: " + strings.Join(list, " "))
        cmd := exec.Command("pkg", append([]string{"install", "-y"}, list...)...)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
            return fmt.Errorf("pkg install failed: %w", err)
        }
        ui.OK("Dependencies installed")
    }

    ui.Step("Preparing workspace")
    if err := os.MkdirAll(filepath.Join(Home(), "bin"), 0o755); err != nil {
        return err
    }
    ui.OK("Workspace ready at " + Home())

    if err := msix.Ensure(Home()); err != nil {
        return err
    }
    return os.WriteFile(marker(), []byte("ok\n"), 0o644)
}
