package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/profile"
	"github.com/srliao/gisim/internal/parse"
	"github.com/srliao/gisim/pkg/combat"

	//characters
	_ "github.com/srliao/gisim/internal/character/bennett"
	_ "github.com/srliao/gisim/internal/character/diona"
	_ "github.com/srliao/gisim/internal/character/eula"
	_ "github.com/srliao/gisim/internal/character/fischl"
	_ "github.com/srliao/gisim/internal/character/ganyu"
	_ "github.com/srliao/gisim/internal/character/sucrose"
	_ "github.com/srliao/gisim/internal/character/xiangling"
	_ "github.com/srliao/gisim/internal/character/xingqiu"

	//weapons
	_ "github.com/srliao/gisim/internal/weapon/bow/favoniuswarbow"
	_ "github.com/srliao/gisim/internal/weapon/bow/prototypecrescent"
	_ "github.com/srliao/gisim/internal/weapon/catalyst/sacrificial"
	_ "github.com/srliao/gisim/internal/weapon/claymore/archaic"
	_ "github.com/srliao/gisim/internal/weapon/claymore/favonius"
	_ "github.com/srliao/gisim/internal/weapon/claymore/pine"
	_ "github.com/srliao/gisim/internal/weapon/claymore/skyrider"
	_ "github.com/srliao/gisim/internal/weapon/claymore/skyward"
	_ "github.com/srliao/gisim/internal/weapon/claymore/spine"
	_ "github.com/srliao/gisim/internal/weapon/claymore/starsilver"
	_ "github.com/srliao/gisim/internal/weapon/claymore/wolf"
	_ "github.com/srliao/gisim/internal/weapon/spear/blacktassel"
	_ "github.com/srliao/gisim/internal/weapon/spear/favonius"
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
	_ "github.com/srliao/gisim/internal/artifact/paleflame"
	_ "github.com/srliao/gisim/internal/artifact/viridescent"
	_ "github.com/srliao/gisim/internal/artifact/wanderer"
)

func main() {
	var source []byte

	var err error

	debug := flag.String("d", "error", "output level: debug, info, warn")
	seconds := flag.Int("s", 120, "how many seconds to run the sim for")
	cfgFile := flag.String("p", "config.txt", "which profile to use")
	f := flag.String("o", "", "detailed log file")
	hp := flag.Float64("hp", 0, "hp mode: how much hp to deal damage to")
	showCaller := flag.Bool("caller", false, "show caller in debug low")
	avgMode := flag.Bool("a", false, "run sim multiple times and calculate avg damage (smooth out randomness). default false. note that there is no debug log in this mode")
	w := flag.Int("w", 24, "number of workers to run when running multiple iterations; default 24")
	i := flag.Int("i", 5000, "number of iterations to run if we're running multiple")
	multi := flag.String("comp", "", "comparison mode")

	flag.Parse()

	// p := "./xl-base.yaml" //xl.yaml expecting 4659 dps

	source, err = ioutil.ReadFile(*cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	switch {
	case *multi != "":
		content, err := ioutil.ReadFile(*multi)
		if err != nil {
			log.Fatal(err)
		}
		files := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
		// lines := strings.Split(string(content), `\n`)
		runMulti(*i, *w, files, *hp, *seconds)
	case *avgMode:
		r := runIter(*i, *w, source, *hp, *seconds)
		fmt.Printf(
			"Simulation done; %v iterations; average dps of %0.0f over %v seconds (min: %0.02f, max: %0.02f, stddev: %0.02f)\n",
			*i,
			r.mean,
			*seconds,
			r.min,
			r.max,
			r.sd,
		)
	default:
		defer profile.Start(profile.ProfilePath("./")).Stop()
		parser := parse.New("single", string(source))
		cfg, err := parser.Parse()
		if err != nil {
			log.Fatal(err)
		}
		cfg.LogConfig.LogLevel = *debug
		cfg.LogConfig.LogFile = *f
		cfg.LogConfig.LogShowCaller = *showCaller
		cfg.LogStats = true
		os.Remove(*f)
		runSingle(cfg, *hp, *seconds)
	}

}

func frameDataToCSV(s combat.SimStats) {
	os.Remove("result.csv")
	file, err := os.Create("result.csv")
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	chars := make([]string, 0, len(s.CharActiveFrame))
	ele := make([]combat.EleType, 0, len(s.ElementActiveFrame))

	header := make([]string, len(s.CharActiveFrame)+len(s.ElementActiveFrame)+3)
	header[0] = "Character"
	header[1] = "Action"
	header[2] = "Action Param"

	i := 3
	for k := range s.CharActiveFrame {
		chars = append(chars, k)

	}
	sort.Strings(chars)
	for _, v := range chars {
		header[i] = v
		i++
	}
	for k := range s.ElementActiveFrame {
		ele = append(ele, k)
		header[i] = k.String()
		i++
	}

	writer.Write(header)

	elen := len(ele)
	clen := len(chars)
	log.Println(elen, clen)

	for i, a := range s.AbilUsageByFrame {

		record := make([]string, elen+clen+3)

		record[0] = a.Actor
		record[1] = a.Action
		record[2] = strconv.FormatInt(int64(a.Param), 10)
		n := 3
		for _, k := range chars {
			record[n] = "0"
			if s.CharActiveFrame[k][i] == 1 {
				record[n] = "1"
			}
			n++
		}

		for _, k := range ele {
			record[n] = "0"
			if s.ElementActiveFrame[k][i] == 1 {
				record[n] = "1"
			}
			n++
		}

		writer.Write(record)
	}

}

func graphToCSV(in []float64) {
	os.Remove("result.csv")
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

func runSingle(cfg combat.Profile, hp float64, dur int) {

	s, err := combat.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	var stats combat.SimStats
	var dmg float64

	start := time.Now()
	if hp > 0 {
		dmg, stats = s.RunHPMode(hp)
		dur = stats.SimDuration / 60
	} else {
		dmg, stats = s.Run(dur)
	}
	elapsed := time.Since(start)
	fmt.Println("------------------------------------------")
	for char, t := range stats.DamageByChar {
		fmt.Printf("%v contributed the following dps:\n", char)
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var total float64
		for _, k := range keys {
			v := t[k]
			fmt.Printf("\t%v: %.2f (%.2f%%; total = %.0f)\n", k, v/float64(dur), 100*v/dmg, v)
			total += v
		}

		fmt.Printf("%v total dps: %.2f (dmg: %.2f); total percentage: %.0f%%\n", char, total/float64(dur), total, 100*total/dmg)
	}
	fmt.Println("------------------------------------------")
	for char, t := range stats.AbilUsageCountByChar {
		fmt.Printf("%v used the following abilities:\n", char)
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := t[k]
			fmt.Printf("\t%v: %v\n", k, v)
		}
	}
	fmt.Println("------------------------------------------")
	ck := make([]string, 0, len(stats.CharActiveTime))
	for k := range stats.CharActiveTime {
		ck = append(ck, k)
	}
	for _, k := range ck {
		v := stats.CharActiveTime[k]
		fmt.Printf("%v active for %v (%v seconds - %.0f%%)\n", k, v, v/60, 100*float64(v)/float64(dur*60))
	}
	fmt.Println("------------------------------------------")
	tk := make([]combat.EleType, 0, len(stats.AuraUptime))
	for k := range stats.AuraUptime {
		tk = append(tk, k)
	}
	for _, k := range tk {
		v := stats.AuraUptime[k]
		fmt.Printf("%v active for %v (%v seconds - %.0f%%)\n", k, v, v/60, 100*float64(v)/float64(dur*60))
	}
	fmt.Println("------------------------------------------")
	rk := make([]combat.ReactionType, 0, len(stats.ReactionsTriggered))
	for k := range stats.ReactionsTriggered {
		rk = append(rk, k)
	}
	for _, k := range rk {
		v := stats.ReactionsTriggered[k]
		fmt.Printf("%v: %v\n", k, v)
	}
	fmt.Println("------------------------------------------")
	fmt.Printf("Running profile %v, total damage dealt: %.2f over %v seconds. DPS = %.2f. Sim took %s\n", cfg.Label, dmg, dur, dmg/float64(dur), elapsed)

	if hp == 0 {
		frameDataToCSV(stats)
	}
}

type result struct {
	mean float64
	min  float64
	max  float64
	sd   float64
}

func runIter(n, w int, src []byte, hp float64, dur int) result {
	// var progress float64
	var sum, ss, min, max float64
	var data []float64
	min = math.MaxFloat64
	max = -1

	count := n

	resp := make(chan float64, n)
	req := make(chan bool)
	done := make(chan bool)

	for i := 0; i < w; i++ {
		go worker(src, hp, dur, resp, req, done)
	}

	go func() {
		var wip int
		for wip < n {
			req <- true
			wip++
		}
	}()

	// fmt.Print("0")

	for count > 0 {
		val := <-resp
		count--
		data = append(data, val)
		sum += val
		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
		// if (1 - float64(count)/float64(n)) > (progress + 0.1) {
		// 	progress = (1 - float64(count)/float64(n))
		// 	fmt.Printf(".%.0f", 100*progress)
		// }
	}
	// fmt.Print(".100")

	close(done)

	mean := sum / float64(n)

	for _, v := range data {
		ss += (v - mean) * (v - mean)
	}

	sd := math.Sqrt(ss / float64(n))

	return result{
		mean: mean,
		min:  min,
		max:  max,
		sd:   sd,
	}
}

func worker(src []byte, hp float64, dur int, resp chan float64, req chan bool, done chan bool) {

	for {
		select {
		case <-req:
			parser := parse.New("single", string(src))
			cfg, err := parser.Parse()
			if err != nil {
				log.Fatal(err)
			}
			cfg.LogConfig.LogLevel = "error"
			cfg.LogConfig.LogFile = ""
			cfg.LogConfig.LogShowCaller = false
			cfg.LogStats = false

			s, err := combat.New(cfg)
			if err != nil {
				log.Fatal(err)
			}

			if hp > 0 {
				dmg, stat := s.RunHPMode(hp)
				resp <- dmg * 60 / float64(stat.SimDuration)
			} else {
				dmg, _ := s.Run(dur)
				resp <- dmg / float64(dur)
			}

		case <-done:
			return
		}
	}
}

func runMulti(n, w int, files []string, hp float64, dur int) {
	fmt.Printf("Simulating %v seconds of combat over %v iterations\n", dur, n)
	start := time.Now()
	fmt.Print("Filename                                 |      Mean|       Min|       Max|   Std Dev|\n")
	fmt.Print("--------------------------------------------------------------------------------------\n")
	for _, f := range files {
		source, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%40.40v |", f)
		r := runIter(n, w, source, hp, dur)
		fmt.Printf("%10.2f|%10.2f|%10.2f|%10.2f|\n", r.mean, r.min, r.max, r.sd)
	}
	elapsed := time.Since(start)
	fmt.Printf("Completed in %s\n", elapsed)
}

// 0.10.20.30.40.50.60.70.80.90.100
