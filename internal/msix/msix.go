package msix

import (
    _ "embed"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "github.com/muriloprimata/uwp4me/internal/ui"
)

//go:embed msix-termux.patch
var termuxPatch []byte

//go:embed commit.txt
var pinned string

const repoURL = "https://github.com/microsoft/msix-packaging"

func BinPath(home string) string { return filepath.Join(home, "bin", "makemsix") }
func LibPath(home string) string { return filepath.Join(home, "lib", "libmsix.so") }

func Installed(home string) bool {
    _, e1 := os.Stat(BinPath(home))
    _, e2 := os.Stat(LibPath(home))
    return e1 == nil && e2 == nil
}

func run(dir string, log *os.File, name string, args ...string) error {
    cmd := exec.Command(name, args...)
    cmd.Dir = dir
    cmd.Stdout = log
    cmd.Stderr = log
    return cmd.Run()
}

func copyFile(src, dst string, mode os.FileMode) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()
    out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
    if err != nil {
        return err
    }
    defer out.Close()
    _, err = io.Copy(out, in)
    return err
}

func short(s string) string {
    if len(s) > 7 {
        return s[:7]
    }
    return s
}

// Ensure builds and installs makemsix (pinned SDK version + Termux patches) if missing.
func Ensure(home string) error {
    if Installed(home) && os.Getenv("UWP4ME_REBUILD") == "" {
        ui.OK("makemsix already installed")
        return nil
    }

    ui.Step("Building makemsix (first time only, takes 10-15 minutes)")
    if err := os.MkdirAll(home, 0o755); err != nil {
        return err
    }
    logPath := filepath.Join(home, "msix-build.log")
    log, err := os.Create(logPath)
    if err != nil {
        return err
    }
    defer log.Close()
    ui.Info("Full log: " + logPath)

    fail := func(what string, err error) error {
        return fmt.Errorf("%s failed: %v (see %s)", what, err, logPath)
    }

    rev := strings.TrimSpace(pinned)
    src := filepath.Join(home, "src", "msix-packaging")
    if _, err := os.Stat(src); err != nil {
        if err := os.MkdirAll(src, 0o755); err != nil {
            return err
        }
        if rev == "" {
            os.RemoveAll(src)
            ui.Info("Cloning MSIX Packaging SDK (latest)...")
            if err := run("", log, "git", "clone", "--depth", "1", "--recurse-submodules", "--shallow-submodules", repoURL, src); err != nil {
                return fail("clone", err)
            }
        } else {
            ui.Info("Fetching MSIX Packaging SDK @ " + short(rev) + "...")
            steps := [][]string{
                {"init", "-q"},
                {"remote", "add", "origin", repoURL},
                {"fetch", "--depth", "1", "origin", rev},
                {"checkout", "-q", "FETCH_HEAD"},
            }
            for _, st := range steps {
                if err := run(src, log, "git", st...); err != nil {
                    os.RemoveAll(src)
                    return fail("fetch", err)
                }
            }
        }
    }

    ui.Info("Applying Termux patches...")
    patchFile := filepath.Join(home, "msix-termux.patch")
    if err := os.WriteFile(patchFile, termuxPatch, 0o644); err != nil {
        return err
    }
    // If the reverse patch applies cleanly, it is already applied.
    if run(src, log, "git", "apply", "--check", "--reverse", patchFile) != nil {
        if err := run(src, log, "git", "apply", "--whitespace=nowarn", patchFile); err != nil {
            return fail("patch", err)
        }
    }

    build := filepath.Join(src, "build")
    if err := os.MkdirAll(build, 0o755); err != nil {
        return err
    }

    ui.Info("Configuring (cmake)...")
    if err := run(build, log, "cmake",
        "-DCMAKE_BUILD_TYPE=MinSizeRel", "-DSKIP_BUNDLES=off",
        "-DUSE_VALIDATION_PARSER=on", "-DCMAKE_TOOLCHAIN_FILE=../cmake/linux.cmake",
        "-DMSIX_PACK=on", "-DMSIX_SAMPLES=off", "-DMSIX_TESTS=off", "-DLINUX=on", ".."); err != nil {
        return fail("cmake", err)
    }

    ui.Info("Compiling (the slow part, keep the screen on)...")
    if err := run(build, log, "make", "-j2"); err != nil {
        return fail("make", err)
    }

    if err := os.MkdirAll(filepath.Join(home, "lib"), 0o755); err != nil {
        return err
    }
    if err := copyFile(filepath.Join(build, "bin", "makemsix"), BinPath(home), 0o755); err != nil {
        return err
    }
    if err := copyFile(filepath.Join(build, "lib", "libmsix.so"), LibPath(home), 0o644); err != nil {
        return err
    }
    ui.OK("makemsix installed")
    return nil
}
