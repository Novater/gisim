package combat

import "go.uber.org/zap"

//eleType is a string representing an element i.e. HYDRO/PYRO/etc...
type eleType string

//ElementType should be pryo, Hydro, Cryo, Electro, Geo, Anemo and maybe dendro
const (
	Pyro     eleType = "pyro"
	Hydro    eleType = "hydro"
	Cryo     eleType = "cryo"
	Electro  eleType = "electro"
	Geo      eleType = "geo"
	Anemo    eleType = "anemo"
	Physical eleType = "physical"
	Frozen   eleType = "frozen"
)

type aura struct {
	gauge    float64
	unit     string
	duration int
}

func auraDur(unit string, gauge float64) int {
	switch unit {
	case "A":
		return int(gauge * 9.5 * 60)
	case "B":
		return int(gauge * 6 * 60)
	case "C":
		return int(gauge * 4.25 * 60)
	}
	return 0
}

//applyAura applies an aura to the Unit, can trigger damage for superconduct, electrocharged, etc..
func (s *Sim) applyAura(ds Snapshot) {
	e := s.Target
	//1A = 9.5s (570 frames) per unit, 2B = 6s (360 frames) per unit, 4C = 4.25s (255 frames) per unit
	//loop through existing auras and apply reactions if any
	if len(e.Auras) > 1 {
		//this case should only happen with electro charge where there's 2 aura active at any one point in time
		for ele, a := range e.Auras {
			if ele != ds.Element {
				zap.S().Debugw("apply aura", "aura", a, "existing ele", ele, "next ele", ds.Element)
			} else {
				zap.S().Debugf("not implemented!!!")
			}
		}
	} else if len(e.Auras) == 1 {
		if a, ok := e.Auras[ds.Element]; ok {
			next := aura{
				gauge:    ds.AuraGauge,
				unit:     a.unit,
				duration: auraDur(a.unit, ds.AuraGauge),
			}
			//refresh duration
			zap.S().Debugf("%v refreshed. unit: %v. new duration: %v", ds.Element, a.unit, next.duration)
			e.Auras[ds.Element] = next
		} else {
			//apply reaction
			//The length of the freeze is based on the lowest remaining duration of the two elements applied.
			zap.S().Debugf("not implemented!!!")
		}
	} else {
		next := aura{
			gauge:    ds.AuraGauge,
			unit:     ds.AuraUnit,
			duration: auraDur(ds.AuraUnit, ds.AuraGauge),
		}
		zap.S().Debugf("%v applied (new). unit: %v. duration: %v", ds.Element, next.unit, next.duration)
		e.Auras[ds.Element] = next
	}
}
