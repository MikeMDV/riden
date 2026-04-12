package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	zerolog.Logger
}

const timeFormat string = "2006-01-02 15:04:05.000000"

func InitializeLogger(logFile, levelStr string) (Logger, error) {
	var logger Logger
	// Open file
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", logFile, ":", err)
		return logger, err
	}

	fmt.Fprintf(f, "\n\n------- Re-starting logging:  %s  -------\n\n",
		time.Now().Format(timeFormat))

	// Set logging level
	var level zerolog.Level
	switch levelStr {
	case zerolog.LevelTraceValue:
		level = zerolog.TraceLevel

	case zerolog.LevelDebugValue:
		level = zerolog.DebugLevel

	case zerolog.LevelInfoValue:
		level = zerolog.InfoLevel

	case zerolog.LevelWarnValue:
		level = zerolog.WarnLevel

	case zerolog.LevelErrorValue:
		level = zerolog.ErrorLevel

	case zerolog.LevelFatalValue:
		level = zerolog.FatalLevel

	case zerolog.LevelPanicValue:
		level = zerolog.PanicLevel

	default:
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	zerolog.TimeFieldFormat = timeFormat
	// Set caller to Lshortfile format
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	logger = Logger{
		zerolog.New(f).With().Timestamp().Caller().Logger(),
	}

	return logger, err
}
