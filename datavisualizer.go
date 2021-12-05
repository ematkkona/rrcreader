package main

import (
	"fmt"
	"os"
	"time"

	charts "github.com/go-echarts/go-echarts/v2/charts"
	components "github.com/go-echarts/go-echarts/v2/components"
	opts "github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

func generateVoltageGaugeItems(dataset rrcBatteryData) []opts.GaugeData {
	chartGItems := make([]opts.GaugeData, 0)
	//chartGItems = append(chartGItems, opts.GaugeData{Value: dataset.DesignVoltage})
	chartGItems = append(chartGItems, opts.GaugeData{Value: dataset.VoltageMeasured})
	//chartGItems = append(chartGItems, opts.GaugeData{Value: dataset.Voltage})
	//chartGItems = append(chartGItems, opts.GaugeData{Value: dataset.ChargingVoltage})
	return chartGItems
}
func generateRelChargeGaugeItem(dataset rrcBatteryData) []opts.GaugeData {
	chartRelChargeGauge := make([]opts.GaugeData, 0)
	chartRelChargeGauge = append(chartRelChargeGauge, opts.GaugeData{Value: dataset.RelativeCharge})
	return chartRelChargeGauge
}

func generateBarItems(dataset rrcBatteryData, dataswitch string) []opts.BarData {
	barItems := make([]opts.BarData, 0)
	switch dataswitch {
	case "Capacity":
		//chartItems = append(chartItems, opts.BarData{Value: dataset.DesignCapacity})
		barItems = append(barItems, opts.BarData{Value: dataset.FullCapacity})
		barItems = append(barItems, opts.BarData{Value: dataset.RemainingCapacity})
	case "Currents":
		barItems = append(barItems, opts.BarData{Value: dataset.Current})
		barItems = append(barItems, opts.BarData{Value: dataset.ChargingCurrent})
	case "BatInfo":
		barItems = append(barItems, opts.BarData{Value: dataset.Name})
		barItems = append(barItems, opts.BarData{Value: dataset.SerialNumber})
		barItems = append(barItems, opts.BarData{Value: dataset.Manufacturer})
		barItems = append(barItems, opts.BarData{Value: dataset.Chemistry})
		barItems = append(barItems, opts.BarData{Value: dataset.Specification})
	case "BatHealth":
		barItems = append(barItems, opts.BarData{Value: dataset.CycleCount})
		barItems = append(barItems, opts.BarData{Value: dataset.MfgDate})
		barItems = append(barItems, opts.BarData{Value: dataset.AbsoluteCharge})
	case "AttatchedDevice":
		barItems = append(barItems, opts.BarData{Value: dataset.DevSerialNumber})
	}
	return barItems
}

func generateLineChart(dataset []rrcBatteryData, profile batteryProfile) *charts.Line {
	line := charts.NewLine()
	capacity := make([]opts.LineData, 0)
	cycles := make([]opts.LineData, 0)
	timestamps := make([]string, 0)
	for cnt := range dataset {
		capacity = append(capacity, opts.LineData{Value: dataset[cnt].FullCapacity, Name: fmt.Sprintf("%v", cnt), YAxisIndex: 1})
		cycles = append(cycles, opts.LineData{Value: dataset[cnt].CycleCount, Name: fmt.Sprintf("%v", cnt)})
		tsISO, _ := time.Parse(fmtDateTime, dataset[cnt].Timestamp)
		timestamps = append(timestamps, tsISO.Format(fmtDateTimeISO))
	}

	timestamps = removeEmptyStrings(timestamps)
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeInfographic}),
		charts.WithTitleOpts(opts.Title{
			Title:    fmt.Sprintf("%s %s @ %s (sn:%s)", dataset[0].Name, dataset[0].SerialNumber, profile.AssociatedDeviceName, dataset[0].DevSerialNumber),
			Subtitle: fmt.Sprintf("%d measurement(s)", len(dataset)),
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name:        "Cycles",
			Type:        "value",
			Show:        true,
			SplitNumber: 0,
			Scale:       true,
			Min:         0,
			Max:         profile.MaxCycles + 50,
			SplitArea:   &opts.SplitArea{Show: false},
			SplitLine:   &opts.SplitLine{Show: true, LineStyle: &opts.LineStyle{Color: "#FF0202", Opacity: 1}},
			AxisLabel:   &opts.AxisLabel{},
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name:      "Date",
			Show:      true,
			Scale:     true,
			AxisLabel: &opts.AxisLabel{Show: true},
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:        true,
			Trigger:     "axis",
			TriggerOn:   "",
			Formatter:   "",
			AxisPointer: &opts.AxisPointer{Type: "cross", Snap: true},
		}),
	)
	line.ExtendYAxis(opts.YAxis{
		Name:        "Capacity",
		Type:        "value",
		Show:        true,
		SplitNumber: 0,
		Min:         int(profile.MinCapacityFactor*float64(dataset[0].DesignCapacity) - 200),
		Max:         dataset[0].DesignCapacity + 200,
		SplitArea:   &opts.SplitArea{Show: false},
		SplitLine:   &opts.SplitLine{Show: true, LineStyle: &opts.LineStyle{Color: "#FF0202", Opacity: 1}},
	})

	line.SetXAxis(timestamps).
		AddSeries("Cycles", cycles, charts.WithLineChartOpts(opts.LineChart{ConnectNulls: true, Stack: "Date"})).
		AddSeries("Capacity", capacity, charts.WithLineChartOpts(opts.LineChart{YAxisIndex: 1, ConnectNulls: true, Stack: "Date"})).
		SetSeriesOptions(charts.WithLabelOpts(opts.Label{Show: false}))

	return line
}

func generateGraphs(dataset rrcBatteryData) {

	BatteryProfile := readBatteryProfile(dataset.DevSerialNumber)
	batMaxCapacity := int(float64(dataset.DesignCapacity) * float64(1.2))
	batMaxVoltage := dataset.DesignVoltage
	batMinCapacity := int(float64(dataset.DesignCapacity) * BatteryProfile.MinCapacityFactor)
	//chargingVoltageString := fmt.Sprintf("Charging voltage: %d mV", dataset.ChargingVoltage)
	desCapacityString := fmt.Sprintf("Design capacity: %dmAh", dataset.DesignCapacity)
	remCapacityString := fmt.Sprintf("Remaining capacity: %dmAh", dataset.RemainingCapacity)
	actCurrentString := fmt.Sprintf("Current: %d mA", dataset.Current)
	chargingCurrentString := fmt.Sprintf("Charging current: %dmA", dataset.ChargingCurrent)
	//absoluteChargeString := fmt.Sprintf("Absolute charge: %d mA", dataset.AbsoluteCharge)
	fullCapacityString := fmt.Sprintf("Full capacity: %dmAh", dataset.FullCapacity)
	batMinVoltage := 0

	capbar := charts.NewBar()
	curbar := charts.NewBar()
	volgauge := charts.NewGauge()
	relcgauge := charts.NewGauge()

	capbar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Capacity",
		Subtitle: desCapacityString,
	}), charts.WithYAxisOpts(opts.YAxis{
		Type: "value",
		Min:  batMinCapacity,
		Max:  batMaxCapacity,
	}), charts.WithInitializationOpts(opts.Initialization{
		//BackgroundColor: "#010101",
		Width:  "600px",
		Height: "400px",
	}))

	capbar.SetSeriesOptions(charts.WithSunburstOpts(opts.SunburstChart{
		Animation: true,
	}), charts.WithBarChartOpts(opts.BarChart{
		BarGap:         "10%",
		BarCategoryGap: "20%",
	}), charts.WithLiquidChartOpts(opts.LiquidChart{
		IsWaveAnimation: true,
	}))
	capbar.SetXAxis([]string{fullCapacityString, remCapacityString}).
		AddSeries("Capacity", generateBarItems(dataset, "Capacity"), charts.WithItemStyleOpts(opts.ItemStyle{
			Color:   "#10F010",
			Color0:  "#F01010",
			Opacity: 0.6,
		}))

	curbar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Current",
		Subtitle: "mA",
	}), charts.WithYAxisOpts(opts.YAxis{
		Type: "value",
	}), charts.WithInitializationOpts(opts.Initialization{
		//BackgroundColor: "#010101",
		Width:  "600px",
		Height: "400px",
	}))

	curbar.SetSeriesOptions(charts.WithSunburstOpts(opts.SunburstChart{
		Animation: true,
	}), charts.WithBarChartOpts(opts.BarChart{
		BarGap:         "10%",
		BarCategoryGap: "20%",
	}), charts.WithLiquidChartOpts(opts.LiquidChart{
		IsWaveAnimation: true,
	}))
	curbar.SetXAxis([]string{actCurrentString, chargingCurrentString}).
		AddSeries("Current", generateBarItems(dataset, "Currents"), charts.WithItemStyleOpts(opts.ItemStyle{
			Color:   "#10F010",
			Color0:  "#F01010",
			Opacity: 0.4,
		}))
	volgauge.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Voltage",
		Subtitle: "mV",
	}), charts.WithSingleAxisOpts(opts.SingleAxis{
		Type: "value",
		Min:  batMinVoltage,
		Max:  batMaxVoltage,
	}), charts.WithRadiusAxisOps(opts.RadiusAxis{
		Inverse: true,
	}))

	relcgauge.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Relative Charge",
		Subtitle: "%",
	}), charts.WithSingleAxisOpts(opts.SingleAxis{
		Type: "value",
		Min:  0,
		Max:  100,
	}))

	volgauge.AddSeries("Voltage", generateVoltageGaugeItems(dataset), charts.WithItemStyleOpts(opts.ItemStyle{
		Color:        "#10F010",
		Color0:       "#F01010",
		BorderColor:  "#10AA10",
		BorderColor0: "#AA1010",
		Opacity:      0.4,
	}), charts.WithLabelOpts(opts.Label{
		Show:     true,
		Position: "insideTop",
	}))

	relcgauge.AddSeries("RelCharge", generateRelChargeGaugeItem(dataset), charts.WithItemStyleOpts(opts.ItemStyle{
		Color:        "#00FF00",
		Color0:       "#FF0000",
		BorderColor:  "#10AA10",
		BorderColor0: "#AA1010",
		Opacity:      0.4,
	}), charts.WithLabelOpts(opts.Label{
		Show:     true,
		Position: "insideTop",
	}))

	datasetAll, retCode := dbhandler("read", dbDir, dataset)
	if retCode != 0 {
		fmt.Printf("Error reading data for histogram!\n")
	}
	histogram := generateLineChart(datasetAll, BatteryProfile)
	relcgauge.Title.Left = "center"
	volgauge.Title.Left = "center"
	capbar.Title.Left = "center"
	curbar.Title.Left = "center"
	histogram.Title.Left = "center"

	page := components.NewPage()
	page.SetLayout(components.PageCenterLayout)
	page.PageTitle = fmt.Sprintf("Battery %s", dataset.Name+dataset.SerialNumber)
	//page.BackgroundColor = "#010101"
	//page.Theme = "white"
	page.AddCharts(histogram, relcgauge, volgauge, capbar, curbar)

	saveAs := fmt.Sprintf("%s/%s-%s.html", htmlDir, dataset.DevSerialNumber, stripValues(dataset.SerialNumber))
	f, _ := os.Create(saveAs)
	page.Render(f)
	launchViewer(saveAs)
}
