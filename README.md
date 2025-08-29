# Hermetica

Hermetica is a Go (1.22+) CLI that maps a domain's web attack surface using ProjectDiscovery tools.

- Commands: `run`, `resume`, `export`, `doctor`
- Platform: Linux (x86_64)

## Quick Start

- Edit `configs/hermetica.yaml` to set tool paths and versions.
- Ensure PD tools are installed: subfinder, dnsx, naabu, httpx (and optional katana, gowitness)

```
make build
./bin/hermetica doctor --dry-run
./bin/hermetica run -d example.com
```

Artifacts are written to `work/<domain>/`.

See `PRD.md` and `docs/tools.md` for details.
