package main

import (
	"log"
	"os"
)

// Sets up logging before the main function executes
func init() {
	logfile, err := os.OpenFile("getwtxt.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("Could not open log file: %v\n", err)
	}

	// Listen for the signal to close the log file
	go func() {
		<-closelog
		log.Printf("Closing log file ...\n")
		err = logfile.Close()
		if err != nil {
			log.Printf("Couldn't close log file: %v\n", err)
		}
	}()

	log.SetOutput(logfile)

}
