package logger

import (
	"io"
	"log"
	"os"
)

var (
	Info    *log.Logger
	Error   *log.Logger
	Debug   *log.Logger
	Verbose bool
)

// Init creates new loggers with custom prefixes and writers
func Init(verbose bool) {
	Verbose = verbose
	
	flags := log.LstdFlags | log.Lmsgprefix

	Info = log.New(os.Stdout, "[INFO] ", flags)
	Error = log.New(os.Stderr, "[ERROR] ", flags)
	
	// Debug logger only writes if verbose mode is enabled
	var debugWriter io.Writer
	if verbose {
		debugWriter = os.Stdout
	} else {
		debugWriter = io.Discard
	}
	Debug = log.New(debugWriter, "[DEBUG] ", flags)
} 