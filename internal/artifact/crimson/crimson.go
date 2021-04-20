package crimson

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("Crimson Witch of Flames", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	//make use of closure to keep track of how many hits
	stacks := 0
	dur := 0

	//effect lasts 10 seconds, max 3 stacks
	if count >= 2 {
		m := make(map[combat.StatType]float64)
		m[combat.PyroP] = 0.15
		c.AddMod("Crimson Witch of Flames 2PC", m)
	}
	if count >= 4 {
		//counter to reduce duration
		s.AddEffect(func(s *combat.Sim) bool {
			//tick down cd
			dur--
			if dur < 0 {
				dur = 0
				stacks = 0 //stacks all gone if duration is gone
			}
			return false
		}, "Crimson Witch 4PC")

		//post snap shot to increase stacks
		s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
			if ds.Actor != c.Name() {
				return false
			}
			if ds.AbilType != combat.ActionSkill {
				return false
			}
			//every exectuion, add 1 stack, to a max of 3, reset cd to 10 seconds
			stacks++
			if stacks > 3 {
				stacks = 3
			}
			dur = 600
			return false
		}, "crimson witch 4pc", combat.PostSnapshot)

		s.AddSnapshotHook(func(snap *combat.Snapshot) bool {
			if snap.Actor != c.Name() {
				return false
			}

			if !snap.WillReact {
				return false
			}

			if snap.ReactionType != combat.Melt && snap.ReactionType != combat.Vaporize && snap.ReactionType != combat.Overload {
				return false
			}

			//increase bonus by 50% per stack
			mult := 0.5*float64(stacks) + 1

			switch snap.ReactionType {
			case combat.Melt:
				snap.ReactBonus += (0.15 * mult)
			case combat.Vaporize:
				snap.ReactBonus += (0.15 * mult)
			case combat.Overload:
				snap.ReactBonus += (0.4 * mult)
			}

			return false
		}, "crimson witch 4pc", combat.PreReaction)
	}
	//add flat stat to char
}
