package blizzard

import (
	"github.com/srliao/gisim/internal/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("Crimson Witch of Flames", set)
}

func set(c *combat.Char, s *combat.Sim, count int) {
	if count >= 2 {
		c.Mods["Crimson Witch of Flames 2PC"] = make(map[combat.StatType]float64)
		c.Mods["Crimson Witch of Flames 2PC"][combat.PyroP] = 0.15
	}
	if count >= 4 {
		s.AddEffect(func(snap *combat.Snapshot) bool {
			if snap.CharName != c.Profile.Name {
				return false
			}

			//check for vaporize
			// if snap.Element == combat.Pyro &&

			//check for melt

			//check for overload
			s.Log.Warnf("Crimson Witch 4PC bonus not yet implemented")

			return false
		}, "crimson witch 4pc", combat.PreReactionDamage)
	}
	//add flat stat to char
}
