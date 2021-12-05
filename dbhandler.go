package main

import (
	"encoding/json"
	"fmt"

	scribble "github.com/nanobox-io/golang-scribble"
)

func dbhandler(oper string, dbdir string, datasetin rrcBatteryData) ([]rrcBatteryData, int) {
	noData := []rrcBatteryData{}
	db, err := scribble.New(dbdir, nil)
	if err != nil {
		fmt.Println("Error", err)
		return noData, 1
	}
	identifier := datasetin.Name + datasetin.SerialNumber
	switch oper {
	case "read":
		return read(identifier, *db)
	case "write":
		err := write(datasetin, *db)
		return noData, err
	case "check":
		if devSN, err := check(identifier, *db); err == 0 {
			datasetin.DevSerialNumber = devSN
			noData = append(noData, datasetin)
			return noData, 0
		} else {
			return noData, 1
		}
	case "sync":
	default:
		fmt.Println("Error: unknown argument")
	}
	return noData, 1
}

func read(identifier string, db scribble.Driver) ([]rrcBatteryData, int) {
	records, err := db.ReadAll(identifier)
	if err != nil {
		fmt.Printf("Database read error: %v\n", err)
		var noData []rrcBatteryData
		return noData, 1
	}
	recordslist := []rrcBatteryData{}
	for _, f := range records {
		recFound := rrcBatteryData{}
		if err := json.Unmarshal([]byte(f), &recFound); err != nil {
			fmt.Printf("Database read error: %v\n", err)
			var noData []rrcBatteryData
			return noData, 1
		}
		recordslist = append(recordslist, recFound)
	}
	return recordslist, 0
}

func write(dataset rrcBatteryData, db scribble.Driver) int {
	err := db.Write(dataset.Name+dataset.SerialNumber, dataset.Timestamp, dataset)
	if err != nil {
		fmt.Printf("Database write error: %v\n", err)
		return 1
	}
	return 0
}

func check(identifier string, db scribble.Driver) (string, int) {
	devSN := ""
	records, err := db.ReadAll(identifier)
	if err != nil {
		return devSN, 1
	}
	for _, f := range records {
		recFound := rrcBatteryData{}
		if err := json.Unmarshal([]byte(f), &recFound); err != nil {
			fmt.Printf("Database read error: %v\n", err)
			return devSN, 1
		} else {
			if recFound.DevSerialNumber != "" && recFound.DevSerialNumber != "(none)" {
				devSN = recFound.DevSerialNumber
			}
		}
	}
	return devSN, 0
}
