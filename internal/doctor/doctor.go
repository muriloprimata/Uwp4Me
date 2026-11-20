package doctor

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "sort"
    "strings"

    "github.com/muriloprimata/uwp4me/internal/deps"
    "github.com/muriloprimata/uwp4me/internal/license"
    "github.com/muriloprimata/uwp4me/internal/msix"
    "github.com/muriloprimata/uwp4me/internal/sdk"
    "github.com/muriloprimata/uwp4me/internal/ui"
)

type counter struct{ ok, warn, fail int }

func (c *counter) pass(s string) { ui.OK(s); c.ok++ }
func (c *counter) note(s string) { ui.Warn(s); c.warn++ }
func (c *counter) bad(s string)  { ui.Fail(s); c.fail++ }

func exists(p string) bool {
    _, err := os.Stat(p)
    return err == nil
}

func firstLine(s string) string {
    s = strings.TrimSpace(s)
    if i := strings.Index(s, "\n"); i >= 0 {
        s = s[:i]
    }
    return s
}

func firstError(s string) string {
    for _, l := range strings.Split(s, "\n") {
        if strings.Contains(l, "error") {
            return strings.TrimSpace(l)
        }
    }
    return firstLine(s)
}

// compileTest builds a tiny Windows x64 program against the local SDK.
func compileTest(home string) error {
    dir := filepath.Join(home, "doctor-tmp")
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return err
    }
    defer os.RemoveAll(dir)

    src := filepath.Join(dir, "t.cpp")
    obj := filepath.Join(dir, "t.obj")
    exe := filepath.Join(dir, "t.exe")
    code := "#include <windows.h>\nint main() { return 0; }\n"
    if err := os.WriteFile(src, []byte(code), 0o644); err != nil {
        return err
    }
    flags, err := sdk.CompileFlags(home)
    if err != nil {
        return err
    }
    out, err := exec.Command("clang++", append(flags, "-c", src, "-o", obj)...).CombinedOutput()
    if err != nil || !exists(obj) {
        return fmt.Errorf("compile: %s", firstError(string(out)))
    }
    lib := filepath.Join(sdk.Root(home), "sdk", "lib", "um", "x64")
    out, err = exec.Command("lld-link", obj, "/out:"+exe, "/subsystem:console", "/entry:main",
        "/nodefaultlib", "/machine:x64", "/libpath:"+lib, "kernel32.lib").CombinedOutput()
    if err != nil || !exists(exe) {
        return fmt.Errorf("link: %s", firstError(string(out)))
    }
    return nil
}

// Run checks the whole toolchain and returns the process exit code.
func Run(home string) int {
    c := &counter{}

    ui.Step("System")
    ui.Info(fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH))
    if strings.Contains(os.Getenv("PREFIX"), "com.termux") {
        c.pass("Termux detected")
    } else {
        c.note("Not running inside Termux (untested)")
    }
    if out, err := exec.Command("df", "-h", home).Output(); err == nil {
        lines := strings.Split(strings.TrimSpace(string(out)), "\n")
        ui.Info("Disk: " + strings.Join(strings.Fields(lines[len(lines)-1]), " "))
    }

    ui.Step("Tools")
    missing := map[string]bool{}
    for _, t := range deps.Required {
        if _, err := exec.LookPath(t.Bin); err != nil {
            c.bad(t.Bin + " not found")
            missing[t.Pkg] = true
        } else {
            c.pass(t.Bin)
        }
    }
    if len(missing) > 0 {
        var list []string
        for p := range missing {
            list = append(list, p)
        }
        sort.Strings(list)
        ui.Info("fix: pkg install " + strings.Join(list, " "))
    }
    if out, err := exec.Command("clang", "--print-targets").CombinedOutput(); err == nil {
        if strings.Contains(strings.ToLower(string(out)), "x86") {
            c.pass("clang can target x86-64 (Windows)")
        } else {
            c.bad("clang has no x86 target")
        }
    }

    ui.Step("Packaging")
    bin := msix.BinPath(home)
    if !exists(bin) {
        c.bad("makemsix not installed (fix: uwp4me setup)")
    } else {
        cmd := exec.Command(bin)
        cmd.Env = append(os.Environ(), "LD_LIBRARY_PATH="+filepath.Dir(msix.LibPath(home)))
        out, _ := cmd.CombinedOutput()
        if strings.Contains(strings.ToLower(string(out)), "makemsix") {
            c.pass("makemsix runs")
        } else {
            c.bad("makemsix does not run: " + firstLine(string(out)))
        }
    }

    ui.Step("License")
    if license.Accepted(home) {
        c.pass("Microsoft terms accepted")
    } else {
        c.note("Microsoft terms not accepted yet (fix: uwp4me sdk)")
    }

    ui.Step("Windows SDK")
    root := sdk.Root(home)
    files := []string{
        "crt/include/vector",
        "sdk/include/um/windows.h",
        "sdk/include/cppwinrt/winrt/base.h",
        "sdk/lib/um/x64/WindowsApp.lib",
        "sdk/lib/um/x64/d3d11.lib",
        "crt/lib/x64/store/vcruntime.lib",
    }
    all := true
    for _, f := range files {
        if exists(filepath.Join(root, f)) {
            c.pass(f)
        } else {
            c.bad("missing " + f)
            all = false
        }
    }
    if !all {
        ui.Info("fix: uwp4me sdk")
    } else if err := compileTest(home); err != nil {
        c.bad("test build failed: " + err.Error())
    } else {
        c.pass("test build (windows.h to .exe) works")
    }

    ui.Step("Summary")
    fmt.Printf("  %d ok, %d warnings, %d problems\n\n", c.ok, c.warn, c.fail)
    if c.fail > 0 {
        return 1
    }
    return 0
}
