package executil

import (
    "bufio"
    "context"
    "os/exec"
    "time"
)

type LineHandler func([]byte) error

type CmdSpec struct {
    Path string
    Args []string
    Timeout time.Duration
    Env []string
    Dir string
}

func RunJSONL(ctx context.Context, spec CmdSpec, onLine LineHandler) error {
    cctx := ctx
    var cancel context.CancelFunc
    if spec.Timeout > 0 {
        cctx, cancel = context.WithTimeout(ctx, spec.Timeout)
        defer cancel()
    }
    cmd := exec.CommandContext(cctx, spec.Path, spec.Args...)
    if spec.Env != nil { cmd.Env = append(cmd.Env, spec.Env...) }
    if spec.Dir != "" { cmd.Dir = spec.Dir }

    stdout, err := cmd.StdoutPipe()
    if err != nil { return err }
    stderr, err := cmd.StderrPipe()
    if err != nil { return err }
    if err := cmd.Start(); err != nil { return err }

    // Drain stderr to avoid blocking
    go func(){
        scanner := bufio.NewScanner(stderr)
        for scanner.Scan() { /* discard or log */ }
    }()

    scanner := bufio.NewScanner(stdout)
    for scanner.Scan() {
        if err := onLine(scanner.Bytes()); err != nil { return err }
    }
    if err := scanner.Err(); err != nil { return err }
    return cmd.Wait()
}

