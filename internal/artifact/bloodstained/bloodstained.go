package bloodstained

import "github.com/srliao/gisim/pkg/combat"

func init() {
	combat.RegisterSetFunc("bloodstained chivalry", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	if count >= 2 {
		m := make(map[combat.StatType]float64)
		m[combat.PhyP] = 0.25
		c.AddMod("Bloodstained Chivalry 2PC", m)
	}
	//add flat stat to char
}
