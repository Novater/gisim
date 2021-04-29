package viridescent

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("archaic petra", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	//make use of closure to keep track of how many hits
	if count >= 2 {
		m := make(map[combat.StatType]float64)
		m[combat.GeoP] = 0.15
		c.AddMod("archaic petra 2pc", m)
	}
	//shard from crystallize, gain 35% bonus dmg to that element for 10s
	//only one form at a time
	if count >= 4 {
		//not implemented yet
	}
	//add flat stat to char
}
