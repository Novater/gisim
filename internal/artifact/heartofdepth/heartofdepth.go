package heartofdepth

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("Heart of Depth", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	//make use of closure to keep track of how many hits
	if count >= 2 {
		m := make(map[combat.StatType]float64)
		m[combat.HydroP] = 0.15
		c.AddMod("heart of depth 2pc", m)
	}
	//after skill, increase normal and charged
	if count >= 4 {
		buff := 0
		s.AddEventHook(func(s *combat.Sim) bool {
			buff = s.F + 900 //activate for 15 seoncds
			return false
		}, "heart of depth 4pc", combat.PostSkillHook)
		s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
			if ds.Actor != c.Name() {
				return false
			}
			if buff < s.F {
				return false
			}
			//only affect normal and charge
			if ds.AbilType != combat.ActionAttack && ds.AbilType != combat.ActionCharge {
				return false
			}
			ds.DmgBonus += 0.30

			return false
		}, "heart of depth 4pc proc", combat.PreDamageHook)
	}
	//add flat stat to char
}
