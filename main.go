package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/srliao/gisim/pkg/combat"

	//characters
	_ "github.com/srliao/gisim/internal/pkg/character/fischl"
	_ "github.com/srliao/gisim/internal/pkg/character/ganyu"
	_ "github.com/srliao/gisim/internal/pkg/character/xiangling"

	//weapons
	_ "github.com/srliao/gisim/internal/pkg/weapon/blacktassel"
	_ "github.com/srliao/gisim/internal/pkg/weapon/favoniuswarbow"
	_ "github.com/srliao/gisim/internal/pkg/weapon/prototypecrescent"
	_ "github.com/srliao/gisim/internal/pkg/weapon/skywardspine"

	//sets
	_ "github.com/srliao/gisim/internal/pkg/artifact/blizzard"
	_ "github.com/srliao/gisim/internal/pkg/artifact/crimson"
	_ "github.com/srliao/gisim/internal/pkg/artifact/gladiator"
	_ "github.com/srliao/gisim/internal/pkg/artifact/noblesse"
	_ "github.com/srliao/gisim/internal/pkg/artifact/wanderer"

	"gopkg.in/yaml.v2"
)

func main() {
	var source []byte
	var cfg combat.Profile
	var err error

	debugPtr := flag.String("d", "warn", "output level: debug, info, warn")
	secondsPtr := flag.Int("s", 60, "how many seconds to run the sim for")
	pPtr := flag.String("p", "config.yaml", "which profile to use")
	f := flag.String("o", "", "detailed log file")
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
	os.Remove(*f)

	s, err := combat.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	dmg := s.Run(*secondsPtr)
	elapsed := time.Since(start)
	log.Printf("Running profile %v, total damage dealt: %.2f over %v seconds. DPS = %.2f. Sim took %s\n", *pPtr, dmg, *secondsPtr, dmg/float64(*secondsPtr), elapsed)
}
