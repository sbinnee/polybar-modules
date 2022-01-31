package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
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
	for {
		buf, err := ioutil.ReadFile("/proc/loadavg")
		if err != nil {
			log.Fatal(err)
		}
		str := string(buf)
		fields := strings.Fields(str)
		fmt.Printf("%v/%v\n", fields[0], fields[1])
		// sleep
		time.Sleep(duration * time.Millisecond)
	}
}
