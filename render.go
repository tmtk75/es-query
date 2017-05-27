package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	ui "github.com/gizak/termui"
)

type ESRes struct {
	Hits struct {
		Total int `json:"total"`
		Hits  []struct {
			Id      string `json:"_id"`
			_Source struct {
				Ipaddr string `jsons:"ipaddr"`
				Path   string `jsons:"path"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations struct {
		Tps struct {
			Buckets []struct {
				KeyStr   string `json:"key_as_string"`
				Key      int64  `json:"key"`
				DocCount int    `json:"doc_count"`
			} `json:"buckets"`
		} `json:"tps"`
	} `json:"aggregations"`
}

type UIOpts struct {
	Width int
}

func (opt *UIOpts) Render(rs, qstr string, start, end time.Time) {
	var res ESRes
	err := json.Unmarshal([]byte(rs), &res)
	if err != nil {
		log.Fatalln(err)
	}

	buckets := res.Aggregations.Tps.Buckets
	if len(buckets) == 0 {
		log.Printf("Quit because TPS is empty.")
		return
	}

	err = ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	tps := (func() []float64 {
		ps := make([]float64, len(buckets))
		for i, b := range buckets {
			ps[i] = float64(b.DocCount)
		}
		return ps
	})()

	leftMargin := 1

	qinfo := ui.NewPar("" +
		"Query: " + qstr + "\n" +
		"Start: " + start.String() + "\n" +
		"End: " + end.String() + "")
	qinfo.Height = 3
	qinfo.Width = opt.Width
	qinfo.X = leftMargin
	qinfo.Y = 0
	qinfo.Border = false

	max := tps[0]
	min := tps[0]
	for _, v := range tps {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}

	minfo := ui.NewPar("" +
		"Hits: " + fmt.Sprintf("%d", res.Hits.Total) + "\n" +
		"Max: " + fmt.Sprintf("%.0f", max) + "\n" +
		"Min: " + fmt.Sprintf("%.0f", min) + "")
	minfo.Height = 3
	minfo.Width = opt.Width
	minfo.X = leftMargin
	minfo.Y = qinfo.Y + qinfo.Height
	minfo.Border = false

	chart := ui.NewLineChart()
	chart.BorderLabel = "TPS"
	chart.Data = tps
	chart.Width = opt.Width
	chart.Height = int(float32(opt.Width)/1.4142) / 3
	chart.X = 1
	chart.Y = minfo.Y + minfo.Height
	chart.AxesColor = ui.ColorWhite
	chart.LineColor = ui.ColorGreen | ui.AttrBold

	ui.Render(qinfo, minfo, chart)

	// event handler...
	stop := func(ui.Event) { ui.StopLoop() }
	ui.Handle("/sys/kbd/q", stop)
	ui.Handle("/sys/kbd/C-c", stop)
	ui.Loop()
}
