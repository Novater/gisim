package combat

import (
	"encoding/json"
	"log"
)

type Aura interface {
	React()
	Tick() bool //remove if true
}

type Element struct {
	Type          EleType
	MaxDurability float64
	Durability    float64
	Expiry        int //when the aura is gone, use this instead of ticks
	Base          int
	Units         int
}

//react with next element, returning the resultant element
func (e *Element) React(next Element, s *Sim) Element {
	//reaction damage can be queued up in the next frame

	//how to deal with multiplier? set a flag on the sim?
	//flag to be cleared right after damage, should be ok
	//since react is only called for apply damage
	return Element{}
}

func (e *Element) Tick(s *Sim) bool {
	return e.Expiry < s.F
}

func (e *Element) reactType(n EleType) ReactionType {
	return ""
}

type Reaction struct {
	Next Aura //the resultant aura, can be nil?
}

type reaction struct {
	DidReact bool
	Type     ReactionType
	Types    []ReactionType
	Next     []aura
}

type aura struct {
	Ele      EleType
	Base     int64
	Units    int64
	Duration int64
}

func EleToDmgP(e EleType) StatType {
	switch e {
	case Anemo:
		return AnemoP
	case Cryo:
		return CryoP
	case Electro:
		return ElectroP
	case Geo:
		return GeoP
	case Hydro:
		return HydroP
	case Pyro:
		return PyroP
	case Dendro:
		return DendroP
	case Physical:
		return PhyP
	}
	return -1
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
	FreezeExtend   ReactionType = "FreezeExtend"
)

//split this out as a separate function so we can call it if we need to apply EC or Burning damage
func (s *Sim) applyReactionDamage(ds Snapshot, target Enemy) float64 {
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

	res := target.Resist(s.Log)[ds.Element]
	resmod := 1 - res/2
	if res >= 0 && res < 0.75 {
		resmod = 1 - res
	} else if res > 0.75 {
		resmod = 1 / (4*res + 1)
	}
	s.Log.Debugw("\t\treact dmg", "em", em, "lvl", cl, "lvl mod", lvlm, "type", ds.ReactionType, "ele", t, "mult", mult, "res", res, "res mod", resmod, "bonus", ds.ReactBonus)

	damage := mult * (1 + ((6.66 * em) / (1400 + em)) + ds.ReactBonus) * lvlm * resmod
	s.Log.Infof("[%v] %v (%v) caused %v, dealt %v damage", s.Frame(), ds.CharName, ds.Abil, ds.ReactionType, damage)
	s.Target.Damage += damage
	s.Target.DamageDetails[ds.CharName][string(ds.ReactionType)] += damage

	return damage
}

func max(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}
func min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
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
			Ele:      ds.Element,
			Duration: ds.AuraBase * ds.AuraUnits,
			Units:    ds.AuraUnits,
			Base:     ds.AuraBase,
		})
		s.Log.Debugf("\t%v applied (new). Base: %v Units: %v Duration: %v", ds.Element, ds.AuraBase, ds.AuraUnits, ds.AuraUnits*ds.AuraBase)
		return result
	}

	//single reaction
	if len(target.Auras) == 1 {
		a := target.Auras[0]
		//check for shatter first
		if a.Ele == Frozen && ds.IsHeavyAttack {
			result.DidReact = true
			result.Type = Shatter
			//the next aura is just the remaining aura but as water
			result.Next = []aura{
				{
					Ele:      Hydro,
					Duration: a.Duration,
					Base:     a.Base,
					Units:    a.Units,
				},
			}
			return result
		}
		//otherwise find the type
		r := reactType(a.Ele, ds.Element)
		//same element, refill
		if r == NoReaction {
			g := max(min(ds.AuraBase*ds.AuraUnits, a.Base*a.Units), a.Duration)

			result.Next = append(result.Next, aura{
				Ele:      ds.Element,
				Duration: g,
				Units:    a.Units,
				Base:     a.Base,
			})
			s.Log.Debugf("\t%v refreshed. base: %v old duration: %v new duration: %v", ds.Element, a.Base, a.Duration, g)
			return result
		}

		//TODO this isn't quite right, should add cryo aura underneath this
		//but need more testing first
		if r == FreezeExtend {
			s.Log.Warnf("\textension of cryo/hydro while affected by Freeze not implemented. No aura applied")
			result.Next = append(result.Next, a)
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
			left := a.Duration - int64(mult*float64(ds.AuraBase*ds.AuraUnits)*(float64(a.Base)/float64(ds.AuraBase)))

			// 2 - 0.624 * 1 = 1.375
			// 1.375 * 360 = 495

			// 720 - 0.625 * 570 * 360/720

			s.Log.Debugw("\taura applied", "src ele", ds.Element, "src units", ds.AuraBase, "t ele", a.Ele, "t units", a.Duration, "rem", left)
			//if reduction > a.gauge, remove it completely
			if left > 0 {
				result.Next = append(result.Next, aura{
					Ele:      a.Ele,
					Duration: left,
					Units:    a.Units,
					Base:     a.Base,
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
			left := a.Duration - int64(mult*float64(ds.AuraBase*ds.AuraUnits)*(float64(a.Base)/float64(ds.AuraBase)))
			s.Log.Debugw("\taura applied", "src ele", ds.Element, "src units", ds.AuraBase, "t ele", a.Ele, "t units", a.Duration, "rem", left)

			//if reduction > a.gauge, remove it completely
			if left > 0 {
				result.Next = append(result.Next, aura{
					Ele:      a.Ele,
					Duration: left,
					Units:    a.Units,
					Base:     a.Base,
				})
			}
			return result
		}

		//overload
		if r == Overload || r == Superconduct {
			left := a.Duration - int64(1.25*float64(ds.AuraBase*ds.AuraUnits)*(float64(a.Base)/float64(ds.AuraBase)))
			s.Log.Debugw("\taura applied", "src ele", ds.Element, "src units", ds.AuraBase, "t ele", a.Ele, "t units", a.Duration, "rem", left)

			//if reduction > a.gauge, remove it completely
			if left > 0 {
				result.Next = append(result.Next, aura{
					Ele:      a.Ele,
					Duration: left,
					Units:    a.Units,
					Base:     a.Base,
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
						Ele:      ds.Element,
						Duration: ds.AuraBase * ds.AuraUnits,
						Units:    ds.AuraUnits,
						Base:     ds.AuraBase,
					},
					{
						Ele:      a.Ele,
						Duration: a.Duration,
						Units:    a.Units,
						Base:     a.Base,
					},
				}
			} else {
				result.Next = []aura{
					{
						Ele:      a.Ele,
						Duration: a.Duration,
						Units:    a.Units,
						Base:     a.Base,
					},
					{
						Ele:      ds.Element,
						Duration: ds.AuraBase * ds.AuraUnits,
						Units:    ds.AuraUnits,
						Base:     ds.AuraBase,
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

		//for now replace freeze with it's own aura
		if r == Freeze {
			//freeze duration can be calculated as -0.3x + 180 if initial application is 1A (weak) or -0.425x + 330 if initial application is 2B (stronger)
			//where x is duration passed
			var left int64
			if a.Base == WeakAuraBase {
				left = int64(-0.3*float64(a.Base*a.Units-a.Duration)) + 180
			} else {
				left = int64(-0.425*float64(a.Base*a.Units-a.Duration)) + 330
				//if the applying aura is weaker, cap it
				if ds.AuraBase == WeakAuraBase && left > 210 {
					left = 210
				}
			}
			s.Log.Debugw("\tis frozen", "src ele", ds.Element, "src base", ds.AuraBase, "t ele", a.Ele, "t base", a.Base, "rem", left)
			result.Next = []aura{
				{
					Ele:      Frozen,
					Duration: left,
					Units:    a.Units,
					Base:     a.Base,
				},
			}
			//set frozen to be true
			s.Target.IsFrozen = true
			return result
		}
	}

	if len(target.Auras) == 2 {
		a := target.Auras[0]
		//for now the first should only be electro for electro charged; otherwise
		//we wouldn't know what to do
		if a.Ele != Electro {
			s.Log.Warnw("multi aura, first element not electro. not implemented", "aura", target.Auras)
			return result
		}

		//implement each reaction separately

		switch ds.Element {
		case Pyro:
			//overload + vaporize
		case Cryo:
			//superconduct no freeze no aura left
			result.DidReact = true
			result.Type = Superconduct
			return result
		case Electro:
			// don't know what to do here
			s.Log.Warnw("reapply electro to EC, not implemented", "aura", target.Auras)
			result.Next = target.Auras
			return result
		case Hydro:
			// don't know what to do here
			s.Log.Warnw("reapply hydro to EC, not implemented", "aura", target.Auras)
			result.Next = target.Auras
			return result
		case Geo:
			//crystallize only electro and remove
		case Anemo:
			//????
		}
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

const (
	WeakAuraBase   int64 = 570
	MedAuraBase    int64 = 360
	StrongAuraBase int64 = 225
)

func reactType(a EleType, b EleType) ReactionType {
	if a == b {
		return NoReaction
	}
	switch {
	case a == Frozen || b == Frozen:
		return FreezeExtend
	//overload
	case a == Pyro && b == Electro:
		fallthrough
	case a == Electro && b == Pyro:
		return Overload
	//superconduct
	case a == Frozen && b == Electro:
		fallthrough
	case a == Electro && b == Frozen:
		fallthrough
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
	case a == Frozen && b == Pyro:
		fallthrough
	case a == Pyro && b == Frozen:
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
