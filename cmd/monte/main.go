package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/srliao/gisim/pkg/combat"
	"github.com/srliao/gisim/pkg/monte"
	"gopkg.in/yaml.v2"
)

func main() {
	var source []byte
	var cfg combat.Profile
	var err error

	t := flag.Int64("t", 1000000, "how many iterations default 1mil")
	prf := flag.String("p", "config.yaml", "which profile to use; default config.yaml")
	index := flag.Int("i", 0, "character index to sim; default 0")
	worker := flag.Int64("w", 24, "number of works, default 24")
	bin := flag.Int64("b", 100, "bin size, default 100")
	out := flag.String("o", "out.html", "output file; default out.html")
	flag.Parse()

	source, err = ioutil.ReadFile(*prf)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(source, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	sim, err := monte.New(cfg, *index)
	r := sim.SimDmgDist(*t, *bin, *worker)
	elapsed := time.Since(start)
	fmt.Printf("Provie %v done in %s\n", *prf, elapsed)

	page := components.NewPage()
	page.PageTitle = "simulation results"

	var bins []int64
	var items []opts.LineData
	var cumul, med float64
	med = -1
	binSize := *bin

	for i, v := range r.Hist {
		bins = append(bins, r.BinStart+binSize*int64(i))
		items = append(items, opts.LineData{Value: v})
		cumul += v / float64(*t)
		if cumul >= 0.5 && med == -1 {
			med = float64(i)
		}
	}

	med = float64(r.BinStart) + med*float64(binSize)
	label := fmt.Sprintf("min: %v, max %v, mean: %.2f, med: %.2f, sd: %.2f", r.Min, r.Max, r.Mean, med, r.SD)

	lineChart := charts.NewLine()
	lineChart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: fmt.Sprintf("%v (n = %v)", *prf, *t),
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Freq",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "DPS",
		}),
		// charts.WithTooltipOpts(opts.Tooltip{Show: true}),
		charts.WithLegendOpts(opts.Legend{Show: true, Top: "5%", Right: "0%", Orient: "vertical", Data: []string{label}}),
	)
	lineChart.AddSeries(label, items)
	lineChart.SetXAxis(bins)

	page.AddCharts(
		lineChart,
	)

	graph, err := os.Create(*out)
	if err != nil {
		panic(err)
	}
	page.Render(io.MultiWriter(graph))

}
