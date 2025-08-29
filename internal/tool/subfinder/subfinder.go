package subfinder

import (
    "context"
    "os"
    "path/filepath"
    "time"

    "hermetica/internal/config"
    "hermetica/internal/executil"
)

func Run(ctx context.Context, cfg *config.Config, domain string, outJSONL string) error {
    if err := os.MkdirAll(filepath.Dir(outJSONL), 0o755); err != nil { return err }
    f, err := os.Create(outJSONL + ".tmp")
    if err != nil { return err }
    defer f.Close()
    args := []string{"-silent", "-all", "-d", domain, "-json"}
    if cfg.Tools.ProviderConfig != "" { args = append(args, "-pc", cfg.Tools.ProviderConfig) }
    spec := executil.CmdSpec{Path: cfg.Tools.Paths["subfinder"], Args: args, Timeout: 60 * time.Minute}
    err = executil.RunJSONL(ctx, spec, func(b []byte) error {
        _, werr := f.Write(append(b, '\n'))
        return werr
    })
    if err != nil { return err }
    f.Close()
    return os.Rename(outJSONL+".tmp", outJSONL)
}

