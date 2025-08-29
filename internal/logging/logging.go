package logging

import (
    "os"
    "time"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func Init(debug bool) {
    output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
    log.Logger = zerolog.New(output).With().Timestamp().Logger()
    if debug {
        zerolog.SetGlobalLevel(zerolog.DebugLevel)
    } else {
        zerolog.SetGlobalLevel(zerolog.InfoLevel)
    }
}

