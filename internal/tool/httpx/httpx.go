package httpx

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "time"

    "hermetica/internal/config"
    "hermetica/internal/executil"
)

func RunBasic(ctx context.Context, cfg *config.Config, inList, outJSONL string) error {
    if err := os.MkdirAll(filepath.Dir(outJSONL), 0o755); err != nil { return err }
    f, err := os.Create(outJSONL+".tmp")
    if err != nil { return err }
    defer f.Close()
    args := []string{"-json", "-fr", "-title", "-sc", "-tech-detect", "-tls-grab", "-no-color", "-silent", "-retries", intToStr(cfg.Limits.Retries), "-timeout", intToStr(cfg.Limits.HTTPXTimeoutSec), "-list", inList}
    spec := executil.CmdSpec{Path: cfg.Tools.Paths["httpx"], Args: args, Timeout: 24 * time.Hour}
    err = executil.RunJSONL(ctx, spec, func(b []byte) error { _, werr := f.Write(append(b, '\n')); return werr })
    if err != nil { return err }
    f.Close()
    return os.Rename(outJSONL+".tmp", outJSONL)
}

func intToStr(i int) string { return fmt.Sprintf("%d", i) }
