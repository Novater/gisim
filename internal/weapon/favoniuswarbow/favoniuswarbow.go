package favoniuswarbow

import (
	"fmt"
	"math/rand"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("Favonius Warbow", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {

	prob := 0.6
	cd := 12 * 60
	switch r {
	case 2:
		prob = .7
		cd = 105 * 6
	case 3:
		prob = .8
		cd = 9 * 60
	case 4:
		prob = .9
		cd = 75 * 6
	case 5:
		prob = 1
		cd = 6 * 60
	}
	//add on crit effect
	s.AddCombatHook(func(snap *combat.Snapshot) bool {
		//check if char is correct?
		if snap.CharName != c.Name() {
			return false
		}
		if _, ok := s.Status["Favonius Warbow Proc"]; ok {
			return false
		}

		if rand.Float64() > prob {
			return false
		}
		s.Log.Debugf("[%v] Favonius Warbox proc triggered", s.Frame())

		orbDelay := 0
		s.AddAction(func(s *combat.Sim) bool {
			if orbDelay < 90+60 { //it takes 90 frames to generate orb, add another 60 frames to get it
				orbDelay++
				return false
			}
			s.GenerateOrb(3, combat.NonElemental, false)
			return true
		}, fmt.Sprintf("%v-FavoniusWarbow-Orb", s.Frame()))

		s.Status["Favonius Warbow Proc"] = cd
		return false
	}, "favonius-warbow-proc", combat.OnCritDamage)
}
