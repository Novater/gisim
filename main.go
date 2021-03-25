package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/srliao/gisim/pkg/combat"

	//characters
	_ "github.com/srliao/gisim/internal/character/bennett"
	_ "github.com/srliao/gisim/internal/character/fischl"
	_ "github.com/srliao/gisim/internal/character/ganyu"
	_ "github.com/srliao/gisim/internal/character/xiangling"
	_ "github.com/srliao/gisim/internal/character/xingqiu"

	//weapons
	_ "github.com/srliao/gisim/internal/weapon/bow/favoniuswarbow"
	_ "github.com/srliao/gisim/internal/weapon/bow/prototypecrescent"
	_ "github.com/srliao/gisim/internal/weapon/spear/blacktassel"
	_ "github.com/srliao/gisim/internal/weapon/spear/skywardspine"
	_ "github.com/srliao/gisim/internal/weapon/sword/festeringdesire"
	_ "github.com/srliao/gisim/internal/weapon/sword/sacrificialsword"

	//sets
	_ "github.com/srliao/gisim/internal/artifact/blizzard"
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

	debugPtr := flag.String("d", "debug", "output level: debug, info, warn")
	secondsPtr := flag.Int("s", 300, "how many seconds to run the sim for")
	pPtr := flag.String("p", "config.yaml", "which profile to use")
	f := flag.String("o", "out.log", "detailed log file")
	showCaller := flag.Bool("c", false, "show caller in debug low")
	w := flag.String("w", "", "test weights for specified character")
	flag.Parse()

	// p := "./xl-base.yaml" //xl.yaml expecting 4659 dps

	source, err = ioutil.ReadFile(*pPtr)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(source, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	cfg.LogLevel = *debugPtr
	cfg.LogFile = *f
	cfg.LogShowCaller = *showCaller
	os.Remove(*f)

	if *w != "" {
		stat := []combat.Stat{
			{
				Type:  combat.ATK,
				Value: 11,
			},
			{
				Type:  combat.ATKP,
				Value: 0.033,
			},
			{
				Type:  combat.EM,
				Value: 13,
			},
			{
				Type:  combat.ER,
				Value: 0.036,
			},
			{
				Type:  combat.CR,
				Value: 0.022,
			},
			{
				Type:  combat.CD,
				Value: 0.044,
			},
		}

		start := time.Now()
		result := make(map[combat.StatType]float64)

		cfg.LogFile = ""
		cfg.LogLevel = "error"

		ok := false
		name := *w
		for _, c := range cfg.Characters {
			if c.Name == name {
				ok = true
			}
		}
		if !ok {
			log.Panic("invalid character to test weights")
		}

		dur := 6000

		s2, err := combat.New(cfg)
		if err != nil {
			log.Fatal(err)
		}

		d1, _ := s2.Run(dur)

		for _, v := range stat {

			start := time.Now()
			fmt.Printf("Testing %v stat %v\n", name, v.Type)

			s, err := combat.New(cfg)
			if err != nil {
				log.Fatal(err)
			}

			val := make(map[combat.StatType]float64)
			val[v.Type] = v.Value * 4
			s.AddCharMod(name, "test", val)
			d, _ := s.Run(dur)

			elapsed := time.Since(start)

			fmt.Printf("Increasing %v by %0.4f; New dps: %0.2f old dps: %0.2f;  team dps increased %0.6f%%, took %v\n", v.Type, v.Value*4, d/float64(dur), d1/float64(dur), (d/d1-1)*100, elapsed)
			result[v.Type] = d/d1 - 1
		}

		elapsed := time.Since(start)
		fmt.Printf("Finished in %v seconds\n", elapsed)
		return
	}

	s, err := combat.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	dmg, details := s.Run(*secondsPtr)
	elapsed := time.Since(start)
	for char, t := range details {
		fmt.Printf("%v contributed the following dps:\n", char)
		for k, v := range t {
			fmt.Printf("\t%v: %.2f (%.2f%%)\n", k, v/float64(*secondsPtr), 100*v/dmg)
		}
	}
	fmt.Printf("Running profile %v, total damage dealt: %.2f over %v seconds. DPS = %.2f. Sim took %s\n", *pPtr, dmg, *secondsPtr, dmg/float64(*secondsPtr), elapsed)

}
