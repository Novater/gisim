package main

import (

	//characters
	_ "github.com/srliao/gisim/internal/character/bennett"
	_ "github.com/srliao/gisim/internal/character/eula"
	_ "github.com/srliao/gisim/internal/character/fischl"
	_ "github.com/srliao/gisim/internal/character/ganyu"
	_ "github.com/srliao/gisim/internal/character/xiangling"
	_ "github.com/srliao/gisim/internal/character/xingqiu"

	//weapons
	_ "github.com/srliao/gisim/internal/weapon/bow/favoniuswarbow"
	_ "github.com/srliao/gisim/internal/weapon/bow/prototypecrescent"
	_ "github.com/srliao/gisim/internal/weapon/claymore/skyward"
	_ "github.com/srliao/gisim/internal/weapon/spear/blacktassel"
	_ "github.com/srliao/gisim/internal/weapon/spear/skywardspine"
	_ "github.com/srliao/gisim/internal/weapon/sword/festeringdesire"
	_ "github.com/srliao/gisim/internal/weapon/sword/sacrificialsword"

	//sets
	_ "github.com/srliao/gisim/internal/artifact/blizzard"
	_ "github.com/srliao/gisim/internal/artifact/bloodstained"
	_ "github.com/srliao/gisim/internal/artifact/crimson"
	_ "github.com/srliao/gisim/internal/artifact/gladiator"
	_ "github.com/srliao/gisim/internal/artifact/noblesse"
	_ "github.com/srliao/gisim/internal/artifact/wanderer"
)

func main() {
	// var source []byte
	// var cfg combat.Profile
	// var err error

	// t := flag.Int("t", 1000000, "how many iterations default 1mil")
	// prf := flag.String("p", "config.yaml", "which profile to use; default config.yaml")
	// index := flag.Int("i", 0, "character index to sim; default 0")
	// worker := flag.Int("w", 24, "number of works, default 24")
	// bin := flag.Int("b", 100, "bin size, default 100")
	// out := flag.String("o", "out.html", "output file; default out.html")
	// flag.Parse()

	// source, err = ioutil.ReadFile(*prf)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// err = yaml.Unmarshal(source, &cfg)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// start := time.Now()
	// sim, err := combat.NewArtifactSim(cfg, *index)
	// if err != nil {
	// 	panic(err)
	// }
	// r := sim.RunDmgSim(*bin, *t, *worker)
	// elapsed := time.Since(start)
	// fmt.Printf("Provie %v done in %s\n", *prf, elapsed)

	// page := components.NewPage()
	// page.PageTitle = "simulation results"

	// var bins []int
	// var items []opts.LineData
	// var cumul, med float64
	// med = -1
	// binSize := *bin

	// for i, v := range r.Hist {
	// 	bins = append(bins, r.BinStart+binSize*i)
	// 	items = append(items, opts.LineData{Value: v})
	// 	cumul += v / float64(*t)
	// 	if cumul >= 0.5 && med == -1 {
	// 		med = float64(i)
	// 	}
	// }

	// med = float64(r.BinStart) + med*float64(binSize)
	// label := fmt.Sprintf("min: %v, max %v, mean: %.2f, med: %.2f, sd: %.2f", r.Min, r.Max, r.Mean, med, r.SD)

	// lineChart := charts.NewLine()
	// lineChart.SetGlobalOptions(
	// 	charts.WithTitleOpts(opts.Title{
	// 		Title: fmt.Sprintf("%v (n = %v)", *prf, *t),
	// 	}),
	// 	charts.WithYAxisOpts(opts.YAxis{
	// 		Name: "Freq",
	// 	}),
	// 	charts.WithXAxisOpts(opts.XAxis{
	// 		Name: "DPS",
	// 	}),
	// 	// charts.WithTooltipOpts(opts.Tooltip{Show: true}),
	// 	charts.WithLegendOpts(opts.Legend{Show: true, Top: "5%", Right: "0%", Orient: "vertical", Data: []string{label}}),
	// )
	// lineChart.AddSeries(label, items)
	// lineChart.SetXAxis(bins)

	// page.AddCharts(
	// 	lineChart,
	// )

	// graph, err := os.Create(*out)
	// if err != nil {
	// 	panic(err)
	// }
	// page.Render(io.MultiWriter(graph))

}
