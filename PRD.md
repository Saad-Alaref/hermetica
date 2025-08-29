# 📄 PRD — **Hermetica** (Attack Surface Mapper)

## 0) Summary
**Hermetica** is a Go (1.22+) CLI that maps a domain’s web attack surface using ProjectDiscovery (PD) tools, designed for red-team engagements with explicit authorization. It:
1) takes a **domain**,
2) collects **subdomains** (passive + optional brute),
3) resolves to **IPv4** by default (IPv6 optional) with wildcard-aware checks,
4) scans **all TCP ports** with configurable profiles: **stealth** (connect, lower rate) or **thorough** (SYN, higher accuracy), both with configurable rates and adaptive backoff,
5) performs **deterministic, multi-strategy web probing** (SNI/Host matrix + direct IP), across the full open-port set,
6) runs an **optional TLS SAN feedback loop** (scope-safe),
7) captures **evidence** (hashes, optional body samples, optional screenshots) and groups near-duplicates,
8) stores results in **SQLite** (pure-Go driver), writes **idempotent JSONL artifacts**, and supports **resume**,
9) includes **pre-flight health checks** for all binaries, versions, configs, and a **dry-run** mode.

Assumptions & Constraints
- Platform: Linux (x86_64) for v1.
- Multi-target execution: sequential per `targets[]` entry.
- Legal scope: operator holds authorization for port scanning and probing.

---

## 1) CLI Design (Cobra)
**Commands**
- `hermetica run -c configs/hermetica.yaml -d example.com`
  - Executes pipeline (stages configurable in YAML).
- `hermetica resume -w work/example.com/`
  - Resumes using existing canonical artifacts.
- `hermetica export --format csv|json --out out/`
  - Exports from SQLite.
- `hermetica doctor`
  - Pre-flight: checks paths/versions of PD tools, provider config, resolvers, DB write perms.
  - Flags: `--dry-run` runs short, safe invocations for each tool to validate exec + JSONL parsing.

**Global Flags**
- `-c` config path (default: `configs/hermetica.yaml`)
- `-d` domain override
- `-w` workdir (default: `./work/`)
- `--force` (re-run stages even if artifacts exist)
- `--debug` (verbose logs)
- `--profile stealth|thorough` (override scanning profile)

---

## 2) Configuration (YAML → Go structs)
```yaml
project: "Hermetica"
workdir: "./work"
database: "./work/hermetica.sqlite"

targets:
  - domain: "example.com"
    include_subdomains: true
    ipv6_enabled: false              # per-target override

scope:
  include_cidrs: []                  # optional allowlist of IP ranges
  exclude_cidrs: []
  allowed_domain_regex: ".*"         # enforce on CT/SAN & vhost attempts
  denied_domain_regex: ""

tools:
  platform: "linux"                  # v1 supports Linux only
  paths:
    subfinder: "/usr/local/bin/subfinder"
    dnsx:      "/usr/local/bin/dnsx"
    naabu:     "/usr/local/bin/naabu"
    httpx:     "/usr/local/bin/httpx"
    katana:    "/usr/local/bin/katana"
    gowitness: "/usr/local/bin/gowitness"    # optional
  versions:
    subfinder: ">=2.6.0"
    dnsx:      ">=1.2.0"
    naabu:     ">=2.3.0"
    httpx:     ">=1.3.5"
    katana:    ">=1.1.0"
    gowitness: ">=2.5.0"
  provider_config: "~/.config/subfinder/provider-config.yaml"
  resolvers_file: "./configs/resolvers.txt"

dns:
  wildcard_filter: true
  verify_count: 2                     # require consistent answers N times
  ipv6_enabled: false                 # AAAA + scan IPv6 (default off)

limits:
  concurrency: 200
  httpx_timeout_seconds: 8
  retries: 1
  request_jitter_ms: 250              # jitter to avoid WAF spikes
  max_body_kb: 128                    # body sample cap for hashing

scan:
  profile: "stealth"                 # stealth | thorough (overridable via --profile)
  naabu_rate: 4000                    # default rate, configurable
  adaptive_backoff:
    enabled: true
    packet_loss_threshold: 0.10       # reduce rate if above
    backoff_multiplier: 0.5           # cut rate by 50% on loss
    recovery_multiplier: 1.25         # slowly recover rate

stages:
  brute_dns:
    enabled: false                    # optional; when true uses puredns-like flow via dnsx
    wordlist: "./configs/words.txt"
    max_candidates: 200000            # safety valve
  tls_san_feedback:
    enabled: true
    max_rounds: 1                     # bounded enrichment
  screenshots:
    enabled: false
    rate_limit_per_min: 60
  crawling:
    enabled: false
    katana:
      concurrency: 10
      timeout_seconds: 8
      max_depth: 2
  vhost_brute:
    enabled: false
    host_wordlist: "./configs/vhost-words.txt"
    max_hosts_per_ip: 100

probe_matrix:
  include_direct_ip: true
  sni_host_combinations:
    - { sni: "subdomain", host: "subdomain" }
    - { sni: "subdomain", host: "" }
    - { sni: "",          host: "subdomain" }
    - { sni: "",          host: "" }

evidence:
  store_body_hash: true
  store_body_sample: false            # off by default; toggle per engagement
  store_screenshots: false            # mirrors screenshots.enabled
  body_hash_algo: "xxhash"            # or "sha1"
  near_dupe: true                     # enable shingle/SimHash grouping

report:
  csv: true
  html: false
```

---

## 3) External Tools & Expected JSON
Minimum Supported Versions (Linux)
- subfinder >= 2.6.0, dnsx >= 1.2.0, naabu >= 2.3.0, httpx >= 1.3.5, katana >= 1.1.0, gowitness >= 2.5.0

### subfinder
**Cmd:** `subfinder -silent -all -d <domain> -json`  
**JSONL Example:**
```json
{"host":"api.example.com","source":"securitytrails"}
```

### dnsx
**Cmd:** `dnsx -l subdomains.txt -a -aaaa -cname -resp -retries 2 -r <resolvers.txt> -json`  
**JSONL Example:**
```json
{"host":"api.example.com","a":["1.2.3.4"],"aaaa":["2001:db8::1"],"cname":"api.gslb.example.net"}
```

### naabu
**Cmd:** `naabu -host <ip_or_file> -p - -rate <rate> -retries 1 -json`  
**JSONL Example:**
```json
{"host":"1.2.3.4","ip":"1.2.3.4","port":8443,"protocol":"tcp"}
```

### httpx
**Cmd:**  
```
httpx -json -follow-redirects -title -status-code -tech-detect -tls-grab       -no-color -silent -retries 1 -timeout 8
```
**JSONL Example:**
```json
{
  "input": "api.example.com:8443",
  "host": "api.example.com",
  "port": 8443,
  "scheme": "https",
  "url": "https://api.example.com:8443/",
  "status_code": 200,
  "title": "API Gateway",
  "final_url": "https://api.example.com/",
  "webserver": "nginx",
  "technologies": ["nginx","Java"],
  "tls": {"issuer_cn":"Let's Encrypt","subject_cn":"api.example.com","dns_names":["api.example.com","cdn.example.com"]}
}
```

### katana (optional)
**Cmd:** `katana -silent -jsonl -no-color -concurrency 10 -timeout 8 -list live_urls.txt`  
**JSONL Example:**
```json
{"request":{"url":"https://app.example.com/"},"output":"https://app.example.com/login"}
```

---

## 4) Project Structure
```
hermetica/
├─ cmd/hermetica/            # cobra CLI: run, resume, export, doctor
├─ internal/
│  ├─ config/                # YAML loader & validation
│  ├─ logging/               # zerolog init
│  ├─ store/                 # sqlite (migrations, upserts)
│  ├─ model/                 # data structs
│  ├─ executil/              # os/exec wrapper (timeouts, retries, JSONL streaming)
│  ├─ tool/
│  │  ├─ subfinder/
│  │  ├─ dnsx/
│  │  ├─ naabu/
│  │  ├─ httpx/
│  │  └─ katana/
│  ├─ pipeline/
│  │  ├─ stages.go           # interfaces + wiring
│  │  ├─ discover_subs.go
│  │  ├─ brute_subs.go       # optional
│  │  ├─ resolve_dns.go
│  │  ├─ scan_ports.go
│  │  ├─ probe_http.go
│  │  ├─ expand_vhosts.go    # optional
│  │  ├─ san_feedback.go     # optional
│  │  └─ crawl.go            # optional
│  └─ report/                # exports
├─ configs/hermetica.yaml
├─ configs/resolvers.txt
├─ Makefile
└─ README.md
```

---

## 5) Data Model
```go
type Asset struct {
  ID        string
  Domain    string
  Subdomain string
  FQDN      string
  IP        string
  RRType    string
  FirstSeen time.Time
  LastSeen  time.Time
  Sources   []string
}

type Service struct {
  AssetID string
  IP      string
  Port    int
  Proto   string
  IsWeb   bool
}

type WebTarget struct {
  ServiceID string
  InputHost string
  SNIMode   string
  URL       string
  Status    int
  Title     string
  FinalURL  string
  TLSIssuer string
  CDNHint   string
  Tech      []string
  BodyHash  string
  PageGroup string
  BodyPath  string
  ShotPath  string
}

type Discovery struct {
  Source   string
  Hostname string
  InScope  bool
  Note     string
  SeenAt   time.Time
}
```

IDs, Uniqueness & Indexes
- IDs: ULID strings for `Asset`, `Service`, `WebTarget` (sortable, unique, stored as TEXT in SQLite).
- `Asset` uniqueness: (`FQDN`, `IP`, `RRType`). Index on (`FQDN`), (`IP`); UNIQUE (`FQDN`,`IP`,`RRType`).
- `Service` uniqueness: (`IP`, `Port`, `Proto`). Index on (`IP`,`Port`).
- `WebTarget` uniqueness: (`ServiceID`, `SNIMode`, `InputHost`, `URL`). Index on (`InputHost`,`URL`), (`Status`).
- Timestamps: set `FirstSeen` at insert; update `LastSeen` on upsert.
- Sources: normalized via `Discovery` table; keep aggregated source hints in JSON for exports where helpful.

SQLite Driver
- Use `modernc.org/sqlite` (pure-Go) for portability and simpler builds on Linux.

---

## 6) Pipeline & Behavior
- Multi-target: process `targets[]` sequentially.  
- **discover_subdomains**: passive via subfinder.  
- **brute_subs** (optional): wordlist → dnsx.  
- **resolve_dns**: resolve A (+AAAA optional), wildcard filter, `verify_count`.  
  - Wildcard detection: probe N random labels; if consistent answers across `verify_count`, mark wildcard and filter accordingly.
- **scan_ports**: naabu across all TCP ports with profiles:
  - Stealth: connect scan, lower `scan.naabu_rate`, jittered scheduling; fallback when raw sockets are unavailable.
  - Thorough: SYN scan (requires CAP_NET_RAW), higher accuracy; still rate-limited with adaptive backoff.
  - Adaptive backoff: monitor timeouts/packet loss; reduce/increase rate per config.
- **probe_http**: httpx with 4-way SNI/Host matrix + direct IP, retries within `limits.retries`, WAF-aware jitter/backoff. Probe full open-port set.  
- **expand_vhosts** (optional): brute host headers. Scope-enforced by regex and CIDRs.  
- **tls_san_feedback** (optional): collect SANs → scope filter → re-resolve (bounded rounds).  
- **crawl** (optional): katana on unique apps only (dedupe by `BodyHash`/`PageGroup`).

---

## 7) Evidence & Uniqueness
- Store `BodyHash` (xxhash/sha1).  
- Near-dupe grouping via SimHash.  
- Screenshots optional via gowitness.  
- Body sample optional.  
- Distinguish **unique app** vs **host binding**.

Storage Layout (under `work/<domain>/`)
- Bodies: `bodies/<ulid>.bin` (first `max_body_kb` bytes when enabled).
- Screens: `shots/<ulid>.png` (when enabled; rate-limited by config).

---

## 8) Artifacts & Resume
- Stage outputs to canonical JSONL in `work/<domain>/`:
  - `subdomains.jsonl`, `resolved.jsonl`, `ports.jsonl`, `web.jsonl`, `vhosts.jsonl`, `san.jsonl`, `crawl.jsonl`.
- Determinism: de-duplicate and sort records in a canonical order before finalizing artifacts.
- Temp file then atomic rename.  
- Resume: skip a stage if its artifact exists, regardless of DB contents; `--force` re-runs and overwrites.
- Run metadata: `run.meta.json` includes config hash, tool versions, timestamps, and stage summaries.

---

## 9) Pre-flight Checks
- `hermetica doctor`: checks PD binaries, pinned versions, configs, DB write perms. Hard-fails on missing/invalid tools.  
- `--dry-run`: executes short, safe validations (e.g., `--version`/`-h` or minimal no-network invocations) and verifies JSONL parsing + temp write permissions.  
- Logs versions at startup and stores them in `run.meta.json`.

---

## 10) Reporting
- Exports CSV/JSON, grouped by Assets/Services/WebTargets.  
- CSV columns:
  - assets.csv: `id,domain,subdomain,fqdn,ip,rrtype,first_seen,last_seen`
  - services.csv: `asset_id,ip,port,proto,is_web`
  - webtargets.csv: `service_id,input_host,sni_mode,url,status,title,final_url,tls_issuer,cdn_hint,tech,body_hash,page_group,body_path,shot_path`
- Export filters: `hermetica export --domain example.com --only-web` supported.
- HTML dashboard: future.

---

## 11) Makefile
- `tools`: verify binaries.  
- `build`: compile Hermetica.  
- `run`: run pipeline.  
- `resume`: resume run.  
- `export`: export results.  
- `lint`, `test`.

Notes
- `make tools` enforces minimum tool versions defined in config; fails hard if mismatched.

---

## 12) Logging
- zerolog with `stage`, `domain`, `ip`, `port`.  
- Summaries per stage: in/out counts.

---

## 13) Deliverables
- Go repo with required 4 stages.  
- Optional stages stubs.  
- SQLite store.  
- Exec wrapper.  
- Pre-flight check.  
- Configs, resolvers.txt.  
- README quick start.

Additional Decisions
- OS support: Linux-only in v1 (documented in README).  
- SQLite driver: `modernc.org/sqlite` (pure-Go).  
- Sequential multi-target execution to minimize cross-target noise.

Risks & Constraints
- Privileges: Thorough/SYN scanning requires CAP_NET_RAW; tool falls back to stealth/connect if unavailable.  
- Performance: Full TCP + matrix probing can be heavy; conservative defaults with adaptive backoff reduce impact.  
- Scope safety: SAN/vhost-derived candidates must match `allowed_domain_regex` and not match `denied_domain_regex` and must satisfy CIDR scope before scanning.  
- Ethical/Legal: Operator is responsible for maintaining explicit authorization.
