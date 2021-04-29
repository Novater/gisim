package lavawalker

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("lavawalker", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	//make use of closure to keep track of how many hits
	if count >= 2 {
		//???
		s.Log.Warnf("Lavawalker 2 pc not implemented - no character damage taken")
	}
	//after skill, increase normal and charged
	if count >= 4 {
		s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
			if ds.Actor != c.Name() {
				return false
			}
			//only if target affected by pyro
			if s.TargetAura.E() == combat.Pyro {
				s.Log.Debugf("\tapplying lavawalker 4pc buff on pyro target")
				ds.DmgBonus += 0.35
			}

			return false
		}, "lavawalker 4pc", combat.PreDamageHook)
	}
	//add flat stat to char
}
