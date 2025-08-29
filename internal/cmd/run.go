package cmd

import (
    "context"
    "fmt"
    "time"

    "hermetica/internal/config"
    "hermetica/internal/logging"
    "hermetica/internal/pipeline"
    "github.com/rs/zerolog/log"
    "github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
    Use:   "run",
    Short: "Execute the full pipeline",
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg, err := config.Load(cfgPath)
        if err != nil {
            return err
        }
        if domainOverride != "" && len(cfg.Targets) > 0 {
            cfg.Targets[0].Domain = domainOverride
        }
        if workdir != "" {
            cfg.Workdir = workdir
        }
        if profile != "" {
            cfg.Scan.Profile = profile
        }
        logging.Init(debug)

        for _, t := range cfg.Targets {
            log.Info().Str("stage", "run").Str("domain", t.Domain).Msg("starting target")
            ctx := context.Background()
            ctx, cancel := context.WithTimeout(ctx, 24*time.Hour)
            defer cancel()
            if err := pipeline.Run(ctx, cfg, t, force); err != nil {
                return fmt.Errorf("pipeline failed for %s: %w", t.Domain, err)
            }
            log.Info().Str("domain", t.Domain).Msg("target completed")
        }
        return nil
    },
}

