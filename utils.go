package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	promptui "github.com/manifoldco/promptui"
)

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

func readCfgFile() generalConfiguration {
	cfgFile, err := os.Open(configFile)
	if err != nil {
		fmt.Printf("Failed to read configuration file:%v\nRunning first time initialization ...\n", err)
		if populateFS() {
			log.Fatalf("Failed to create files & folders!\n")
		} else {
			fmt.Printf("First time initialization completed!\n")
		}
	}
	defer cfgFile.Close()
	byteValue, fsErr := ioutil.ReadFile(configFile)
	if fsErr != nil {
		log.Fatalf("Error:%v\n", fsErr)
	}
	var configuration generalConfiguration
	json.Unmarshal(byteValue, &configuration)
	return configuration
}

func populateFS() bool {
	err := os.MkdirAll(dbDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Error:%v\n", err)
		return true
	}
	err = os.MkdirAll(htmlDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Error:%v\n", err)
		return true
	}
	err = os.MkdirAll(miscDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Error:%v\n", err)
		return true
	}
	cfgF, err := os.Create(configFile)
	if err != nil {
		fmt.Printf("Error:%v\n", err)
		return true
	}
	defer cfgF.Close()
	var defConfig generalConfiguration
	switch runtime.GOOS {
	case "linux":
		defConfig.SerialPort = "/dev/ttyUSB0"
	case "windows":
		defConfig.SerialPort = "COM4"
	default:
		defConfig.SerialPort = "none"
	}
	defConfig.RemoteHost = "no-such.server.info"
	defConfig.RemotePort = "8080"
	defConfig.RemoteUser = "defUser"
	defConfig.RemotePassword = "defPassword"
	err = writeCfgFile(defConfig)
	if err != nil {
		fmt.Printf("Error:%v\n", err)
		return true
	}
	profF, err := os.Create(batteryProfiles)
	if err != nil {
		fmt.Printf("Error:%v\n", err)
		return true
	}
	defer profF.Close()
	var demoProfile batteryProfile
	demoProfile.AssociatedDeviceName = "Demo Device"
	demoProfile.AssociateDevSnPrefix = "1234."
	demoProfile.MaxCycles = 200
	demoProfile.MinCapacityFactor = 0.75
	demoProfile.WarnCapacityFactor = 0.8
	demoProfile.WarnCycles = 178
	demoProfile.ImageFileBattery = "demobat.png"
	demoProfile.ImageFileDevice = "demodev.png"
	byteWriter, err := json.Marshal(demoProfile)
	if err != nil {
		fmt.Printf("Error:%v", err)
		return true
	}
	fsErr := ioutil.WriteFile(batteryProfiles, byteWriter, fs.ModeAppend)
	if fsErr != nil {
		fmt.Printf("Error:%v", err)
		return true
	}
	return false
}

func writeCfgFile(config generalConfiguration) error {
	byteWriter, err := json.Marshal(config)
	if err != nil {
		fmt.Printf("Error:%v", err)
		return err
	}
	fsErr := ioutil.WriteFile(configFile, byteWriter, fs.ModePerm)
	return fsErr
}

func removeEmptyStrings(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func readBatteryProfile(batSerial string) batteryProfile {
	profFile, err := os.Open(batteryProfiles)
	if err != nil {
		log.Fatalf("Failed to open \"%s\":%v\n", batteryProfiles, err)
	}

	var emptyProfile batteryProfile = batteryProfile{
		AssociatedDeviceName: "",
		AssociateDevSnPrefix: "",
		MaxCycles:            0,
		MinCapacityFactor:    0.0,
		WarnCycles:           0,
		WarnCapacityFactor:   0.0,
		ImageFileDevice:      "",
		ImageFileBattery:     "",
	}

	defer profFile.Close()
	byteValue, fsErr := ioutil.ReadAll(profFile)
	if fsErr != nil {
		fmt.Printf("Error:%v\n", err)
		return emptyProfile
	}
	var profiles []batteryProfile
	err = json.Unmarshal(byteValue, &profiles)
	if err != nil {
		fmt.Printf("Error:%v\n", err)
		return emptyProfile
	}
	for i := range profiles {
		if strings.HasPrefix(batSerial, profiles[i].AssociateDevSnPrefix) {
			fmt.Printf("Device profile matched! Sn:%s belongs to %s\n", batSerial, profiles[i].AssociatedDeviceName)
			return profiles[i]
		}
	}
	fmt.Printf("No profile found for prefix: %s", batSerial)
	return emptyProfile
}

func demoBat(DevSN string) rrcBatteryData {
	var demoBat rrcBatteryData
	demoBat.Manufacturer = "RND"
	demoBat.Name = "RND 1420"
	demoBat.Chemistry = "LION"
	demoBat.Specification = "ID3.1 Vs0 IPs0"
	if DevSN == "" {
		demoBat.DevSerialNumber = fmt.Sprintf("1234.%d", rand.Int())
	} else {
		demoBat.DevSerialNumber = DevSN
	}
	demoBat.MfgDate = fmt.Sprintf("%d / %d / %d", (time.Now().Year() - 2), time.Now().Month(), time.Now().Day())
	demoBat.Voltage = 11155
	demoBat.VoltageMeasured = 11202
	demoBat.Current = -21
	demoBat.TemperatureK = 305.3
	demoBat.TemperatureC = 32
	demoBat.NTC = 275
	demoBat.ChargingVoltage = 12600
	demoBat.ChargingCurrent = 4830
	demoBat.RelativeCharge = 45
	demoBat.RemainingCapacity = 5900
	demoBat.FullCapacity = 6990
	demoBat.AbsoluteCharge = 44
	demoBat.DesignCapacity = 7200
	demoBat.DesignVoltage = 10800
	demoBat.StateRegister = "0080 hex"
	demoBat.ModeRegister = "0001 hex"
	demoBat.CycleCount = 0
	demoBat.MaxError = 1
	demoBat.TimeAlarm = 10
	demoBat.TimeToFull = 65535
	demoBat.TimeToEmpty = 65535
	demoBat.CapacityAlarm = 690
	demoBat.BatteryUsesPEC = "Yes"
	demoBat.OptMfg2f = "0014 hex"
	demoBat.OptMfg3c = "0000 hex"
	demoBat.OptMfg3d = "0e85 hex"
	demoBat.OptMfg3e = "0e86 hex"
	demoBat.OptMfg3f = "0e87 hex"
	demoBat.DevSerialNumber = fmt.Sprintf("d_%s", DevSN)
	return demoBat
}

func platformSpecifics() (string, string) {
	switch runtime.GOOS {
	case "linux":
		return "\n", "Linux"
	case "windows":
		return "\r\n", "Windows"
	default:
		return "\n", "Unknown"
	}
}

func promptMainMenu(menulabel string) string {
	prompt := promptui.Select{
		Label: menulabel,
		Items: []string{"Read battery", "Serial config", "Read-only", "Demo-mode", "Cancel"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}
	return result
}
