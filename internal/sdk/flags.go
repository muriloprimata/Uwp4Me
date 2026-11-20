package sdk

import (
    "os/exec"
    "path/filepath"
    "strings"
)

// Root is the folder where the ready-to-use SDK layout lives.
func Root(home string) string { return filepath.Join(home, "sdk") }

// CompileFlags returns the clang flags to build a UWP x64 app against the local SDK.
func CompileFlags(home string) ([]string, error) {
    out, err := exec.Command("clang", "-print-resource-dir").Output()
    if err != nil {
        return nil, err
    }
    res := strings.TrimSpace(string(out))
    s := Root(home)
    j := func(p ...string) string { return filepath.Join(append([]string{s}, p...)...) }
    return []string{
        "--target=x86_64-pc-windows-msvc", "-std=c++20",
        "-fms-extensions", "-fms-compatibility", "-fms-compatibility-version=19.44",
        "-fdelayed-template-parsing", "-fexceptions", "-fcxx-exceptions",
        "-D_MT", "-D_DLL", "-DUNICODE", "-D_UNICODE", "-DWINAPI_FAMILY=2", "-D_WIN32_WINNT=0x0A00",
        "-nostdinc", "-Wno-everything",
        "-isystem", filepath.Join(res, "include"),
        "-isystem", j("crt", "include"),
        "-isystem", j("sdk", "include", "ucrt"),
        "-isystem", j("sdk", "include", "um"),
        "-isystem", j("sdk", "include", "shared"),
        "-isystem", j("sdk", "include", "winrt"),
        "-isystem", j("sdk", "include", "cppwinrt"),
    }, nil
}
