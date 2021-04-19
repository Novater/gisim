package favonius

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("Favonius Lance", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	p := 0.50 + float64(r)*0.1
	cd := 810 - r*90

	//add on crit effect
	s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		//check if char is correct?
		if ds.Actor != c.Name() {
			return false
		}
		if s.StatusActive("Favonius Lance Proc") {
			return false
		}

		if s.Rand.Float64() > p {
			return false
		}
		s.Log.Infof("[%v] favonius lance proc'd", s.Frame())

		s.AddEnergyParticles("Favonius Lance", 3, combat.NoElement, 150) //90 to generate, 60 to get it

		s.Status["Favonius Lance Proc"] = s.F + cd
		return false
	}, "favonius-sword-lance", combat.OnCritDamage)

}
