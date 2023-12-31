package util

import (
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Info writes logs in the color blue with "INFO: " as prefix
var Info = log.New(os.Stdout, "\u001b[34mINFO: \u001B[0m", log.LstdFlags|log.Lshortfile)

// Warning writes logs in the color yellow with "WARNING: " as prefix
var Warning = log.New(os.Stdout, "\u001b[33mWARNING: \u001B[0m", log.LstdFlags|log.Lshortfile)

// Error writes logs in the color red with "ERROR: " as prefix
var Error = log.New(os.Stdout, "\u001b[31mERROR: \u001b[0m", log.LstdFlags|log.Lshortfile)

// Same as error, but don't print the file and line. It is expected that what is being printed will include a file and line
// Useful for a generic panic cather function, where the line of the catcher function is not useful
var ErrorCatcher = log.New(os.Stdout, "\u001b[31mERROR: \u001b[0m", log.LstdFlags)

// Debug writes logs in the color cyan with "DEBUG: " as prefix
func getDebug() *log.Logger {
	godotenv.Load()
	if ShowDebug() {
		return log.New(os.Stdout, "\u001b[36mDEBUG: \u001B[0m", log.LstdFlags|log.Lshortfile)
	} else {
		return log.New(io.Discard, "\u001b[36mDEBUG: \u001B[0m", log.LstdFlags|log.Lshortfile)
	}
}
var Debug = getDebug()