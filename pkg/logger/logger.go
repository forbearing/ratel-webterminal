package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/forbearing/ratel-webterminal/pkg/args"
	"github.com/sirupsen/logrus"
)

func New() *logrus.Logger {
	logger := logrus.New()

	logLevel := args.GetLogLevel()
	logFormat := args.GetLogFormat()
	logFile := args.GetLogFile()
	fmt.Println(logLevel)
	fmt.Println(logFormat)
	fmt.Println(logFile)

	// set log output, default is os.Stdout.
	switch strings.ToUpper(logLevel) {
	case "ERROR":
		logger.SetLevel(logrus.ErrorLevel)
	case "WARN":
		logger.SetLevel(logrus.WarnLevel)
	case "WARNING":
		logger.SetLevel(logrus.WarnLevel)
	case "INFO":
		logger.SetLevel(logrus.InfoLevel)
	case "DEBUG":
		logger.SetLevel(logrus.DebugLevel)
	case "TRACE":
		logger.SetLevel(logrus.TraceLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// set log format, default is text format.
	switch strings.ToUpper(logFormat) {
	case "TEXT":
		//logger.SetFormatter(&logrus.TextFormatter{})
		logger.SetFormatter(&logrus.JSONFormatter{})
	case "JSON":
		logger.SetFormatter(&logrus.JSONFormatter{})
	default:
		logger.SetFormatter(&logrus.TextFormatter{})
	}

	// set log file, default is os.Stdout.
	switch logFile {
	case "/dev/stdout":
		logger.SetOutput(os.Stdout)
	case "/dev/stderr":
		logger.SetOutput(os.Stderr)
	default:
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
		if err != nil {
			panic(err)
		}
		logger.SetOutput(file)
	}

	//// SetReportCaller sets whether the standard logger will include the calling
	//// method as a field.
	//logger.SetReportCaller(false)

	return logger
}
