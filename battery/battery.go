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
		fmt.Println("err: ReadFile")
		// log.Fatal(err)
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
		fmt.Println("err: ReadFile")
		// log.Fatal(err)
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
	const pIntv = 10000 // print interval in ms
	const rIntv = 10000  // read interval in ms
	const pr = pIntv / rIntv  // p over r
	// durIntv := time.Duration(pIntv * 1000)
	durRIntv := time.Duration(rIntv)
	var now time.Time
	var status string
	var chargeFull float64 = 0
	var chargeNow float64 = 0
	var currentNow float64 = 0
	var capacity int = 0
	var capacitySum int = 0
	var chargeSum float64 = 0
	var currentSum float64 = 0
	var voltageSum float64 = 0
	var seconds float64
	var h int64
	var m int64
	var voltageNow float64 = 0
	var wattage float64 = 0
	// var strWattage string
	row := make([]string, 8)

	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("err: Can't find $HOME")
	}
	logdir := path.Join(homedir, ".cache", "battery")
	if _, err := os.Stat(logdir); os.IsNotExist(err) {
		os.Mkdir(logdir, 0755)
	}

	for {
		var count int = 0
		for count < pr {
			now = time.Now()
			status = parseString("status") // "Charging" | "Discharging"
			chargeFull = parseFloat("charge_full")
			chargeNow = parseFloat("charge_now")
			currentNow = parseFloat("current_now")
			voltageNow = parseFloat("voltage_now") / math.Pow10(3)
			capacity, err = strconv.Atoi(parseString("capacity"))
			if err != nil {
				log.Fatal(err)
			}

			// fmt.Printf("charge_full : %T %v\n", chargeFull, chargeFull)
			// fmt.Printf("charge_now  : %v\n", chargeNow)
			// fmt.Printf("current_now : %v\n", currentNow)

			if status == "Charging" {
				// wattage is not meaningful when Charging
				wattage = -1
			} else if status == "Discharging" {
				wattage = currentNow * voltageNow / math.Pow10(3)
			} else {
				// wattage is not meaningful when Full
				wattage = -1
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
			if err := addrow(path.Join(logdir, now.Format("2006-01-02")+".csv"), row); err != nil {
				fmt.Printf("Warn: %v\n", err)
			}
			chargeSum += chargeNow
			currentSum += currentNow
			voltageSum += voltageNow
			capacitySum += capacity
			count++
			time.Sleep(durRIntv * time.Millisecond)
		}
		chargeSum /= pr
		currentSum /= pr
		voltageSum /= pr
		capacitySum /= pr
		if status == "Charging" {
			seconds = 3600 * (chargeFull - chargeSum) / currentSum
			h, m = parseStoHM(seconds)
			if capacitySum >= 100 {
				if POLYBAR_COLOR {
					fmt.Printf("%%{F%s}%s%%{F-}\n", C_GOOD, "FULL")
				} else {
					fmt.Println("FULL")
				}
			} else {
				if POLYBAR_COLOR {
					fmt.Printf("%%{F%s}%v%% %%{F-}%02d:%02d\n", C_GOOD, capacitySum, h, m)
				} else {
					fmt.Printf("%v%% %02d:%02d\n", capacitySum, h, m)
				}
			}
		} else if status == "Discharging" {
			// seconds = 3600 * chargeSum / currentSum
			// h, m = parseStoHM(seconds)
			// wattage = currentSum * voltageSum / math.Pow10(3)
			if POLYBAR_COLOR {
				// if wattage > 10 {
				// 	strWattage = fmt.Sprintf("%%{F%s}%.1fW%%{F-}", C_CRITICAL, wattage)
				// } else if wattage > 8 {
				// 	strWattage = fmt.Sprintf("%%{F%s}%.1fW%%{F-}", C_WARNING, wattage)
				// } else if wattage > 6 {
				// 	strWattage = fmt.Sprintf("%%{F%s}%.1fW%%{F-}", C_CAUTION, wattage)
				// } else {
				// 	strWattage = fmt.Sprintf("%.1fW", wattage)
				// }
				if capacitySum < 10 {
					seconds = 3600 * chargeSum / currentSum
					h, m = parseStoHM(seconds)
					if h < 1 {
						fmt.Printf("%%{F%s}▮▯▯▯▯ %02d:%02d%%{F-}\n", C_CRITICAL, h, m)
					} else {
						fmt.Printf("%%{F%s}▮▯▯▯▯%%{F-}\n", C_CRITICAL)
					}
				} else if capacitySum < 20 {
					fmt.Printf("%%{F%s}▮▮▯▯▯%%{F-}\n", C_WARNING)
				} else if capacitySum < 40 {
					fmt.Printf("%%{F%s}▮▮▮▯▯%%{F-}\n", C_CAUTION)
				} else if capacitySum < 80 {
					fmt.Println("▮▮▮▮▯")
				} else {
					fmt.Println("▮▮▮▮▮")
				}
			} else {
				if capacitySum < 10 {
					fmt.Println("▮▯▯▯▯")
				} else if capacitySum < 20 {
					fmt.Println("▮▮▯▯▯")
				} else if capacitySum < 40 {
					fmt.Println("▮▮▮▯▯")
				} else if capacitySum < 80 {
					fmt.Println("▮▮▮▮▯")
				} else {
					fmt.Println("▮▮▮▮▮")
				}
			}
		} else {
			if POLYBAR_COLOR {
				fmt.Printf("%%{F%s}%s%%{F-}\n", C_GOOD, "FULL")
			} else {
				fmt.Println("FULL")
			}
		}
		chargeSum = 0
		currentSum = 0
		voltageSum = 0
		capacitySum = 0
		count = 0
		// time.Sleep(durIntv * time.Millisecond)
	}
}
