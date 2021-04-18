package gladiator

import (
	"github.com/srliao/gisim/internal/rotation"
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("Gladiator's Finale", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	if count >= 2 {
		m := make(map[combat.StatType]float64)
		m[combat.ATKP] = 0.18
		c.AddMod("Gladiator's Finale 2PC", m)
	}
	if count >= 4 {
		s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
			if ds.AbilType != rotation.ActionAttack {
				return false
			}
			if ds.WeaponClass != combat.WeaponClassSpear && ds.WeaponClass != combat.WeaponClassSword && ds.WeaponClass != combat.WeaponClassClaymore {
				return false
			}
			ds.DmgBonus += 0.35
			return false
		}, "gladiator's finale 4pc", combat.PostSnapshot)
	}
	//add flat stat to char
}
