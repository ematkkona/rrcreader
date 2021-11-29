// PDX-License-Identifier: MIT
// RRC SMBus reader
// 2021 eetu@kkona.xyz
// https://github.com/ematkkona/rrcreader

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	promptui "github.com/manifoldco/promptui"
	serial "github.com/tarm/serial"
)

const fmtDateTime string = "20060102150405"

type rrcBatteryData struct {
	Manufacturer      string  `json:"manufacturer"`      // "RRC"
	Name              string  `json:"name"`              // "RRC2020"
	Chemistry         string  `json:"chemistry"`         // "LION"
	Specification     string  `json:"specification"`     // "ID3.1 Vs0 IPs0"
	SerialNumber      string  `json:"serial"`            // "#0000"
	MfgDate           string  `json:"mfgdate"`           // "YEAR / MONTH / DAY"
	Voltage           int     `json:"voltage"`           // "00000 mV"
	VoltageMeasured   int     `json:"voltagemeasured"`   // "00000 mV"
	Current           int     `json:"current"`           // "-00 mA"
	TemperatureK      float64 `json:"kelvin"`            // "00.0 K"
	TemperatureC      float64 `json:"celsius"`           // "00.0 C"
	NTC               int     `json:"ntc"`               // "000 ohm"
	ChargingVoltage   int     `json:"chargingvoltage"`   // 00000 mV
	ChargingCurrent   int     `json:"chargingcurrent"`   // 0000 mA
	RelativeCharge    int     `json:"relativecharge"`    // "00 %"
	RemainingCapacity int     `json:"remainingcapacity"` // "0000 mAh"
	FullCapacity      int     `json:"fullcapacity"`      // "0000 mAh"
	AbsoluteCharge    int     `json:"absolutecharge"`    // "00 %"
	DesignCapacity    int     `json:"designcapacity"`    // "0000 mAh"
	DesignVoltage     int     `json:"designvoltage"`     // "00000 mV"
	StateRegister     string  `json:"stateregister"`     // "00e0 hex"
	ModeRegister      string  `json:"moderegister"`      // "0001 hex"
	CycleCount        int     `json:"cyclecount"`        // "#0"
	MaxError          int     `json:"maxerror"`          // "1 %"
	TimeAlarm         int     `json:"timealarm"`         // "10 min"
	TimeToFull        int     `json:"timetofull"`        // "0 min"
	TimeToEmpty       int     `json:"timetoempty"`       // "00000 min"
	CapacityAlarm     int     `json:"capacityalarm"`     // "000 mAh"
	BatteryUsesPEC    string  `json:"batteryusespec"`    // "Yes"
	OptMfg2f          string  `json:"optmfg2f"`          // "000a hex"
	OptMfg3c          string  `json:"optmfg3c"`          // "0000 hex"
	OptMfg3d          string  `json:"optmfg3d"`          // "0fdc hex
	OptMfg3e          string  `json:"optmfg3e"`          // "0fd1 hex"
	OptMfg3f          string  `json:"optmfg3f"`          // "0fde hex"
	DevSerialNumber   string  `json:"devserialnumber"`   // device under test sn
	Timestamp         string  `json:"timestamp"`         // current time
}

func launchViewer(target string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", target).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", target).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func parseTemps(raw string) (float64, float64, string) {
	var tempK, tempC float64
	var err error
	var rerr string
	tempsstr := strings.Split(raw, "K")
	tempsstr[0] = strings.TrimSpace(tempsstr[0])
	tempsstr[1] = strings.TrimSpace(stripValues(tempsstr[1]))
	tempK, err = strconv.ParseFloat(tempsstr[0], 64)
	if err != nil {
		rerr = "[parserr:K]"
	}
	tempC, err = strconv.ParseFloat(tempsstr[1], 64)
	if err != nil {
		rerr = rerr + "[parserr:C]"
	}
	return tempK, tempC, rerr
}

func promptMainMenu(menulabel string) string {
	prompt := promptui.Select{
		Label: menulabel,
		Items: []string{"Serial config", "Attach device SN", "Read data", "Cancel"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}
	return result
}

func stripValues(in string) string {
	reg, _ := regexp.Compile("[^0-9 . -]+")
	return reg.ReplaceAllString(in, "")
}

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		data[len(data)-1] = '\n'
		return data[0 : len(data)-1]
	}
	return data
}

func ScanCR(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte{'\r'}); i >= 0 {
		return i + 1, dropCR(data[0:i]), nil
	}
	if atEOF {
		return len(data), dropCR(data), nil
	}
	return 0, nil, nil
}

func clearScreen() {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func main() {

	config := &serial.Config{
		Baud:        9600,
		ReadTimeout: time.Millisecond * 30000,
	}
	replaceInputStr := "\n"
	platformName := "Linux"
	switch runtime.GOOS {
	case "linux":
		config.Name = "/dev/ttyUSB0"
	case "windows":
		platformName = "Windows"
		config.Name = "COM4"
		replaceInputStr = "\r\n"
	default:
		log.Fatal("unsupported platform")
	}
	proceedCondition := false
	DevSerialNumber := ""
	DevSNFMT := "<none>"
	for {
		if DevSerialNumber == "" {
			DevSNFMT = "<none>"
		} else {
			DevSNFMT = DevSerialNumber
		}
		clearScreen()
		menulabel := fmt.Sprintf("Platform:%s  Port:%s  Device SN:%s", platformName, config.Name, DevSNFMT)
		switch promptMainMenu(menulabel) {
		case "Serial config":
			time.Sleep(time.Millisecond * 100)
			fmt.Printf("Enter port [%s]:> ", config.Name)
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			text = strings.Replace(text, replaceInputStr, "", -1)
			if text != "" {
				config.Name = text
			}
		case "Attach device SN":
			time.Sleep(time.Millisecond * 100)
			fmt.Printf("Attach battery to device (%s):> ", DevSerialNumber)
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			text = strings.Replace(text, replaceInputStr, "", -1)
			if text != "" {
				DevSerialNumber = text
			}
		case "Read data":
			time.Sleep(time.Millisecond * 100)
			fmt.Printf("Waiting for data (%s) ... ", config.Name)
			proceedCondition = true
		default:
			os.Exit(0)
		}
		if proceedCondition {
			break
		}
	}
	const startendLine string = "-----------------------------------"
	stream, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(stream)
	const maxCapacity = 1130
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	scanner.Split(ScanCR)
	scannerState := false
	scannedLine := ""
	var unknownFields []string
	thisBattery := new(rrcBatteryData)
	for scanner.Scan() {
		scannedLine = scanner.Text()
		if scannerState {
			if scannedLine != startendLine {
				splitted := strings.Split(scannedLine, ":")
				splitted[0] = strings.TrimSpace(splitted[0])
				splitted[1] = strings.TrimSpace(splitted[1])
				switch splitted[0] {
				case "MANUFACTURER":
					thisBattery.Manufacturer = splitted[1]
				case "BATTERY NAME":
					thisBattery.Name = splitted[1]
				case "CHEMISTRY":
					thisBattery.Chemistry = splitted[1]
				case "SPECIFICATION":
					thisBattery.Specification = splitted[1]
				case "SERIAL NUMBER":
					thisBattery.SerialNumber = splitted[1]
				case "MANUFACT. DATE":
					thisBattery.MfgDate = splitted[1]
				case "VOLTAGE":
					thisBattery.Voltage, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error(voltage): %v", err)
					}
				case "VOLTAGE MEASURED":
					thisBattery.VoltageMeasured, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error(voltmeasured): %v", err)
					}
				case "CURRENT":
					thisBattery.Current, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error(current): %v", err)
					}
				case "TEMPERATURE":
					tempK, tempC, err := parseTemps(splitted[1])
					if err != "" {
						fmt.Printf("Error(s) encountered: parseTemps(%s) %v\n", splitted[1], err)
						thisBattery.TemperatureK = 0.0
						thisBattery.TemperatureC = 0.0
					} else {
						thisBattery.TemperatureK = tempK
						thisBattery.TemperatureC = tempC
					}
				case "NTC MEASURED":
					thisBattery.NTC, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error: %v", err)
					}
				case "RELATIVE CHARGE":
					thisBattery.RelativeCharge, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error: %v", err)
					}
				case "ABSOLUTE CHARGE":
					thisBattery.AbsoluteCharge, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error: %v", err)
					}
				case "DESIGN CAPACITY":
					thisBattery.DesignCapacity, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error: %v", err)
					}
				case "DESIGN VOLTAGE":
					thisBattery.DesignVoltage, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error: %v", err)
					}
				case "REMAIN. CAPACITY":
					thisBattery.RemainingCapacity, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error: %v", err)
					}
				case "FULL CAPACITY":
					thisBattery.FullCapacity, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error: %v", err)
					}
				case "CHARGING VOLTAGE":
					thisBattery.ChargingVoltage, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error: %v", err)
					}
				case "CHARGING CURRENT":
					thisBattery.ChargingCurrent, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error: %v", err)
					}
				case "TIME TO EMPTY":
					thisBattery.TimeToEmpty, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error: %v", err)
					}
				case "TIME TO FULL":
					thisBattery.TimeToFull, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error: %v", err)
					}
				case "CAPACITY ALARM":
					thisBattery.CapacityAlarm, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error: %v", err)
					}
				case "TIME ALARM":
					thisBattery.TimeAlarm, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error: %v", err)
					}
				case "CYCLE COUNT":
					thisBattery.CycleCount, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error: %v", err)
					}
				case "MAX ERROR":
					thisBattery.MaxError, err = strconv.Atoi(strings.TrimSpace(stripValues(splitted[1])))
					if err != nil {
						fmt.Printf("Conversion error: %v", err)
					}
				case "STATE REGISTER":
					thisBattery.StateRegister = splitted[1]
				case "MODE REGISTER":
					thisBattery.ModeRegister = splitted[1]
				case "OptMfg 0x2f":
					thisBattery.OptMfg2f = splitted[1]
				case "OptMfg 0x3c":
					thisBattery.OptMfg3c = splitted[1]
				case "OptMfg 0x3d":
					thisBattery.OptMfg3d = splitted[1]
				case "OptMfg 0x3e":
					thisBattery.OptMfg3e = splitted[1]
				case "OptMfg 0x3f":
					thisBattery.OptMfg3f = splitted[1]
				case "BATTERY USES PEC":
					thisBattery.BatteryUsesPEC = splitted[1]
				default:
					unknownFields = append(unknownFields, fmt.Sprintf("\"Unspecified: %s (= %s)\"", splitted[0], splitted[1]))
				}
			}
		}
		if scannedLine == startendLine {
			scannerState = !scannerState
			if !scannerState {
				stream.Flush()
				stream.Close()
				break
			} else {
				fmt.Printf("OK!\n")
			}
		}
	}
	tStamp := time.Now()
	thisBattery.Timestamp = tStamp.Format(fmtDateTime)
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	thisBattery.DevSerialNumber = DevSerialNumber
	if len(unknownFields) != 0 {
		fmt.Println("Warning! Following entries were discarded (unknown data):")
		fmt.Printf("%s\n", unknownFields)
	}
	u, err := json.Marshal(thisBattery)
	if err != nil {
		panic(err)
	}
	saveAs := fmt.Sprintf("%sT%v-%s", thisBattery.DevSerialNumber, tStamp.Format(fmtDateTime), stripValues(thisBattery.SerialNumber))
	fmt.Printf("Data from \"%s %s\" extracted successfully\nSaving readout to \"%s.json\"\n", thisBattery.Name, thisBattery.SerialNumber, saveAs)
	err = os.WriteFile(saveAs+".json", u, 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Launching viewer for \"%s.json\" ...\n", saveAs)
	launchViewer(saveAs + ".json")
	fmt.Printf("All done!\n")
	os.Exit(0)
}
