// PDX-License-Identifier: MIT
// RRC SMBus reader
// 2021 eetu@kkona.xyz
// https://github.com/ematkkona/rrcreader

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	serial "github.com/tarm/serial"
)

func main() {

	config := &serial.Config{
		Baud:        9600,
		ReadTimeout: time.Millisecond * 30000,
	}
	replaceInputStr, platformName := platformSpecifics()
	genConfig := readCfgFile()
	config.Name = genConfig.SerialPort
	proceedCondition := false
	DevSNFMT := "(none)"
	demoData := false
	omitWrites := false
	for {
		clearScreen()
		menulabel := fmt.Sprintf("OS:\"%s\" Serial port:\"%s\"", platformName, config.Name)
		if demoData {
			menulabel = fmt.Sprintf("OS:\"%s\" *Demo-mode", platformName)
		}
		if omitWrites {
			menulabel = fmt.Sprintf("OS:\"%s\" Serial port:\"%s\" *Omit write-operations", platformName, config.Name)
		}
		switch promptMainMenu(menulabel) {
		case "Read battery":
			time.Sleep(time.Millisecond * 100)
			fmt.Printf("Waiting for data (%s) ... ", config.Name)
			proceedCondition = true
		case "Serial config":
			time.Sleep(time.Millisecond * 100)
			fmt.Printf("Enter port [%s]:> ", config.Name)
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			text = strings.Replace(text, replaceInputStr, "", -1)
			if text != "" {
				config.Name = text
			}
		case "Read-only":
			time.Sleep(time.Millisecond * 100)
			demoData = false
			omitWrites = !omitWrites
		case "Demo-mode":
			time.Sleep(time.Millisecond * 100)
			demoData = !demoData
		default:
			os.Exit(0)
		}
		if proceedCondition {
			break
		}
	}
	const startendLine string = "-----------------------------------"
	thisBattery := new(rrcBatteryData)
	if !demoData {
		stream, err := serial.OpenPort(config)
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(stream)
		buf := make([]byte, maxRx)
		scanner.Buffer(buf, maxRx)
		scanner.Split(ScanCR)
		scannerState := false
		scannedLine := ""
		var unknownFields []string
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
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		if len(unknownFields) != 0 {
			fmt.Println("Warning! Following entries were discarded (unknown data):")
			fmt.Printf("%s\n", unknownFields)
		}
	} else {
		*thisBattery = demoBat(DevSNFMT)
	}
	var retData []rrcBatteryData
	var retCode int

	retData, retCode = dbhandler("check", dbDir, *thisBattery)
	if retCode != 0 {
		fmt.Printf("New battery? Attach to device :>")
		time.Sleep(time.Millisecond * 100)
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, replaceInputStr, "", -1)
		if text != "" {
			thisBattery.DevSerialNumber = text
		} else {
			thisBattery.DevSerialNumber = ""
		}
	} else {
		for i := range retData {
			thisBattery.DevSerialNumber = retData[i].DevSerialNumber
		}
		fmt.Printf("Battery identified! Associated device sn:\"%s\"\n", thisBattery.DevSerialNumber)
	}
	tStamp := time.Now()
	thisBattery.Timestamp = tStamp.Format(fmtDateTime)

	if !omitWrites {
		_, retCode = dbhandler("write", dbDir, *thisBattery)
		if retCode != 0 {
			fmt.Printf("dbhandler(write>%s) returned: %d\n", dbDir, retCode)
		}
	}
	var dbArgData rrcBatteryData
	dbArgData.Name = thisBattery.Name
	dbArgData.SerialNumber = thisBattery.SerialNumber
	dbArgData.DevSerialNumber = DevSNFMT

	retData, retCode = dbhandler("read", dbDir, dbArgData)
	recEntryAmt := 0
	if retCode == 0 {
		for _, f := range retData {
			if f.DevSerialNumber != thisBattery.DevSerialNumber {
				fmt.Printf("Warning! Record %d: \"%s-%sT%s\" device association mismatch!\nGot: \"%s\". Expected: \"%s\".\n", recEntryAmt, f.Name, f.SerialNumber, f.Timestamp, f.DevSerialNumber, thisBattery.DevSerialNumber)
			}
			recEntryAmt++
		}
	}
	if retCode != 0 {
		fmt.Printf("dbhandler(read>%s) returned: %d\n", dbDir, retCode)
	}

	/*saveAs := fmt.Sprintf("./data/%sT%v-%s", thisBattery.DevSerialNumber, tStamp.Format(fmtDateTime), stripValues(thisBattery.SerialNumber))
	fmt.Printf("Data from \"%s %s\" extracted successfully\nSaving readout to \"%s.json/html\"\n", thisBattery.Name, thisBattery.SerialNumber, saveAs)
	err = os.WriteFile(saveAs+".json", u, 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v", err)
		os.Exit(1)
	}*/
	generateGraphs(*thisBattery)
	err := writeCfgFile(genConfig)
	if err != nil {
		fmt.Printf("Error writing\"%s\":%v\n", configFile, err)
	}
	fmt.Printf("All done!\n")
	os.Exit(0)
}
