package blizzard

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("Blizzard Strayer", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	if count >= 2 {
		m := make(map[combat.StatType]float64)
		m[combat.CryoP] = 0.15
		c.AddMod("Blizzard Strayer 2PC", m)
	}
	if count >= 4 {
		s.AddHook(func(snap *combat.Snapshot) bool {
			if snap.CharName != c.Name() {
				return false
			}

			//if len > 0 and first one is cryo then we can apply
			if len(s.Target.Auras) > 0 {
				if s.Target.Auras[0].Ele == combat.Cryo {
					s.Log.Debugf("\tapplying blizzard strayer 4pc buff on cryo target")
					snap.Stats[combat.CR] += .4
				}
			}

			return false
		}, "blizzard strayer 4pc", combat.PreDamageHook)
	}
	//add flat stat to char
}
