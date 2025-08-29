package cmd

import (
    "fmt"
    "os"

    "hermetica/internal/logging"
    "github.com/spf13/cobra"
)

var (
    cfgPath string
    domainOverride string
    workdir string
    force bool
    debug bool
    profile string
)

var rootCmd = &cobra.Command{
    Use:   "hermetica",
    Short: "Hermetica - web attack surface mapper",
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        logging.Init(debug)
    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func init() {
    rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "configs/hermetica.yaml", "Config file path")
    rootCmd.PersistentFlags().StringVarP(&domainOverride, "domain", "d", "", "Domain override")
    rootCmd.PersistentFlags().StringVarP(&workdir, "workdir", "w", "./work", "Working directory")
    rootCmd.PersistentFlags().BoolVar(&force, "force", false, "Re-run stages even if artifacts exist")
    rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Verbose logs")
    rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "Scan profile override: stealth|thorough")

    rootCmd.AddCommand(runCmd)
    rootCmd.AddCommand(resumeCmd)
    rootCmd.AddCommand(exportCmd)
    rootCmd.AddCommand(doctorCmd)
}

