package log_init

import (
	"fmt"
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

func PrintAndLogIfError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		log.Println("Error:", err)
	}
}

func PrintAndLog(args ...any) {
	fmt.Println(args...)
	log.Println(args...)
}
