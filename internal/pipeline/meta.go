package pipeline

import (
    "encoding/json"
    "os"
    "time"
    "hermetica/internal/config"
)

type runMeta struct {
    GeneratedAt time.Time        `json:"generated_at"`
    ConfigWorkdir string         `json:"workdir"`
    ToolVersions map[string]string `json:"tool_versions"`
}

func writeRunMeta(path string, cfg *config.Config) error {
    m := runMeta{
        GeneratedAt: time.Now(),
        ConfigWorkdir: cfg.Workdir,
        ToolVersions: cfg.Tools.Versions,
    }
    b, _ := json.MarshalIndent(m, "", "  ")
    return os.WriteFile(path, b, 0o644)
}

