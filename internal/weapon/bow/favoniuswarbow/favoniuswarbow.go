package favoniuswarbow

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("Favonius Warbow", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {

	prob := 0.6
	cd := 12 * 60
	switch r {
	case 2:
		prob = .7
		cd = 105 * 6
	case 3:
		prob = .8
		cd = 9 * 60
	case 4:
		prob = .9
		cd = 75 * 6
	case 5:
		prob = 1
		cd = 6 * 60
	}
	//add on crit effect
	s.AddSnapshotHook(func(snap *combat.Snapshot) bool {
		//check if char is correct?
		if snap.CharName != c.Name() {
			return false
		}
		//return if cd is active
		if s.StatusActive("Favonius Warbow Proc") {
			return false
		}

		if s.Rand.Float64() > prob {
			return false
		}
		s.Log.Debugf("\t Favonius Warbox proc triggered")

		s.AddEnergyParticles("Favonius Warbow", 3, combat.NonElemental, 150) //90 to generate, 60 to get it

		s.Status["Favonius Warbow Proc"] = s.F + cd
		return false
	}, "favonius-warbow-proc", combat.OnCritDamage)
}
