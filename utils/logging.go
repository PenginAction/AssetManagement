package utils

import (
	"io"
	"log"
	"os"
)

func SetupLogging(LogFile string) {
	logFile, err := os.OpenFile(LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("file=LogFile err=%s", err.Error())
	}
	// defer logFile.Close()

	multipleLogOutputs := io.MultiWriter(os.Stdout, logFile)

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(multipleLogOutputs)
}