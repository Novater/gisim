package blizzard

import (
	"github.com/srliao/gisim/pkg/combat"
	"go.uber.org/zap"
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
		s.AddEffect(func(snap *combat.Snapshot) bool {
			if snap.CharName != c.Name() {
				return false
			}

			if _, ok := s.Target.Auras[combat.Frozen]; ok {
				zap.S().Debugf("\tapplying blizzard strayer 4pc buff on frozen target")
				snap.Stats[combat.CR] += .4
			} else if _, ok := s.Target.Auras[combat.Cryo]; ok {
				zap.S().Debugf("\tapplying blizzard strayer 4pc buff on cryo target")
				snap.Stats[combat.CR] += .2
			}

			return false
		}, "blizzard strayer 4pc", combat.PreDamageHook)
	}
	//add flat stat to char
}
