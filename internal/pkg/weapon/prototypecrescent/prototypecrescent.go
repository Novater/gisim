package prototypecrescent

import (
	"fmt"

	"github.com/srliao/gisim/internal/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("Prototype Crescent", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	//add on hit effect to sim?
	s.AddEffect(func(snap *combat.Snapshot) bool {
		//check if char is correct?
		if snap.CharName != c.Name() {
			return false
		}
		//check if weakpoint triggered
		if !snap.HitWeakPoint {
			return false
		}
		//add a new action that adds % dmg to current char and removes itself after
		//10 seconds
		tick := 0
		s.AddAction(func(s *combat.Sim) bool {
			if tick >= 10*60 {
				c.RemoveMod("Prototype-Crescent-Proc")
				s.Log.Debugw("\tprototype crescent buff expired", "tick", tick)
				return true
			}
			tick++
			if !c.HasMod("Prototype-Crescent-Proc") {
				m := make(map[combat.StatType]float64)
				atkmod := 0.36
				switch r {
				case 2:
					atkmod = 0.45
				case 3:
					atkmod = 0.54
				case 4:
					atkmod = 0.63
				case 5:
					atkmod = 0.72
				}
				m[combat.ATKP] = atkmod
				c.AddMod("Prototype-Crescent-Proc", m)
			}

			return false
		}, fmt.Sprintf("Prototype-Crescent-Proc-%v", c.Name()))
		return false
	}, "prototype-crescent-proc", combat.PostDamageHook)
}
