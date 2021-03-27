package blacksword

import "github.com/srliao/gisim/pkg/combat"

func init() {
	combat.RegisterWeaponFunc("Black Sword", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	dmg := 0.2
	switch r {
	case 2:
		dmg = .25
	case 3:
		dmg = .3
	case 4:
		dmg = .35
	case 5:
		dmg = .4
	}
	//add on hit effect to sim?
	s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.AbilType == combat.ActionTypeAttack || ds.AbilType == combat.ActionTypeChargedAttack {
			s.Log.Debugf("\t\tBlack sword adding %v dmg", dmg)
			ds.DmgBonus += dmg
		}
		return false
	}, "Black Sword", combat.PreDamageHook)

}
