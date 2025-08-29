package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
    Use:   "export",
    Short: "Export data from SQLite to CSV/JSON",
    RunE: func(cmd *cobra.Command, args []string) error {
        fmt.Println("Export not yet implemented — will query SQLite and write CSV/JSON.")
        return nil
    },
}

