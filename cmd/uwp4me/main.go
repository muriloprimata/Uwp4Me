package main

import (
    "fmt"
    "os"

    "github.com/muriloprimata/uwp4me/internal/clean"
    "github.com/muriloprimata/uwp4me/internal/deps"
    "github.com/muriloprimata/uwp4me/internal/doctor"
    "github.com/muriloprimata/uwp4me/internal/license"
    "github.com/muriloprimata/uwp4me/internal/sdk"
    "github.com/muriloprimata/uwp4me/internal/ui"
)

var version = "0.1.0"

func usage() {
    ui.Banner(version)
    fmt.Println("Usage: uwp4me <command> [options]")
    fmt.Println()
    fmt.Println("Commands:")
    fmt.Println("  setup         install tools and build makemsix")
    fmt.Println("  sdk           download the Windows SDK (asks for license)")
    fmt.Println("  doctor        check that everything works")
    fmt.Println("  clean         remove built tools (--sdk also SDK, --all everything)")
    fmt.Println("  init <name>   create a new project (coming soon)")
    fmt.Println("  build         compile the app (coming soon)")
    fmt.Println("  pack          package and sign the .msix (coming soon)")
    fmt.Println("  deploy        install on the Xbox (coming soon)")
    fmt.Println("  version       print version")
}

func runSetup() {
    ui.Banner(version)
    ui.Info("Welcome! Setting things up for the first time...")
    if err := deps.Setup(); err != nil {
        ui.Fail(err.Error())
        os.Exit(1)
    }
    ui.OK("All set! Run 'uwp4me' to see the commands.")
    fmt.Println()
}

func hasFlag(name string) bool {
    for _, a := range os.Args[2:] {
        if a == name {
            return true
        }
    }
    return false
}

func rest() []string {
    if len(os.Args) > 2 {
        return os.Args[2:]
    }
    return nil
}

func main() {
    cmd := ""
    if len(os.Args) > 1 {
        cmd = os.Args[1]
    }

    auto := cmd != "doctor" && cmd != "clean" && cmd != "version"
    if cmd == "setup" || (auto && deps.FirstRun()) {
        runSetup()
        if cmd == "" || cmd == "setup" {
            return
        }
    }

    switch cmd {
    case "sdk":
        if err := license.Ensure(deps.Home(), hasFlag("--accept-license")); err != nil {
            ui.Fail(err.Error())
            os.Exit(1)
        }
        if err := sdk.Run(deps.Home(), rest()); err != nil {
            ui.Fail(err.Error())
            os.Exit(1)
        }
    case "doctor":
        os.Exit(doctor.Run(deps.Home()))
    case "clean":
        if err := clean.Run(deps.Home(), rest()); err != nil {
            ui.Fail(err.Error())
            os.Exit(1)
        }
    case "init", "build", "pack", "deploy", "run":
        ui.Warn("'" + cmd + "' is not implemented yet")
    case "version":
        fmt.Println(version)
    default:
        usage()
        os.Exit(1)
    }
}
