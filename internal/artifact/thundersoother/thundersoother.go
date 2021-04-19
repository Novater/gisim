package thundersoother

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("Thundersoother", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	//make use of closure to keep track of how many hits
	if count >= 2 {
		//???
		s.Log.Warnf("Thundersoother 2 pc not implemented - no character damage taken")
	}
	//after skill, increase normal and charged
	if count >= 4 {
		s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
			if ds.Actor != c.Name() {
				return false
			}
			//only if target affected by pyro
			if s.TargetAura.E() == combat.Electro {
				s.Log.Debugf("\tapplying thundersoother 4pc buff on electro target")
				ds.DmgBonus += 0.35
			}

			return false
		}, "thundersoother 4pc", combat.PreDamageHook)
	}
	//add flat stat to char
}
