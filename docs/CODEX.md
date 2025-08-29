# Codex Brief — Hermetica

Purpose: give Codex CLI clear, actionable context to continue building Hermetica on the VPS.

## Goal
- Build a Linux-only Go (1.22+) CLI that maps a domain’s web attack surface using ProjectDiscovery tools, with safe defaults for red-team work and reproducible artifacts.

## Environment
- Host: Debian VPS (Linux x86_64) with raw socket support.
- Go: 1.24.x (installed by `scripts/bootstrap-debian.sh`).
- Tools (minimum versions; installed to `$HOME/go/bin` or `/usr/local/bin`):
  - subfinder >= 2.8.0, dnsx >= 1.2.2, naabu >= 2.3.5, httpx >= 1.7.1, katana >= 1.2.1, gowitness >= 3.0.5.
- Config file: `configs/hermetica.yaml` (paths pinned via `hermetica doctor --fix-paths`).

## What’s Implemented
- CLI (Cobra): `run`, `resume`(stub), `export`(stub), `doctor`.
  - `internal/cmd/root.go`, `internal/cmd/run.go`, `internal/cmd/doctor.go`.
- Config loader + validation scaffolding: `internal/config/config.go`.
- Logging: `internal/logging/logging.go` (zerolog).
- Store (SQLite, pure-Go): `internal/store/store.go` (migrations + WAL).
- Exec wrapper (JSONL streaming): `internal/executil/executil.go`.
- Pipeline (M2-lite):
  - Stages: subfinder → dnsx → naabu → httpx (baseline probing on ip:port).
  - Files: `internal/pipeline/stages.go`, `internal/pipeline/meta.go`.
- Tool wrappers:
  - subfinder: `internal/tool/subfinder/subfinder.go`
  - dnsx: `internal/tool/dnsx/dnsx.go`
  - naabu: `internal/tool/naabu/naabu.go` (profile: stealth=connect, thorough=SYN; rate configurable)
  - httpx: `internal/tool/httpx/httpx.go` (basic flags)
- Docs & defaults: `PRD.md` (updated), `docs/tools.md` (integration guide), `configs/resolvers.txt`, `.gitignore`, `Makefile`, `README.md`.

Build status: `go build ./cmd/hermetica` succeeds.

## How To Run
- Bootstrap (one-time): `bash scripts/bootstrap-debian.sh`
- Build: `make build`
- Doctor: `./bin/hermetica doctor --fix-paths --dry-run -c configs/hermetica.yaml`
- Run: `./bin/hermetica run -d example.com -c configs/hermetica.yaml`
- Thorough/SYN scans: `sudo setcap cap_net_raw+ep $(which naabu)` or run with sudo and `--profile thorough`.

Artifacts per target under `work/<domain>/`:
- `subdomains.jsonl`, `subdomains.txt`, `resolved.jsonl`, `ips.txt`, `ports.jsonl`, `targets.txt`, `web.jsonl`, `run.meta.json`.

## Next Steps (Prioritized)
1) HTTP probing matrix
- Generate and probe SNI/Host combinations + direct IP per config `probe_matrix`.
- Ensure scope regex and CIDR allowlists filter candidate hosts.
- Promote “unique app” grouping via BodyHash/PageGroup.

2) TLS SAN feedback (optional stage)
- Extract SANs from httpx TLS data, filter by scope, re-resolve → enqueue to discovery.
- Bound rounds by `stages.tls_san_feedback.max_rounds`.

3) VHost brute (optional stage)
- For likely web ports, brute Host headers from wordlist with scope enforcement.
- Merge discoveries into probing queue.

4) Export command
- Implement CSV/JSON exports for Assets, Services, WebTargets per PRD schema.

5) Resume logic
- Respect existing canonical artifacts and stage skip unless `--force`.
- Verify determinism: de-dupe and stable sort before finalizing files.

6) Adaptive scanning
- Wire naabu rate adjustments from observed timeouts/loss (configurable thresholds).

7) Evidence & screenshots
- Optionally store body samples and integrate gowitness with rate limiting.

8) Crawling (katana)
- Run on unique apps only, depth/time bounded; write `crawl.jsonl`.

9) Doctor enhancements
- Record tool versions in `run.meta.json`; add more dry-run checks.

## Definition of Done (near-term)
- M2 complete: Core pipeline + artifacts + baseline HTTP probing matrix + determinism + resume.
- M3 complete: Exports + SAN feedback + vhost brute.
- M4 complete: Screenshots + crawling + polish.

## Notes & Guardrails
- Linux-only v1; IPv6 disabled by default (enable per target/config).
- Legal scope enforced via regex + CIDR before scanning or probing.
- SYN scans require raw socket capability; fallback to connect scans.

## Open Items (lightweight)
- Finalize default rates for stealth vs thorough profiles.
- Tune httpx timeouts/retries for WAF-heavy targets.
- Decide on preferred BodyHash algo default (xxhash vs sha1) per engagement.

