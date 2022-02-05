package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	const format = "(Mon) Jan 02 🕖 15:04"

	// duration
	interval := os.Getenv("INTERVAL")
	// default: 5 seconds
	if interval == "" {
		interval = "5"
	}
	d, err := strconv.Atoi(interval)
	if err != nil {
		log.Fatal(err)
	}
	duration := time.Duration(d * 1000)

	// timezone
	timezone := os.Getenv("TZ")
	var location *time.Location
	if timezone == "" {
		location, err = time.LoadLocation("Local")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		location, err = time.LoadLocation(timezone)
		if err != nil {
			log.Fatal(err)
		}
	}

	for {
		// time now
		now := time.Now()
		nowInLocation := now.In(location)
		fmt.Printf("%v\n", nowInLocation.Format(format))
		// sleep
		time.Sleep(duration * time.Millisecond)
	}
}
