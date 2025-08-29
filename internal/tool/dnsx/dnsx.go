package dnsx

import (
    "bufio"
    "context"
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "time"

    "hermetica/internal/config"
    "hermetica/internal/executil"
)

// Build input file from subfinder JSONL to a plain list of hostnames
func BuildInputFromSubfinder(subsJSONL string, outList string) error {
    in, err := os.Open(subsJSONL)
    if err != nil { return err }
    defer in.Close()
    if err := os.MkdirAll(filepath.Dir(outList), 0o755); err != nil { return err }
    out, err := os.Create(outList)
    if err != nil { return err }
    defer out.Close()
    seen := make(map[string]struct{})
    sc := bufio.NewScanner(in)
    for sc.Scan() {
        var obj map[string]any
        if err := json.Unmarshal(sc.Bytes(), &obj); err == nil {
            if h, ok := obj["host"].(string); ok {
                h = strings.TrimSpace(h)
                if h == "" { continue }
                if _, dup := seen[h]; !dup {
                    seen[h] = struct{}{}
                    out.WriteString(h+"\n")
                }
            }
        }
    }
    return sc.Err()
}

func Run(ctx context.Context, cfg *config.Config, inList string, outJSONL string) error {
    if err := os.MkdirAll(filepath.Dir(outJSONL), 0o755); err != nil { return err }
    f, err := os.Create(outJSONL + ".tmp")
    if err != nil { return err }
    defer f.Close()
    args := []string{"-l", inList, "-a", "-cname", "-retry", "2", "-json"}
    if cfg.DNS.IPv6Enabled { args = append(args, "-aaaa") }
    if cfg.Tools.ResolversFile != "" { args = append(args, "-r", cfg.Tools.ResolversFile) }
    spec := executil.CmdSpec{Path: cfg.Tools.Paths["dnsx"], Args: args, Timeout: 60 * time.Minute}
    err = executil.RunJSONL(ctx, spec, func(b []byte) error { _, werr := f.Write(append(b, '\n')); return werr })
    if err != nil { return err }
    f.Close()
    return os.Rename(outJSONL+".tmp", outJSONL)
}

