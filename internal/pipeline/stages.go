package pipeline

import (
    "context"
    "fmt"
    "encoding/json"
    "bufio"
    "os"
    "path/filepath"

    "github.com/rs/zerolog/log"
    "hermetica/internal/config"
    dtool "hermetica/internal/tool/dnsx"
    htool "hermetica/internal/tool/httpx"
    ntool "hermetica/internal/tool/naabu"
    stool "hermetica/internal/tool/subfinder"
)

type Target = config.Target

func Run(ctx context.Context, cfg *config.Config, t Target, force bool) error {
    wdir := filepath.Join(cfg.Workdir, t.Domain)
    if err := os.MkdirAll(wdir, 0o755); err != nil { return err }
    // Stage 1: discover_subdomains
    subsPath := filepath.Join(wdir, "subdomains.jsonl")
    if force || !exists(subsPath) {
        log.Info().Str("stage","discover_subdomains").Str("domain", t.Domain).Msg("running subfinder")
        if err := stool.Run(ctx, cfg, t.Domain, subsPath); err != nil { return fmt.Errorf("subfinder: %w", err) }
    } else { log.Info().Str("stage","discover_subdomains").Msg("skipping (artifact exists)") }

    // Stage 2: resolve_dns
    listPath := filepath.Join(wdir, "subdomains.txt")
    if force || !exists(listPath) { if err := dtool.BuildInputFromSubfinder(subsPath, listPath); err != nil { return err } }
    resolvedPath := filepath.Join(wdir, "resolved.jsonl")
    if force || !exists(resolvedPath) {
        log.Info().Str("stage","resolve_dns").Msg("running dnsx")
        if err := dtool.Run(ctx, cfg, listPath, resolvedPath); err != nil { return fmt.Errorf("dnsx: %w", err) }
    } else { log.Info().Str("stage","resolve_dns").Msg("skipping (artifact exists)") }

    // Stage 3: scan_ports
    ipsPath := filepath.Join(wdir, "ips.txt")
    if force || !exists(ipsPath) { if err := ntool.BuildIPsFromDNSX(resolvedPath, ipsPath, cfg.DNS.IPv6Enabled || t.IPv6Enabled); err != nil { return err } }
    portsPath := filepath.Join(wdir, "ports.jsonl")
    if force || !exists(portsPath) {
        log.Info().Str("stage","scan_ports").Msg("running naabu")
        if err := ntool.Run(ctx, cfg, ipsPath, portsPath); err != nil { return fmt.Errorf("naabu: %w", err) }
    } else { log.Info().Str("stage","scan_ports").Msg("skipping (artifact exists)") }

    // Stage 4: probe_http (basic version)
    // Derive host:port list for httpx input using resolved hosts and open ports.
    // For v1 minimal, probe IP:port directly. Host/SNI matrix will be added in a follow-up.
    hpList := filepath.Join(wdir, "targets.txt")
    if force || !exists(hpList) {
        if err := buildIPPortList(portsPath, hpList); err != nil { return err }
    }
    webPath := filepath.Join(wdir, "web.jsonl")
    if force || !exists(webPath) {
        log.Info().Str("stage","probe_http").Msg("running httpx")
        if err := htool.RunBasic(ctx, cfg, hpList, webPath); err != nil { return fmt.Errorf("httpx: %w", err) }
    } else { log.Info().Str("stage","probe_http").Msg("skipping (artifact exists)") }

    // TODO: optional stages (TLS SAN feedback, vhost brute, crawl, screenshots)

    // Write run.meta.json
    _ = writeRunMeta(filepath.Join(wdir, "run.meta.json"), cfg)
    return nil
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

// buildIPPortList creates a list of ip:port pairs from naabu JSONL output.
func buildIPPortList(portsJSONL, outList string) error {
    in, err := os.Open(portsJSONL)
    if err != nil { return err }
    defer in.Close()
    out, err := os.Create(outList)
    if err != nil { return err }
    defer out.Close()
    sc := bufio.NewScanner(in)
    type rec struct { IP string `json:"ip"`; Port int `json:"port"` }
    for sc.Scan() {
        var r rec
        if err := json.Unmarshal(sc.Bytes(), &r); err == nil && r.IP != "" && r.Port != 0 {
            _, _ = out.WriteString(fmt.Sprintf("%s:%d\n", r.IP, r.Port))
        }
    }
    return sc.Err()
}
