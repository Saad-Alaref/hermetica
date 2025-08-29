package config

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "gopkg.in/yaml.v3"
)

type Config struct {
    Project  string        `yaml:"project"`
    Workdir  string        `yaml:"workdir"`
    Database string        `yaml:"database"`
    Targets  []Target      `yaml:"targets"`
    Scope    Scope         `yaml:"scope"`
    Tools    Tools         `yaml:"tools"`
    DNS      DNS           `yaml:"dns"`
    Limits   Limits        `yaml:"limits"`
    Scan     Scan          `yaml:"scan"`
    Stages   Stages        `yaml:"stages"`
    Probe    ProbeMatrix   `yaml:"probe_matrix"`
    Evidence Evidence      `yaml:"evidence"`
    Report   Report        `yaml:"report"`
}

type Target struct {
    Domain            string `yaml:"domain"`
    IncludeSubdomains bool   `yaml:"include_subdomains"`
    IPv6Enabled       bool   `yaml:"ipv6_enabled"`
}

type Scope struct {
    IncludeCIDRs       []string `yaml:"include_cidrs"`
    ExcludeCIDRs       []string `yaml:"exclude_cidrs"`
    AllowedDomainRegex string   `yaml:"allowed_domain_regex"`
    DeniedDomainRegex  string   `yaml:"denied_domain_regex"`
}

type Tools struct {
    Platform       string            `yaml:"platform"`
    Paths          map[string]string `yaml:"paths"`
    Versions       map[string]string `yaml:"versions"`
    ProviderConfig string            `yaml:"provider_config"`
    ResolversFile  string            `yaml:"resolvers_file"`
}

type DNS struct {
    WildcardFilter bool `yaml:"wildcard_filter"`
    VerifyCount    int  `yaml:"verify_count"`
    IPv6Enabled    bool `yaml:"ipv6_enabled"`
}

type Limits struct {
    Concurrency        int `yaml:"concurrency"`
    HTTPXTimeoutSec    int `yaml:"httpx_timeout_seconds"`
    Retries            int `yaml:"retries"`
    RequestJitterMs    int `yaml:"request_jitter_ms"`
    MaxBodyKB          int `yaml:"max_body_kb"`
}

type Scan struct {
    Profile string `yaml:"profile"`
    NaabuRate int  `yaml:"naabu_rate"`
    AdaptiveBackoff AdaptiveBackoff `yaml:"adaptive_backoff"`
}

type AdaptiveBackoff struct {
    Enabled bool    `yaml:"enabled"`
    PacketLossThreshold float64 `yaml:"packet_loss_threshold"`
    BackoffMultiplier   float64 `yaml:"backoff_multiplier"`
    RecoveryMultiplier  float64 `yaml:"recovery_multiplier"`
}

type Stages struct {
    BruteDNS StageBruteDNS `yaml:"brute_dns"`
    TLSSANFeedback StageTLSSAN `yaml:"tls_san_feedback"`
    Screenshots StageScreens `yaml:"screenshots"`
    Crawling StageCrawl `yaml:"crawling"`
    VHostBrute StageVHost `yaml:"vhost_brute"`
}

type StageBruteDNS struct {
    Enabled bool   `yaml:"enabled"`
    Wordlist string `yaml:"wordlist"`
    MaxCandidates int `yaml:"max_candidates"`
}
type StageTLSSAN struct {
    Enabled bool `yaml:"enabled"`
    MaxRounds int `yaml:"max_rounds"`
}
type StageScreens struct {
    Enabled bool `yaml:"enabled"`
    RateLimitPerMin int `yaml:"rate_limit_per_min"`
}
type StageCrawl struct {
    Enabled bool `yaml:"enabled"`
    Katana struct {
        Concurrency int `yaml:"concurrency"`
        TimeoutSeconds int `yaml:"timeout_seconds"`
        MaxDepth int `yaml:"max_depth"`
    } `yaml:"katana"`
}
type StageVHost struct {
    Enabled bool `yaml:"enabled"`
    HostWordlist string `yaml:"host_wordlist"`
    MaxHostsPerIP int `yaml:"max_hosts_per_ip"`
}

type ProbeMatrix struct {
    IncludeDirectIP bool `yaml:"include_direct_ip"`
    SNIHostCombinations []struct {
        SNI  string `yaml:"sni"`
        Host string `yaml:"host"`
    } `yaml:"sni_host_combinations"`
}

type Evidence struct {
    StoreBodyHash     bool   `yaml:"store_body_hash"`
    StoreBodySample   bool   `yaml:"store_body_sample"`
    StoreScreenshots  bool   `yaml:"store_screenshots"`
    BodyHashAlgo      string `yaml:"body_hash_algo"`
    NearDupe          bool   `yaml:"near_dupe"`
}

type Report struct {
    CSV  bool `yaml:"csv"`
    HTML bool `yaml:"html"`
}

func Load(path string) (*Config, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var cfg Config
    if err := yaml.Unmarshal(b, &cfg); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }
    if cfg.Workdir == "" {
        cfg.Workdir = "./work"
    }
    // Expand tilde in provider config
    if cfg.Tools.ProviderConfig != "" && strings.HasPrefix(cfg.Tools.ProviderConfig, "~") {
        home, _ := os.UserHomeDir()
        cfg.Tools.ProviderConfig = filepath.Join(home, strings.TrimPrefix(cfg.Tools.ProviderConfig, "~"))
    }
    return &cfg, nil
}

