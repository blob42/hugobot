package logging

import (
	"log"
	"os"
)

const (
	logFileName = ".log"
)

var (
	logFile *os.File
)

func Close() error {
	return logFile.Close()
}

func init() {
	//var err error
	//logFile, err = os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	//if err != nil {
	//log.Fatal(err)
	//}

	//log.SetOutput(io.MultiWriter(logFile, os.Stdout))
	log.SetOutput(os.Stdout)
}
