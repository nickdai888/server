package utils

import (
	"log"
	"os"
)

var debugLog *log.Logger
var gLogFile *os.File

func init() {
	logFileName := "server.log"
	logFile, err := os.Create(logFileName)
	gLogFile = logFile
	if nil != err {
		log.Fatalln("Open log file Err !")
	}
	debugLog = log.New(logFile, "[Debug]", log.Lshortfile|log.Ltime|log.Ldate)
}

func uninit() {
	Log("uninite called")
	gLogFile.Close()
}

func Log(v ...interface{}) {
	debugLog.Println(v...)
}

func CheckError(err error) {
	if err != nil {
		Log("Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
