package main

const dbDir = "./data/db"
const htmlDir = "./data/html"
const miscDir = "./data/misc"
const configFile = "./data/GeneralConfiguration.json"
const batteryProfiles = "./data/BatteryProfiles.json"

const fmtDateTime string = "20060102150405"
const fmtDateTimeISO string = "2006-01-02"
const maxRx = 1130

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

type generalConfiguration struct {
	SerialPort     string `json:"serialport"`     // Serial port
	RemoteHost     string `json:"remotehost"`     // Remote host for syncing database
	RemotePort     string `json:"remoteport"`     // Port for syncing database
	RemoteUser     string `json:"remoteuser"`     // Username for remote access
	RemotePassword string `json:"remotepassword"` // Password for remote access
}

type batteryProfile struct {
	AssociatedDeviceName string  `json:"associateddevicename"` // Associated device name
	AssociateDevSnPrefix string  `json:"assosiatedevsnprefix"` // Associate devices with serial expected with prefix (1234.******)
	MaxCycles            int     `json:"maxcycles"`            // MAX cycles as defined by device manufacturer
	MinCapacityFactor    float64 `json:"mincapacityfactor"`    // MIN capacity as defined by device manufacturer
	WarnCycles           int     `json:"warncycles"`           // (optional) number of cycles to trigger yellow health status
	WarnCapacityFactor   float64 `json:"warncapacityfactor"`   // (optional) capacity level to trigger yellow heatlh status
	ImageFileDevice      string  `json:"imagefiledevice"`      // (optional) image file for the associated device
	ImageFileBattery     string  `json:"imagefilebattery"`     // (optional) image file for the battery
}
