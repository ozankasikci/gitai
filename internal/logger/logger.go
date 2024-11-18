package logger

import (
	"io"
	"log"
	"os"
	"github.com/ozankasikci/gitai/internal/config"
)

var (
	Info    *log.Logger
	Error   *log.Logger
	Debug   *log.Logger
	Verbose bool
)

func Init() error {
	cfg := config.Get()
	Verbose = cfg.Logger.Verbose
	
	infoPrefix := "\033[32m[INFO]\033[0m "
	errorPrefix := "\033[31m[ERROR]\033[0m "
	debugPrefix := "\033[34m[DEBUG]\033[0m "

	flags := log.LstdFlags

	Info = log.New(os.Stdout, infoPrefix, flags)
	Error = log.New(os.Stderr, errorPrefix, flags)
	
	var debugWriter io.Writer
	if cfg.Logger.Verbose {
		debugWriter = os.Stdout
	} else {
		debugWriter = io.Discard
	}
	Debug = log.New(debugWriter, debugPrefix, flags)
	
	return nil
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