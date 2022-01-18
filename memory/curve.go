package main

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/spf13/cobra"
)

func Str2Uint(s string) uint64 {
	n, _ := strconv.ParseUint(s, 10, 64)

	return n
}

func GeneratelineItems(idx int, data [][]string) []opts.LineData {
	items := make([]opts.LineData, len(data))

	for i := range data {
		items[i] = opts.LineData{Value: Str2Uint(data[i][idx])}
	}

	return items
}

func GeneratelineItems2(idx int, data [][]string) []opts.LineData {
	items := make([]opts.LineData, len(data))

	for i := range data {
		items[i] = opts.LineData{Value: []interface{}{data[i][0], Str2Uint(data[i][idx])}}
	}

	return items
}

func GeneratelineXItems(idx int, data [][]string) []opts.LineData {
	items := make([]opts.LineData, len(data))

	for i := range data {
		items[i] = opts.LineData{Value: data[i][idx]}
	}

	return items
}

// // 无法展示x轴（时间）
// func PrintCurve(cmd *cobra.Command, args []string) error {
// 	data, err := LoadCSV()
// 	CheckErr(err)

// 	if len(data) <= 1 {
// 		log.Println("no data")
// 	}

// 	heads := data[0]
// 	data = data[1:]

// 	line := charts.NewLine()
// 	// set some global options like Title/Legend/ToolTip or anything else
// 	line.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
// 		Title: "memory",
// 	}),
// 		charts.WithLegendOpts(opts.Legend{
// 			Show: true,
// 			Data: heads[1:],
// 		}),
// 		charts.WithTooltipOpts(opts.Tooltip{
// 			Show:    true,
// 			Trigger: "axis",
// 		}),
// 	)

// 	// Put data into instance
// 	line.SetXAxis(GeneratelineXItems(0, data))
// 	for i, v := range heads[1:] {
// 		line.AddSeries(v, GeneratelineItems(i+1, data))
// 	}

// 	// Where the magic happens
// 	f, _ := os.Create("curve.html")
// 	line.Render(f)

// 	return nil
// }

func PrintCurve2(cmd *cobra.Command, args []string) error {
	data, err := LoadCSV()
	CheckErr(err)

	if len(data) <= 1 {
		log.Println("no data")
	}

	heads := data[0]
	data = data[1:]

	line := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title: "memory",
	}),
		charts.WithLegendOpts(opts.Legend{
			Show: true,
			Data: heads[1:],
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    true,
			Trigger: "axis",
		}),
		charts.WithDataZoomOpts(opts.DataZoom{ // zoom
			Type: "inside",
		}),
	)

	line.XYAxis.XAxisList = []opts.XAxis{
		{
			Show: true,
			Type: "time",
		},
	}
	line.XYAxis.YAxisList = []opts.YAxis{
		{
			Show: true,
			Type: "value",
		},
	}
	for i, v := range heads[1:] {
		line.AddSeries(v, GeneratelineItems2(i+1, data))
	}

	// Where the magic happens
	f, _ := os.Create("curve.html")
	line.Render(f)

	return nil
}

func LoadCSV() ([][]string, error) {
	f, err := os.Open(OutputFile)
	CheckErr(err)
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ','

	return r.ReadAll()
}
