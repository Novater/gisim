package noblesse

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("Noblesse Oblige", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	if count >= 2 {
		s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
			// s.Log.Debugw("\t\tNoblesse 2 pc", "name", ds.CharName, "abil", ds.AbilType)
			if ds.Actor != c.Name() {
				return false
			}
			if ds.AbilType != combat.ActionBurst {
				return false
			}
			s.Log.Debugf("\t Noblesse 2 pc adding %v damage; pre buff %v", 0.2, ds.DmgBonus)
			ds.DmgBonus += 0.2

			return false
		}, "noblesse oblige 2pc", combat.PostSnapshot)
	}
	dur := 0
	if count >= 4 {
		//add an effect to count down duration
		s.AddEffect(func(s *combat.Sim) bool {
			dur--
			if dur < 0 {
				dur = 0
			}
			return false
		}, "noblesse oblige 4pc")

		s.AddEventHook(func(s *combat.Sim) bool {
			// s.Log.Debugw("\t\tNoblesse 2 pc", "name", ds.CharName, "abil", ds.AbilType)
			if s.ActiveChar != c.Name() {
				return false
			}
			s.Log.Debugf("\t Noblesse 4 pc proc'd")
			dur = 12 * 60

			return false
		}, "noblesse oblige 2pc", combat.PostBurstHook)

		s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
			//effect is not active, dont bother
			if dur <= 0 {
				return false
			}
			s.Log.Debugf("\t Noblesse 4 pc adding %v atk; pre buff %v", 0.2, ds.Stats[combat.ATKP])
			ds.Stats[combat.ATKP] += 0.2
			return false
		}, "noblesse oblige 4pc", combat.PostSnapshot)

	}
	//add flat stat to char
}
