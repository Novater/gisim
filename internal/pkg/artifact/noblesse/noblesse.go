package noblesse

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("Noblesse Oblige", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	if count >= 2 {
		s.AddEffect(func(snap *combat.Snapshot) bool {
			if snap.CharName != c.Name() {
				return false
			}
			if snap.AbilType != combat.ActionTypeBurst {
				return false
			}
			snap.DmgBonus += 0.2

			return false
		}, "noblesse oblige 2pc", combat.PreDamageHook)
	}
	if count >= 4 {
		s.Log.Warnf("Noblesse Oblige 4PC bonus not yet implemented")
	}
	//add flat stat to char
}
