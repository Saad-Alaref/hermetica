package cmd

import (
    "fmt"
    "os/exec"
    "strings"

    "hermetica/internal/config"
    "github.com/Masterminds/semver/v3"
    "github.com/rs/zerolog/log"
    "github.com/spf13/cobra"
    "os"
    "path/filepath"
    "gopkg.in/yaml.v3"
)

var dryRun bool
var fixPaths bool

var doctorCmd = &cobra.Command{
    Use:   "doctor",
    Short: "Pre-flight checks for external tools and environment",
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg, err := config.Load(cfgPath)
        if err != nil {
            return err
        }
        // Optionally attempt to auto-detect tool paths on PATH
        if fixPaths {
            for name, p := range cfg.Tools.Paths {
                if p == "" || !fileExists(p) {
                    if found, err := exec.LookPath(name); err == nil {
                        cfg.Tools.Paths[name] = found
                        log.Info().Str("tool", name).Str("path", found).Msg("auto-detected path")
                    }
                }
            }
            // Persist updated config
            if err := writeConfig(cfgPath, cfg); err != nil {
                return fmt.Errorf("update config paths: %w", err)
            }
        }

        // Verify tool paths exist and versions meet minimums
        for name, path := range cfg.Tools.Paths {
            min, ok := cfg.Tools.Versions[name]
            if !ok {
                continue
            }
            if path == "" {
                return fmt.Errorf("tool %s path not set", name)
            }
            // Run version command
            out, verr := exec.Command(path, "-version").CombinedOutput()
            if verr != nil {
                // Fallbacks
                out2, verr2 := exec.Command(path, "--version").CombinedOutput()
                if verr2 != nil {
                    return fmt.Errorf("%s version check failed: %v", name, verr)
                }
                out = out2
            }
            verText := strings.TrimSpace(string(out))
            parsed, perr := parseSemver(verText)
            if perr != nil {
                log.Warn().Str("tool", name).Str("raw", verText).Msg("could not parse version; skipping strict compare")
            } else {
                constraint, cErr := semver.NewConstraint(min)
                if cErr != nil {
                    return cErr
                }
                if !constraint.Check(parsed) {
                    return fmt.Errorf("%s version %s does not satisfy %s", name, parsed.String(), min)
                }
            }
            log.Info().Str("tool", name).Str("path", path).Msg("ok")
        }

        if dryRun {
            // Minimal dry-runs / health checks
            tryRun(cfg.Tools.Paths["dnsx"], "-hc")
            tryRun(cfg.Tools.Paths["httpx"], "-hc")
            tryRun(cfg.Tools.Paths["katana"], "-hc")
        }
        log.Info().Msg("doctor checks passed")
        return nil
    },
}

func init() {
    doctorCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Run tool health checks/dry run")
    doctorCmd.Flags().BoolVar(&fixPaths, "fix-paths", false, "Auto-detect tool paths from PATH and write back to config")
}

func tryRun(bin string, args ...string) {
    if bin == "" {
        return
    }
    _ = exec.Command(bin, args...).Run()
}

func parseSemver(s string) (*semver.Version, error) {
    // Attempt to extract something that looks like vX.Y.Z
    fields := strings.Fields(s)
    for _, f := range fields {
        f = strings.TrimPrefix(f, "v")
        if v, err := semver.NewVersion(f); err == nil {
            return v, nil
        }
    }
    // Try as-is
    if strings.HasPrefix(s, "v") {
        s = strings.TrimPrefix(s, "v")
    }
    return semver.NewVersion(s)
}

func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }

func writeConfig(path string, cfg *config.Config) error {
    b, err := yaml.Marshal(cfg)
    if err != nil { return err }
    // ensure dir exists
    _ = os.MkdirAll(filepath.Dir(path), 0o755)
    return os.WriteFile(path, b, 0o644)
}
