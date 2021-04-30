package blizzard

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("blizzard strayer", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	if count >= 2 {
		m := make(map[combat.StatType]float64)
		m[combat.CryoP] = 0.15
		c.AddMod("Blizzard Strayer 2PC", m)
	}
	if count >= 4 {
		s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
			if ds.Actor != c.Name() {
				return false
			}

			if s.TargetAura.E() == combat.Cryo {
				ds.Stats[combat.CR] += .2
				s.Log.Debugf("\t applying blizzard strayer 2pc buff on cryo target, new crit %v", ds.Stats[combat.CR])
			}

			if s.TargetAura.E() == combat.Frozen {
				ds.Stats[combat.CR] += .4
				s.Log.Debugf("\t applying blizzard strayer 4pc buff on frozen target, new crit %v", ds.Stats[combat.CR])
			}

			return false
		}, "blizzard strayer 4pc", combat.PreDamageHook)
	}
	//add flat stat to char
}
