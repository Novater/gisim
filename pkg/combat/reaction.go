package combat

import (
	"encoding/json"
	"log"
	"math"
)

type reaction struct {
	DidReact bool
	Type     ReactionType
	Next     []aura
}

type aura struct {
	Ele   EleType
	Rate  string
	Gauge float64 //gauge remaining
}

type ReactionType string

const (
	Overload       ReactionType = "overload"
	Superconduct   ReactionType = "superconduct"
	Freeze         ReactionType = "freeze"
	Melt           ReactionType = "melt"
	Vaporize       ReactionType = "vaporize"
	Crystallize    ReactionType = "crystallize"
	Swirl          ReactionType = "swirl"
	ElectroCharged ReactionType = "electrocharged"
	Shatter        ReactionType = "shatter"
	NoReaction     ReactionType = ""
)

//split this out as a separate function so we can call it if we need to apply EC or Burning damage
func (s *Sim) applyReactionDamage(ds Snapshot) float64 {
	var mult float64
	var t EleType
	switch ds.ReactionType {
	case Overload:
		mult = 4
		t = Pyro
	case Superconduct:
		mult = 1
		t = Cryo
	case ElectroCharged:
		mult = 2.4
		t = Electro
	case Shatter:
		mult = 3
		t = Physical
	case Swirl:
		mult = 1.2
		//what element is this??
	default:
		//either not implemented or no dmg
		return 0
	}
	em := ds.Stats[EM]
	//lvl bonus is clearly a line of best fit...
	//need some datamining here
	cl := float64(ds.CharLvl)
	lvlm := 0.0002325*cl*cl*cl + 0.05547*cl*cl - 0.2523*cl + 14.74
	res := ds.TargetRes[t] + ds.ResMod[t]
	resmod := 1 - res/2
	if res >= 0 && res < 0.75 {
		resmod = 1 - res
	} else if res > 0.75 {
		resmod = 1 / (4*res + 1)
	}
	s.Log.Debugw("\t\treact dmg", "em", em, "lvl", cl, "lvl mod", lvlm, "type", ds.ReactionType, "ele", t, "mult", mult, "res", res, "res mod", resmod, "bonus", ds.ReactBonus)

	damage := mult * (1 + ((6.66 * em) / (1400 + em)) + ds.ReactBonus) * lvlm * resmod
	s.Log.Debugf("[%v] %v (%v) caused %v, dealt %v damage", s.Frame(), ds.CharName, ds.Abil, ds.ReactionType, damage)
	s.Target.Damage += damage

	return damage
}

//gaugeMul returns multiplier in seconds for each rate type
func gaugeMul(rate string) float64 {
	switch rate {
	case "A":
		return 9.5
	case "B":
		return 6
	case "C":
		return 4.25
	}
	return 0
}

//checkReact checks if a reaction has occured. for now it's only handling what happens when target has 1
//element only and therefore only one reaction.
func (s *Sim) checkReact(ds Snapshot) reaction {
	target := s.Target
	result := reaction{}
	//if target has no aura then we apply the current aura and carry on with damage calc
	//since no other aura, no reaction will occur
	if len(target.Auras) == 0 {
		result.Next = append(result.Next, aura{
			Ele:   ds.Element,
			Gauge: ds.AuraGauge,
			Rate:  ds.AuraDecayRate,
		})
		s.Log.Debugf("\t%v applied (new). GU: %v Decay: %v", ds.Element, ds.AuraGauge, ds.AuraDecayRate)
		return result
	}

	//single reaction
	if len(target.Auras) == 1 {
		a := target.Auras[0]
		r := reactType(a.Ele, ds.Element)

		//same element, refill
		if r == NoReaction {
			g := math.Max(ds.AuraGauge, a.Gauge)
			result.Next = append(result.Next, aura{
				Ele:   ds.Element,
				Gauge: g,
				Rate:  a.Rate,
			})
			s.Log.Debugf("\t%v refreshed. old gu: %v new gu: %v. rate: %v", ds.Element, a.Gauge, g, a.Rate)
			return result
		}

		result.DidReact = true
		result.Type = r

		//melt
		if r == Melt {
			mult := 0.625
			if ds.Element == Pyro {
				mult = 2.5
			}
			g := ds.AuraGauge * mult
			s.Log.Debugw("\taura applied", "src ele", ds.Element, "src gu", ds.AuraGauge, "t ele", a.Ele, "t gu", a.Gauge, "red", g, "rem", a.Gauge-g)
			//if reduction > a.gauge, remove it completely
			if g < a.Gauge {
				result.Next = append(result.Next, aura{
					Ele:   a.Ele,
					Gauge: a.Gauge - g,
					Rate:  a.Rate,
				})
			}
			return result
		}

		//vaporize
		if r == Vaporize {
			mult := 0.625
			if ds.Element == Hydro {
				mult = 2.5
			}
			g := ds.AuraGauge * mult
			s.Log.Debugw("\taura applied", "src ele", ds.Element, "src gu", ds.AuraGauge, "t ele", a.Ele, "t gu", a.Gauge, "red", g, "rem", a.Gauge-g)
			//if reduction > a.gauge, remove it completely
			if g < a.Gauge {
				result.Next = append(result.Next, aura{
					Ele:   a.Ele,
					Gauge: a.Gauge - g,
					Rate:  a.Rate,
				})
			}
			return result
		}

		//overload
		if r == Overload || r == Superconduct {
			g := ds.AuraGauge * 1.25
			s.Log.Debugw("\taura applied", "src ele", ds.Element, "src gu", ds.AuraGauge, "t ele", a.Ele, "t gu", a.Gauge, "red", g, "rem", a.Gauge-g)
			//if reduction > a.gauge, remove it completely
			if g < a.Gauge {
				result.Next = append(result.Next, aura{
					Ele:   a.Ele,
					Gauge: a.Gauge - g,
					Rate:  a.Rate,
				})
			}
			return result
		}

		if r == ElectroCharged {
			//here we simply re build the result starting with electro
			//then hydro. no tax is applied
			if ds.Element == Electro {
				result.Next = []aura{
					{
						Ele:   ds.Element,
						Gauge: ds.AuraGauge,
						Rate:  ds.AuraDecayRate,
					},
					{
						Ele:   a.Ele,
						Gauge: a.Gauge,
						Rate:  a.Rate,
					},
				}
			} else {
				result.Next = []aura{
					{
						Ele:   a.Ele,
						Gauge: a.Gauge,
						Rate:  a.Rate,
					},
					{
						Ele:   ds.Element,
						Gauge: ds.AuraGauge,
						Rate:  ds.AuraDecayRate,
					},
				}
			}
			//add the profile
			j, err := json.Marshal(ds)
			if err != nil {
				s.Log.Panicw("checkReact marshal source", "err", err)
			}
			s.Target.ecTrigger = j
			return result
		}

		if r == Freeze {
			g := ds.AuraGauge * 1.25
			s.Log.Debugw("\taura applied", "src ele", ds.Element, "src gu", ds.AuraGauge, "t ele", a.Ele, "t gu", a.Gauge, "red", g, "rem", a.Gauge-g)
			//if reduction > a.gauge, remove it completely
			if g < a.Gauge {
				result.Next = append(result.Next, aura{
					Ele:   a.Ele,
					Gauge: a.Gauge - g,
					Rate:  a.Rate,
				})
			}
			return result
		}
	}

	if len(target.Auras) == 2 {
		//the first should either be electro for EC
		//or cryo for frozen
		return result
	}

	if len(target.Auras) > 2 {
		log.Panicf("unexpected error, target has more than 2 auras active: %v", target.Auras)
	}

	//if target has more than one aura then it gets complicated....
	//this should only happen on targets with electrocharged
	//since electro will only result in transformative, that should just trigger a separate damage calc
	//the main ds still carries through to the hydro portion
	return result
}

//EleType is a string representing an element i.e. HYDRO/PYRO/etc...
type EleType string

//ElementType should be pryo, Hydro, Cryo, Electro, Geo, Anemo and maybe dendro
const (
	Pyro         EleType = "pyro"
	Hydro        EleType = "hydro"
	Cryo         EleType = "cryo"
	Electro      EleType = "electro"
	Geo          EleType = "geo"
	Anemo        EleType = "anemo"
	Dendro       EleType = "dendro"
	Physical     EleType = "physical"
	Frozen       EleType = "frozen"
	NonElemental EleType = "non-elemental"
)

func reactType(a EleType, b EleType) ReactionType {
	if a == b {
		return NoReaction
	}
	switch {
	//overload
	case a == Pyro && b == Electro:
		fallthrough
	case a == Electro && b == Pyro:
		return Overload
	//superconduct
	case a == Electro && b == Cryo:
		fallthrough
	case a == Cryo && b == Electro:
		return Superconduct
	//freeze
	case a == Cryo && b == Hydro:
		fallthrough
	case a == Hydro && b == Cryo:
		return Freeze
	//melt
	case a == Pyro && b == Cryo:
		fallthrough
	case a == Cryo && b == Pyro:
		return Melt
	//vape
	case a == Pyro && b == Hydro:
		fallthrough
	case a == Hydro && b == Pyro:
		return Vaporize
	//electrocharged
	case a == Electro && b == Hydro:
		fallthrough
	case a == Hydro && b == Electro:
		return ElectroCharged
		//special reactions??
		//crystal
	case a == Geo || b == Geo:
		return Crystallize
	//swirl
	case a == Anemo || b == Anemo:
		return Swirl
	}
	return NoReaction
}
