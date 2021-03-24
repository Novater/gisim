package blizzard

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("Crimson Witch of Flames", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	if count >= 2 {
		m := make(map[combat.StatType]float64)
		m[combat.PyroP] = 0.15
		c.AddMod("Crimson Witch of Flames 2PC", m)
	}
	if count >= 4 {
		s.AddCombatHook(func(snap *combat.Snapshot) bool {
			if snap.CharName != c.Name() {
				return false
			}

			if !snap.WillReact {
				return false
			}

			if snap.ReactionType != combat.Melt && snap.ReactionType != combat.Vaporize && snap.ReactionType != combat.Overload {
				return false
			}

			switch snap.ReactionType {
			case combat.Melt:
				snap.ReactBonus += 0.15
			case combat.Vaporize:
				snap.ReactBonus += 0.15
			case combat.Overload:
				snap.ReactBonus += 0.4
			}

			//check for vaporize
			// if snap.Element == combat.Pyro &&

			//check for melt

			//check for overload
			s.Log.Warnf("Crimson Witch 4PC bonus not yet implemented")

			return false
		}, "crimson witch 4pc", combat.PreReaction)
	}
	//add flat stat to char
}
