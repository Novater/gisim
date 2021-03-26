package prototypecrescent

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("Prototype Crescent", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	atkmod := 0.36
	switch r {
	case 2:
		atkmod = 0.45
	case 3:
		atkmod = 0.54
	case 4:
		atkmod = 0.63
	case 5:
		atkmod = 0.72
	}
	//add on hit effect
	s.AddCombatHook(func(snap *combat.Snapshot) bool {
		//check if char is correct?
		if snap.CharName != c.Name() {
			return false
		}
		//check if weakpoint triggered
		if !snap.HitWeakPoint {
			return false
		}
		s.Status["Prototype Crescent"] = 600
		return false
	}, "prototype-crescent-proc", combat.PostDamageHook)

	//add snapshot effect
	s.AddCombatHook(func(ds *combat.Snapshot) bool {
		if c.Name() != ds.CharName {
			return false
		}
		if _, ok := s.Status["Prototype Crescent"]; !ok {
			return false
		}
		ds.Stats[combat.ATKP] += atkmod

		return false
	}, "prototype-crescent", combat.PostSnapshot)
}
