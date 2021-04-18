package favonius

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("Favonius Sword", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	p := 0.6
	cd := 120 * 6
	switch r {
	case 2:
		p = .7
		cd = 105 * 6
	case 3:
		p = .8
		cd = 90 * 6
	case 4:
		p = .9
		cd = 75 * 6
	case 5:
		p = .10
		cd = 60 * 6
	}
	//add on crit effect
	s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		//check if char is correct?
		if ds.Actor != c.Name() {
			return false
		}
		if s.StatusActive("Favonius Sword Proc") {
			return false
		}

		if s.Rand.Float64() > p {
			return false
		}
		s.Log.Infof("[%v] favonius sword proc'd", s.Frame())

		s.AddEnergyParticles("Favonius Sword", 3, combat.NoElement, 150) //90 to generate, 60 to get it

		s.Status["Favonius Sword Proc"] = s.F + cd
		return false
	}, "favonius-sword-proc", combat.OnCritDamage)

}
