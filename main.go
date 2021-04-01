package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/pkg/profile"
	"github.com/srliao/gisim/pkg/combat"

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
	_ "github.com/srliao/gisim/internal/weapon/sword/blackcliff"
	_ "github.com/srliao/gisim/internal/weapon/sword/festeringdesire"
	_ "github.com/srliao/gisim/internal/weapon/sword/sacrificialsword"

	//sets
	_ "github.com/srliao/gisim/internal/artifact/blizzard"
	_ "github.com/srliao/gisim/internal/artifact/bloodstained"
	_ "github.com/srliao/gisim/internal/artifact/crimson"
	_ "github.com/srliao/gisim/internal/artifact/gladiator"
	_ "github.com/srliao/gisim/internal/artifact/noblesse"
	_ "github.com/srliao/gisim/internal/artifact/wanderer"

	"gopkg.in/yaml.v2"
)

func main() {
	var source []byte
	var cfg combat.Profile
	var err error

	debugPtr := flag.String("d", "error", "output level: debug, info, warn")
	secondsPtr := flag.Int("s", 600, "how many seconds to run the sim for")
	pPtr := flag.String("p", "basetest.yaml", "which profile to use")
	f := flag.String("o", "", "detailed log file")
	showCaller := flag.Bool("c", false, "show caller in debug low")
	flag.Parse()

	defer profile.Start(profile.ProfilePath("./")).Stop()

	// p := "./xl-base.yaml" //xl.yaml expecting 4659 dps

	source, err = ioutil.ReadFile(*pPtr)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(source, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	cfg.LogConfig.LogLevel = *debugPtr
	cfg.LogConfig.LogFile = *f
	cfg.LogConfig.LogShowCaller = *showCaller
	os.Remove(*f)

	s, err := combat.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	dmg, details, graph := s.Run(*secondsPtr)
	elapsed := time.Since(start)
	for char, t := range details {
		fmt.Printf("%v contributed the following dps:\n", char)
		var total float64
		for k, v := range t {
			fmt.Printf("\t%v: %.2f (%.2f%%; total = %.0f)\n", k, v/float64(*secondsPtr), 100*v/dmg, v)
			total += v
		}
		fmt.Printf("%v total dps: %.2f (dmg: %.2f); total percentage: %.0f\n", char, total/float64(*secondsPtr), total, 100*total/dmg)
	}
	fmt.Printf("Running profile %v, total damage dealt: %.2f over %v seconds. DPS = %.2f. Sim took %s\n", *pPtr, dmg, *secondsPtr, dmg/float64(*secondsPtr), elapsed)

	graphToCSV(graph)

}

func graphToCSV(in []float64) {
	os.Remove("graph.csv")
	file, err := os.Create("result.csv")
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	t := 10
	l := len(in) / (60 * t)

	out := make([][]string, 0, l)

	stop := false
	next := 0
	var prev float64
	var i int64
	for !stop {
		val := in[next] - prev
		prev = in[next]
		out = append(out, []string{strconv.FormatInt(i, 10), strconv.FormatFloat(val/float64(t), 'f', 2, 64)})
		i++
		if next == len(in)-1 {
			stop = true
		} else {
			next += 600
			if next >= len(in) {
				next = len(in) - 1
			}
		}
	}

	for _, v := range out {
		writer.Write(v)
	}

}
