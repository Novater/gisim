package noblesse

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("Noblesse Oblige", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	if count >= 2 {
		s.AddHook(func(ds *combat.Snapshot) bool {
			// s.Log.Debugw("\t\tNoblesse 2 pc", "name", ds.CharName, "abil", ds.AbilType)
			if ds.CharName != c.Name() {
				return false
			}
			if ds.AbilType != combat.ActionTypeBurst {
				return false
			}
			s.Log.Debugf("\t\t\tNoblesse 2 pc adding % damage")
			ds.DmgBonus += 0.2

			return false
		}, "noblesse oblige 2pc", combat.PreSnapshot)
	}
	if count >= 4 {
		s.Log.Warnf("Noblesse Oblige 4PC bonus not yet implemented")
	}
	//add flat stat to char
}
