package license

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/muriloprimata/uwp4me/internal/ui"
)

// TODO: confirm the exact terms link before publishing.
const termsURL = "https://visualstudio.microsoft.com/license-terms/"

func file(home string) string { return filepath.Join(home, "license-accepted") }

func Accepted(home string) bool {
    _, err := os.Stat(file(home))
    return err == nil
}

func record(home string) error {
    if err := os.MkdirAll(home, 0o755); err != nil {
        return err
    }
    return os.WriteFile(file(home), []byte("accepted\n"), 0o644)
}

// Ensure asks the user to accept Microsoft's terms once.
func Ensure(home string, assumeYes bool) error {
    if Accepted(home) {
        return nil
    }
    if assumeYes || os.Getenv("UWP4ME_ACCEPT_LICENSE") == "1" {
        ui.Info("License accepted via flag/environment")
        return record(home)
    }
    st, err := os.Stdin.Stat()
    if err != nil || st.Mode()&os.ModeCharDevice == 0 {
        return fmt.Errorf("license not accepted: run interactively or pass --accept-license")
    }

    ui.Step("Microsoft license terms")
    fmt.Println("  UWP4Me downloads the Windows SDK and the Visual C++ runtime")
    fmt.Println("  headers and libraries directly from Microsoft's servers.")
    fmt.Println("  These files are NOT part of UWP4Me and are covered by")
    fmt.Println("  Microsoft's license terms:")
    fmt.Println("    " + termsURL)
    fmt.Println()
    fmt.Print("  Do you accept the terms? [y/N] ")

    ans, _ := bufio.NewReader(os.Stdin).ReadString('\n')
    ans = strings.ToLower(strings.TrimSpace(ans))
    if ans != "y" && ans != "yes" {
        return fmt.Errorf("license not accepted")
    }
    return record(home)
}
