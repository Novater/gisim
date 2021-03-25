package festeringdesire

import "github.com/srliao/gisim/pkg/combat"

func init() {
	combat.RegisterWeaponFunc("Festering Desire", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	//add on hit effect to sim?
	dmg := 0.16
	crit := 0.06
	switch r {
	case 2:
		dmg = .2
		crit = .075
	case 3:
		dmg = .24
		crit = .09
	case 4:
		dmg = .28
		crit = .105
	case 5:
		dmg = .32
		crit = .12
	}

	s.AddCombatHook(func(ds *combat.Snapshot) bool {
		if ds.CharName != c.Name() {
			return false
		}
		if ds.AbilType == combat.ActionTypeSkill {
			s.Log.Debugf("\t\tFestering desire adding %v dmg %v crit", dmg, crit)
			ds.Stats[combat.CR] += crit
			ds.DmgBonus += dmg
		}
		return false
	}, "Festering Desire", combat.PreDamageHook)

}
