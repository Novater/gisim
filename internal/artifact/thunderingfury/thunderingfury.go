package thunderingfury

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("Thundering Fury", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	//make use of closure to keep track of how many hits
	if count >= 2 {
		m := make(map[combat.StatType]float64)
		m[combat.ElectroP] = 0.15
		c.AddMod("thundering fury 2pc", m)
	}
	icd := 0
	//increase dmg caused by overload, ec, superconduct by 40%
	//decrease ele cd by 1s every 0.8s
	if count >= 4 {
		s.AddSnapshotHook(func(snap *combat.Snapshot) bool {
			if snap.Actor != c.Name() {
				return false
			}

			if !snap.WillReact {
				return false
			}

			if snap.ReactionType != combat.Overload && snap.ReactionType != combat.ElectroCharged && snap.ReactionType != combat.Superconduct {
				return false
			}

			if icd >= s.F {
				return false
			}

			snap.ReactBonus += 0.4
			icd = s.F + 48
			c.ReduceActionCooldown(combat.ActionSkill, 60)

			return false
		}, "thundering fury 4pc", combat.PreReaction)
	}
	//add flat stat to char
}
