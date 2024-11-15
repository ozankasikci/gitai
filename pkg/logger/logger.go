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
	
	// Add color codes for different log levels
	infoPrefix := "\033[32m[INFO]\033[0m "  // Green
	errorPrefix := "\033[31m[ERROR]\033[0m " // Red
	debugPrefix := "\033[34m[DEBUG]\033[0m " // Blue

	flags := log.LstdFlags

	// Create multi-writer for Info to write to both file and stdout
	Info = log.New(os.Stdout, infoPrefix, flags)
	Error = log.New(os.Stderr, errorPrefix, flags)
	
	// Debug logger only writes if verbose mode is enabled
	var debugWriter io.Writer
	if verbose {
		debugWriter = os.Stdout
	} else {
		debugWriter = io.Discard
	}
	Debug = log.New(debugWriter, debugPrefix, flags)

	// Log initial state
	Info.Printf("Logger initialized (verbose: %v)", verbose)
}

// Helper functions for consistent logging
func Infof(format string, v ...interface{}) {
	Info.Printf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	Error.Printf(format, v...)
}

func Debugf(format string, v ...interface{}) {
	if Verbose {
		Debug.Printf(format, v...)
	}
}

// Print functions for when format isn't needed
func Infoln(v ...interface{}) {
	Info.Println(v...)
}

func Errorln(v ...interface{}) {
	Error.Println(v...)
}

func Debugln(v ...interface{}) {
	if Verbose {
		Debug.Println(v...)
	}
} 