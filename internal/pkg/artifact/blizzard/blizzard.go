package blizzard

import (
	"github.com/srliao/gisim/internal/pkg/combat"
	"go.uber.org/zap"
)

func init() {
	combat.RegisterSetFunc("Blizzard Strayer", set)
}

func set(c *combat.Char, s *combat.Sim, count int) {
	if count >= 2 {
		c.Mods["Blizzard Strayer 2PC"] = make(map[combat.StatType]float64)
		c.Mods["Blizzard Strayer 2PC"][combat.CryoP] = 0.15
	}
	if count >= 4 {
		s.AddEffect(func(snap *combat.Snapshot) bool {
			if snap.CharName != c.Profile.Name {
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
