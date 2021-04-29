package archaic

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("serpent spine", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	buff := 0.05 + float64(r)*.01
	last := 0
	stacks := 0

	s.AddTask(func(s *combat.Sim) {
		if s.ActiveChar == c.Name() {
			s.Log.Debugf("setting initial spine stacks to 5")
			stacks = 5
			val := make(map[combat.StatType]float64)
			val[combat.DmgP] = buff * float64(stacks)
			c.AddMod("spine", val)
		}
	}, "spine-init", 0)

	s.AddEffect(func(s *combat.Sim) bool {
		//if char is not active then everything is reset
		if s.ActiveChar != c.Name() {
			last = s.F
			return false
		}
		if stacks == 5 {
			return false
		}
		//if more than 4 sec has passed last
		if s.F-last > 240 {
			last = s.F //reset to current spot
			stacks++
			s.Log.Debugf("spine updating stacks to %v", stacks)
			//update the buff on the char
			val := make(map[combat.StatType]float64)
			val[combat.DmgP] = buff * float64(stacks)
			c.AddMod("spine", val)
		}

		return false
	}, "spine")

}
