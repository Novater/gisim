package main

import (
	"flag"
	"io/ioutil"
	"log"
	"time"

	"github.com/srliao/gisim/internal/pkg/combat"
	//characters
	_ "github.com/srliao/gisim/internal/pkg/character/ganyu"
	//sets
	_ "github.com/srliao/gisim/internal/pkg/artifact/blizzard"
	//weapons
	_ "github.com/srliao/gisim/internal/pkg/weapon/prototypecrescent"
	"gopkg.in/yaml.v2"
)

func main() {

	debugPtr := flag.String("d", "warn", "output level: debug, info, warn")
	flag.Parse()

	var source []byte
	var cfg combat.Profile
	var err error

	p := "./current.yaml"

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
		// {
		// 	TargetCharIndex: 0,
		// 	Type:            ActionTypeBurst,
		// },
		{
			TargetCharIndex: 0,
			Type:            combat.ActionTypeChargedAttack,
		},
	}
	start := time.Now()
	seconds := 6
	dmg := s.Run(seconds, actions)
	elapsed := time.Since(start)
	log.Printf("Running profile %v, total damage dealt: %.2f over %v seconds. DPS = %.2f. Sim took %s\n", p, dmg, seconds, dmg/float64(seconds), elapsed)
}
