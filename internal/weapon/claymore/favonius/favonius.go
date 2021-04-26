package favonius

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("Favonius Greatsword", weapon)
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
		if s.StatusActive("Favonius Greatsword Proc") {
			return false
		}

		if s.Rand.Float64() > p {
			return false
		}
		s.Log.Infof("[%v] favonius greatsword proc'd", s.Frame())

		s.AddEnergyParticles("Favonius Greatsword", 3, combat.NoElement, 150) //90 to generate, 60 to get it

		s.Status["Favonius Greatsword Proc"] = s.F + cd
		return false
	}, "favonius-sword-greatsword", combat.OnCritDamage)

}
