package noblesse

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("Noblesse Oblige", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	if count >= 2 {
		s.AddCombatHook(func(ds *combat.Snapshot) bool {
			// s.Log.Debugw("\t\tNoblesse 2 pc", "name", ds.CharName, "abil", ds.AbilType)
			if ds.CharName != c.Name() {
				return false
			}
			if ds.AbilType != combat.ActionTypeBurst {
				return false
			}
			s.Log.Debugf("\t\t\tNoblesse 2 pc adding %v damage; pre buff %v", 0.2, ds.DmgBonus)
			ds.DmgBonus += 0.2

			return false
		}, "noblesse oblige 2pc", combat.PostSnapshot)
	}
	if count >= 4 {
		s.AddEventHook(func(s *combat.Sim) bool {
			// s.Log.Debugw("\t\tNoblesse 2 pc", "name", ds.CharName, "abil", ds.AbilType)
			if s.ActiveChar != c.Name() {
				return false
			}
			//add bonus hook to all dmg char for 12seconds
			//refresh duration if already exists
			s.Log.Debugf("\t\t\tNoblesse 4 pc proc'd")

			s.Status["Noblesse Oblige 4PC"] = 12 * 60

			//this should forcefully overwrite any existing
			tick := 0
			s.AddCombatHook(func(ds *combat.Snapshot) bool {
				tick++
				if tick == 12*60 {
					return true
				}
				s.Log.Debugf("\t\t\tNoblesse 4 pc adding %v atk; pre buff %v", 0.2, ds.Stats[combat.ATKP])
				ds.Stats[combat.ATKP] += 0.2
				return false
			}, "noblesse oblige 4pc", combat.PostSnapshot)

			return false
		}, "noblesse oblige 4pc", combat.PreBurstHook)

	}
	//add flat stat to char
}
