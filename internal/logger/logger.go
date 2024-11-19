package logger

import (
	"io"
	"log"
	"os"
)

const (
	infoPrefix  = "\033[32m[INFO]\033[0m "
	errorPrefix = "\033[31m[ERROR]\033[0m "
	debugPrefix = "\033[34m[DEBUG]\033[0m "
)

var (
	Info    *log.Logger
	Error   *log.Logger
	Debug   *log.Logger
	Verbose bool
)

func InitDefault() {
	flags := log.LstdFlags | log.Lmsgprefix
	
	Info = log.New(os.Stdout, infoPrefix, flags)
	Error = log.New(os.Stderr, errorPrefix, flags)
	Debug = log.New(io.Discard, debugPrefix, flags)
	Verbose = false
}

func UpdateConfig(verbose bool) {
	println("UpdateConfig")
	println(verbose)
	Verbose = verbose
	if verbose {
		Debug.SetOutput(os.Stdout)
	} else {
		Debug.SetOutput(io.Discard)
	}
	println(Verbose)
}

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

func Infoln(v ...interface{}) {
	Info.Println(v...)
}

func Errorln(v ...interface{}) {
	Error.Println(v...)
}

func Debugln(v ...interface{}) {
	println("Debugln")
	println(Verbose)
	if Verbose {
		Debug.Println(v...)
	}
} 