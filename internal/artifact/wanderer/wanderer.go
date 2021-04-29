package wanderer

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("wanderer's troupe", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	if count >= 2 {
		m := make(map[combat.StatType]float64)
		m[combat.EM] = 80
		c.AddMod("Wanderer's Troupe 2PC", m)
	}
	if count >= 4 {
		//NOT YET IMPLEMENTED
		//we now need a weapon type flag....
		s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
			if ds.AbilType != combat.ActionCharge {
				return false
			}
			if ds.WeaponClass != combat.WeaponClassCatalyst && ds.WeaponClass != combat.WeaponClassBow {
				return false
			}
			ds.DmgBonus += 0.35
			return false
		}, "wanderer’s troupe 4pc", combat.PreDamageHook)
	}
	//add flat stat to char
}
