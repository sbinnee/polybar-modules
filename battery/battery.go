package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

const root = "/sys/class/power_supply/BAT0/"

func formatSeconds(seconds int64) string {
	h := seconds / 3600
	m := int((seconds - h * 3600) / 60)
	return fmt.Sprintf("%02d:%02d", h, m)
}

func parseFloat(path string) float64 {
	buf, err := ioutil.ReadFile(root + path)
	if err != nil {
		log.Fatal(err)
	}
	str := string(buf)
	str = strings.Replace(str, "\n", "", -1)
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Fatal(err)
	}
	return val / 1000
}

func parseString(path string) string {
	buf, err := ioutil.ReadFile(root + path)
	if err != nil {
		log.Fatal(err)
	}
	str := string(buf)
	str = strings.Replace(str, "\n", "", -1)
	return str
}

func main() {
	status := parseString("status")  // "Charging" | "Discharging"
	chargeFull := parseFloat("charge_full")
	chargeNow := parseFloat("charge_now")
	currentNow := parseFloat("current_now")
	capacity := parseString("capacity")

	// fmt.Printf("charge_full : %T %v\n", chargeFull, chargeFull)
	// fmt.Printf("charge_now  : %v\n", chargeNow)
	// fmt.Printf("current_now : %v\n", currentNow)

	var seconds float64
	if status == "Charging" {
		seconds = 3600 * (chargeFull - chargeNow) / currentNow
		fmt.Printf("%v%% %v\n", capacity, formatSeconds(int64(seconds)))
	} else if status == "Discharging" {
		seconds = 3600 * chargeNow / currentNow
		fmt.Printf("%v%% %v\n", capacity, formatSeconds(int64(seconds)))
	} else {
		fmt.Println("FULL")
	}
}
