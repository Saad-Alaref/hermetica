package naabu

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"

    "hermetica/internal/config"
    "hermetica/internal/executil"
)

// Build input IP list from dnsx JSONL
func BuildIPsFromDNSX(resolvedJSONL, outList string, includeIPv6 bool) error {
    in, err := os.Open(resolvedJSONL)
    if err != nil { return err }
    defer in.Close()
    if err := os.MkdirAll(filepath.Dir(outList), 0o755); err != nil { return err }
    out, err := os.Create(outList)
    if err != nil { return err }
    defer out.Close()
    seen := map[string]struct{}{}
    sc := bufio.NewScanner(in)
    for sc.Scan() {
        var obj struct{
            A []string `json:"a"`
            AAAA []string `json:"aaaa"`
        }
        if err := json.Unmarshal(sc.Bytes(), &obj); err == nil {
            for _, ip := range obj.A { if _,ok:=seen[ip]; !ok { seen[ip]=struct{}{}; out.WriteString(ip+"\n") } }
            if includeIPv6 {
                for _, ip := range obj.AAAA { if _,ok:=seen[ip]; !ok { seen[ip]=struct{}{}; out.WriteString(ip+"\n") } }
            }
        }
    }
    return sc.Err()
}

func Run(ctx context.Context, cfg *config.Config, inList, outJSONL string) error {
    if err := os.MkdirAll(filepath.Dir(outJSONL), 0o755); err != nil { return err }
    f, err := os.Create(outJSONL+".tmp")
    if err != nil { return err }
    defer f.Close()
    // Determine scan type based on profile
    scanType := "c" // connect
    if cfg.Scan.Profile == "thorough" { scanType = "s" } // SYN
    args := []string{"-list", inList, "-p", "-", "-s", scanType, "-rate",  fmtInt(cfg.Scan.NaabuRate), "-json"}
    spec := executil.CmdSpec{Path: cfg.Tools.Paths["naabu"], Args: args, Timeout: 24 * time.Hour}
    err = executil.RunJSONL(ctx, spec, func(b []byte) error { _, werr := f.Write(append(b, '\n')); return werr })
    if err != nil { return err }
    f.Close()
    return os.Rename(outJSONL+".tmp", outJSONL)
}

func fmtInt(i int) string { return fmt.Sprintf("%d", i) }
