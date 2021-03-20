package main

import (
	"flag"
	"io/ioutil"
	"log"
	"time"

	"github.com/srliao/gisim/internal/pkg/combat"

	//characters
	_ "github.com/srliao/gisim/internal/pkg/character/ganyu"
	_ "github.com/srliao/gisim/internal/pkg/character/xiangling"

	//weapons
	_ "github.com/srliao/gisim/internal/pkg/weapon/blacktassel"
	_ "github.com/srliao/gisim/internal/pkg/weapon/prototypecrescent"
	_ "github.com/srliao/gisim/internal/pkg/weapon/skywardspine"

	//sets
	_ "github.com/srliao/gisim/internal/pkg/artifact/blizzard"
	_ "github.com/srliao/gisim/internal/pkg/artifact/crimson"
	_ "github.com/srliao/gisim/internal/pkg/artifact/noblesse"

	"gopkg.in/yaml.v2"
)

func main() {

	debugPtr := flag.String("d", "warn", "output level: debug, info, warn")
	secondsPtr := flag.Int("s", 20, "how many seconds to run the sim for")
	flag.Parse()

	var source []byte
	var cfg combat.Profile
	var err error

	p := "./xl.yaml"

	source, err = ioutil.ReadFile(p)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(source, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	cfg.LogLevel = *debugPtr

	s, err := combat.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	var actions = []combat.Action{
		{
			TargetCharIndex: 0,
			Type:            combat.ActionTypeSkill,
		},
		{
			TargetCharIndex: 0,
			Type:            combat.ActionTypeBurst,
		},
		{
			TargetCharIndex: 0,
			Type:            combat.ActionTypeChargedAttack,
		},
		// {
		// 	TargetCharIndex: 0,
		// 	Type:            combat.ActionTypeAttack,
		// },
	}
	start := time.Now()
	dmg := s.Run(*secondsPtr, actions)
	elapsed := time.Since(start)
	log.Printf("Running profile %v, total damage dealt: %.2f over %v seconds. DPS = %.2f. Sim took %s\n", p, dmg, *secondsPtr, dmg/float64(*secondsPtr), elapsed)
}
