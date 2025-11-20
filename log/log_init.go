package log_init

import (
	"log"
	"os"
)

var logFile *os.File

func InitializeLog() {
	logFile, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
}

func FinalizeLog() {
	logFile.Close()
}
