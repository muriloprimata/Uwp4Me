package sdk

// Run is the entry point of the "sdk" command.
func Run(home string, args []string) error {
    for _, a := range args {
        switch a {
        case "plan":
            return Plan(home)
        case "layout":
            return Layout(home)
        }
    }
    if err := Fetch(home); err != nil {
        return err
    }
    if err := Extract(home); err != nil {
        return err
    }
    return Layout(home)
}
