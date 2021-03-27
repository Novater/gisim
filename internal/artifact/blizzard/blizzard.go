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
		s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
			if ds.CharName != c.Name() {
				return false
			}

			if len(s.Target.Auras) == 0 {
				return false
			}

			if s.Target.Auras[0].Ele == combat.Cryo {
				s.Log.Debugf("\tapplying blizzard strayer 2pc buff on cryo target")
				ds.Stats[combat.CR] += .2
			}

			if s.Target.Auras[0].Ele == combat.Frozen {
				s.Log.Debugf("\tapplying blizzard strayer 4pc buff on cryo target")
				ds.Stats[combat.CR] += .4
			}

			return false
		}, "blizzard strayer 4pc", combat.PreDamageHook)
	}
	//add flat stat to char
}
