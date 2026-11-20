package ui

import (
    "fmt"
    "os"
)

const (
    reset  = "\033[0m"
    bold   = "\033[1m"
    dim    = "\033[2m"
    red    = "\033[31m"
    green  = "\033[32m"
    yellow = "\033[33m"
    cyan   = "\033[36m"
)

func c(code, s string) string {
    if os.Getenv("NO_COLOR") != "" {
        return s
    }
    return code + s + reset
}

func Banner(version string) {
    fmt.Println()
    fmt.Println(c(bold+cyan, "  ╦ ╦╦ ╦╔═╗╦  ╦╔╦╗╔═╗"))
    fmt.Println(c(bold+cyan, "  ║ ║║║║╠═╝╚╗╔╝║║║║╣ "))
    fmt.Println(c(bold+cyan, "  ╚═╝╚╩╝╩   ╚╝ ╩ ╩╚═╝"))
    fmt.Println(c(dim, "  Build and package Xbox UWP apps from Android · v"+version))
    fmt.Println()
}

func Step(s string) { fmt.Println(c(bold, "\n▸ "+s)) }
func Info(s string) { fmt.Println(c(cyan, "  • ") + s) }
func OK(s string)   { fmt.Println(c(green, "  ✔ ") + s) }
func Warn(s string) { fmt.Println(c(yellow, "  ! ") + s) }
func Fail(s string) { fmt.Fprintln(os.Stderr, c(red, "  ✘ ")+s) }
