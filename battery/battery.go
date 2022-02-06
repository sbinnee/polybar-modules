package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	POLYBAR_COLOR bool   = true // color format
	root          string = "/sys/class/power_supply/BAT0/"
	C_CRITICAL    string = "#ff4500"
	C_WARNING     string = "#ffa500"
	C_CAUTION     string = "#ffff00"
	C_GOOD        string = "#00fa9a"
	T_FORMAT      string = "2006-01-02 15:04:05"
)

func parseStoHM(s float64) (int64, int64) {
	h := int64(s) / 3600
	m := (s - float64(h)*3600) / 60
	return int64(h), int64(m)
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

// https://stackoverflow.com/questions/17629451/append-slice-to-csv-golang
func addrow(fname string, row []string) error {
	// read the file
	f, err := os.OpenFile(fname, os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	r := csv.NewReader(f)
	lines, err := r.ReadAll()
	if err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	lines = append(lines, row)

	// write
	f, err = os.Create(fname)
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)
	if err = w.WriteAll(lines); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func floatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', -1, 64)
}

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
	var now time.Time
	var status string
	var chargeFull float64 = 0
	var chargeNow float64 = 0
	var currentNow float64 = 0
	var capacity int = 0
	var seconds float64
	var h int64
	var m int64
	var voltageNow float64 = 0
	var wattage float64 = 0
	var strWattage string
	row := make([]string, 8)

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	logdir := path.Join(homedir, ".cache", "battery")
	if _, err := os.Stat(logdir); os.IsNotExist(err) {
		os.Mkdir(logdir, 0755)
	}

	for {
		now = time.Now()
		status = parseString("status") // "Charging" | "Discharging"
		chargeFull = parseFloat("charge_full")
		chargeNow = parseFloat("charge_now")
		currentNow = parseFloat("current_now")
		capacity, err = strconv.Atoi(parseString("capacity"))
		if err != nil {
			log.Fatal(err)
		}

		// fmt.Printf("charge_full : %T %v\n", chargeFull, chargeFull)
		// fmt.Printf("charge_now  : %v\n", chargeNow)
		// fmt.Printf("current_now : %v\n", currentNow)

		if status == "Charging" {
			seconds = 3600 * (chargeFull - chargeNow) / currentNow
			h, m = parseStoHM(seconds)
			if capacity >= 100 {
				if POLYBAR_COLOR {
					fmt.Printf("%%{F%s}%s%%{F-}\n", C_GOOD, "FULL")
				} else {
					fmt.Println("FULL")
				}
			} else {
				if POLYBAR_COLOR {
					fmt.Printf("%%{F%s}%v%% %%{F-}%02d:%02d\n", C_GOOD, capacity, h, m)
				} else {
					fmt.Printf("%v%% %02d:%02d\n", capacity, h, m)
				}
			}
		} else if status == "Discharging" {
			seconds = 3600 * chargeNow / currentNow
			h, m = parseStoHM(seconds)
			voltageNow = parseFloat("voltage_now")
			wattage = currentNow * voltageNow / math.Pow10(6)
			if POLYBAR_COLOR {
				if wattage > 10 {
					strWattage = fmt.Sprintf("%%{F%s}%.1fW%%{F-}", C_CRITICAL, wattage)
				} else if wattage > 8 {
					strWattage = fmt.Sprintf("%%{F%s}%.1fW%%{F-}", C_WARNING, wattage)
				} else if wattage > 6 {
					strWattage = fmt.Sprintf("%%{F%s}%.1fW%%{F-}", C_CAUTION, wattage)
				} else {
					strWattage = fmt.Sprintf("%.1fW", wattage)
				}
				if h < 1 {
					fmt.Printf("%%{F%s}%v%% %%{F-}%02d:%02d (%s)\n", C_CRITICAL, capacity, h, m, strWattage)
				} else {
					fmt.Printf("%v%% %02d:%02d (%s)\n", capacity, h, m, strWattage)
				}
			} else {
				fmt.Printf("%v%% %02d:%02d\n", capacity, h, m)
			}
		} else {
			if POLYBAR_COLOR {
				fmt.Printf("%%{F%s}%s%%{F-}\n", C_GOOD, "FULL")
			} else {
				fmt.Println("FULL")
			}
		}
		// Write to file
		row[0] = now.Format(T_FORMAT)
		row[1] = floatToString(chargeFull)
		row[2] = floatToString(chargeNow)
		row[3] = floatToString(currentNow)
		row[4] = strconv.Itoa(capacity)
		row[5] = floatToString(voltageNow)
		row[6] = floatToString(wattage)
		row[7] = status
		if err := addrow(path.Join(logdir, now.Format("2006-01-02") + ".csv"), row); err != nil {
			fmt.Printf("Warn: %v\n", err)
		}
		time.Sleep(duration * time.Millisecond)
	}
}
