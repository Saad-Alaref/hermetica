package cmd

import (
    "fmt"
    "hermetica/internal/config"
    "github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
    Use:   "resume",
    Short: "Resume from existing artifacts",
    RunE: func(cmd *cobra.Command, args []string) error {
        _, err := config.Load(cfgPath)
        if err != nil {
            return err
        }
        // Placeholder: resume logic will use artifacts in workdir
        fmt.Println("Resume not yet implemented — will use artifacts when available.")
        return nil
    },
}

