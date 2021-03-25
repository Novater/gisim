package favonius

import (
	"fmt"
	"math/rand"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("Favonius Sword", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	p := 0.6
	cd := 120 * 6
	switch r {
	case 2:
		p = .7
		cd = 105 * 6
	case 3:
		p = .8
		cd = 90 * 6
	case 4:
		p = .9
		cd = 75 * 6
	case 5:
		p = .10
		cd = 60 * 6
	}
	//add on crit effect
	s.AddCombatHook(func(ds *combat.Snapshot) bool {
		//check if char is correct?
		if ds.CharName != c.Name() {
			return false
		}
		if _, ok := s.Status["Favonius Sword Proc"]; ok {
			return false
		}

		if rand.Float64() > p {
			return false
		}
		s.Log.Debugf("[%v] Favonius Sword proc triggered", s.Frame())

		orbDelay := 0
		s.AddAction(func(s *combat.Sim) bool {
			if orbDelay < 90+60 { //it takes 90 frames to generate orb, add another 60 frames to get it
				orbDelay++
				return false
			}
			s.GenerateOrb(3, combat.NonElemental, false)
			return true
		}, fmt.Sprintf("%v-Favonius-Sword-Orb", s.Frame()))

		s.Status["Favonius Sword Proc"] = cd
		return false
	}, "favonius-sword-proc", combat.OnCritDamage)

}
