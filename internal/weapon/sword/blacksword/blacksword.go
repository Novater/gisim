package blacksword

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("The Black Sword", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	dmg := 0.15 + float64(r)*0.05
	//add on hit effect to sim?
	s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.AbilType == combat.ActionAttack || ds.AbilType == combat.ActionCharge {
			s.Log.Debugf("\t\tBlack sword adding %v dmg", dmg)
			ds.DmgBonus += dmg
		}
		return false
	}, "The Black Sword", combat.PreDamageHook)

}
