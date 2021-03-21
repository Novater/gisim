package combat

import (
	"log"
	"math"
	"math/rand"

	"go.uber.org/zap"
)

//Enemy keeps track of the status of one enemy Enemy
type Enemy struct {
	Level  int64
	Resist map[EleType]float64

	//resist mods
	ResMod map[string]float64

	//auras should be stored in an array
	//there seems to be a priority system on what gets stored first; don't know how it works
	//for now we're only concerned about EC and Freeze; so we'll just hard code though; EC is electro,hydro; Freeze is cryo,hydro
	//in EC case, any additional reaction applied will first react with electro, then hydro
	//	not sure if this is only applicable if first reaction is transformative i.e. overload or superconduct;
	//	^ may make sense b/c additional application of hydro will prob trigger additional reaction? need to test this somehow??
	//  but the code would be the same, i.e hydro gets applied to electro (does nothing), then added to more hydro
	//in frozen's case we just have to hard code that any element will only react with cryo and hydro just dies off why cryo dies off
	//unless we shatter in which case cryo gets removed
	Auras []aura
	//tracking
	Status map[string]int //countdown to how long status last

	//stats
	Damage float64 //total damage received
}

type reaction struct {
	react bool
	t     ReactionType
	next  []aura
}

type aura struct {
	Ele   EleType
	rate  string
	gauge float64 //gauge remaining
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
	NoReaction     ReactionType = ""
)

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

func (e *Enemy) tick(s *Sim) {
	//tick down buffs and debuffs
	for k, v := range e.Status {
		if v == 0 {
			delete(e.Status, k)
		} else {
			e.Status[k]--
		}
	}
	//tick down gauge, reduction is based on the rate of the aura
	//multipliers are A: 9.5, B:6, C:4.25
	//decay per frame (60fps) = 1 unit / (mult * 60)
	var next []aura
	for _, v := range e.Auras {
		//decay first, then delete if < -1
		v.gauge -= 1 / (gaugeMul(v.rate) * 60)
		if v.gauge > 0 {
			next = append(next, v)
		} else {
			s.Log.Debugf("[%v] aura %v expired", s.Frame(), v.Ele)
		}
	}
	e.Auras = next
}

//ApplyDamage applies damage to the target given a snapshot
func (s *Sim) ApplyDamage(ds Snapshot) float64 {

	target := s.Target

	ds.TargetLvl = target.Level
	ds.TargetRes = target.Resist

	for k, f := range s.effects[PreDamageHook] {
		if f(&ds) {
			s.Log.Debugf("[%v] effect (pre damage) %v expired", s.Frame(), k)
			delete(s.effects[PreDamageHook], k)
		}
	}

	s.Log.Debugf("[%v] %v - %v triggered dmg", s.Frame(), ds.CharName, ds.Abil)

	//in general, transformative reaction does not change the snapshot
	//they will only trigger a sep damage calc
	if ds.ApplyAura {
		r := s.checkReact(ds)
		if r.react {
			ds.WillReact = true
			ds.ReactionType = r.t
			//handle pre reaction
			for k, f := range s.effects[PreReaction] {
				if f(&ds) {
					s.Log.Debugf("[%v] effect (pre reaction) %v expired", s.Frame(), k)
					delete(s.effects[PreReaction], k)
				}
			}
			//either adjust damage snap, adjust stats, or add effect to deal damage after initial damage
			switch r.t {
			case Melt:
				if ds.Element == Pyro {
					//tag it for melt
					ds.IsMeltVape = true
					ds.ReactMult = 2.0 //strong reaction
				} else {
					//tag it for melt
					ds.IsMeltVape = true
					ds.ReactMult = 1.5 //weak reaction
				}
			case Vaporize:
				if ds.Element == Pyro {
					//tag it for melt
					ds.IsMeltVape = true
					ds.ReactMult = 1.5 //weak, triggered by pyro
				} else {
					//tag it for melt
					ds.IsMeltVape = true
					ds.ReactMult = 2.0 //strong, triggered by hydro
				}
			}
		}
		target.Auras = r.next
	}

	//for each reaction damage to occur -> call any pre reaction hooks
	//we can have multiple reaction so snapshot should be made a copy
	//of for each

	//determine type of reaction

	dr := calcDmg(ds, s.Log)
	s.Target.Damage += dr.damage

	for k, f := range s.effects[OnCritDamage] {
		if f(&ds) {
			s.Log.Debugf("[%v] effect (on crit dmg) %v expired", s.Frame(), k)
			delete(s.effects[PostDamageHook], k)
		}
	}

	for k, f := range s.effects[PostDamageHook] {
		if f(&ds) {
			s.Log.Debugf("[%v] effect (pre damage) %v expired", s.Frame(), k)
			delete(s.effects[PostDamageHook], k)
		}
	}

	//apply reaction damage now! not sure if this timing is right though; maybe we can add this to the next frame as a tick instead?
	if ds.WillReact {
		s.applyReactionDamage(ds)
		for k, f := range s.effects[PostReaction] {
			if f(&ds) {
				s.Log.Debugf("[%v] effect (pre reaction) %v expired", s.Frame(), k)
				delete(s.effects[PostReaction], k)
			}
		}
	}

	return dr.damage
}

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

//checkReact checks if a reaction has occured. for now it's only handling what happens when target has 1
//element only and therefore only one reaction.
func (s *Sim) checkReact(ds Snapshot) reaction {
	target := s.Target
	result := reaction{}
	//if target has no aura then we apply the current aura and carry on with damage calc
	//since no other aura, no reaction will occur
	if len(target.Auras) == 0 {
		result.next = append(result.next, aura{
			Ele:   ds.Element,
			gauge: ds.AuraGauge,
			rate:  ds.AuraDecayRate,
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
			g := math.Max(ds.AuraGauge, a.gauge)
			result.next = append(result.next, aura{
				Ele:   ds.Element,
				gauge: g,
				rate:  a.rate,
			})
			s.Log.Debugf("\t%v refreshed. old gu: %v new gu: %v. rate: %v", ds.Element, a.gauge, g, a.rate)
			return result
		}

		result.react = true
		result.t = r

		//melt
		if r == Melt {
			mult := 0.625
			if ds.Element == Pyro {
				mult = 2.5
			}
			g := ds.AuraGauge * mult
			s.Log.Debugw("\taura applied", "src ele", ds.Element, "src gu", ds.AuraGauge, "t ele", a.Ele, "t gu", a.gauge, "red", g, "rem", a.gauge-g)
			//if reduction > a.gauge, remove it completely
			if g < a.gauge {
				result.next = append(result.next, aura{
					Ele:   a.Ele,
					gauge: a.gauge - g,
					rate:  a.rate,
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
			s.Log.Debugw("\taura applied", "src ele", ds.Element, "src gu", ds.AuraGauge, "t ele", a.Ele, "t gu", a.gauge, "red", g, "rem", a.gauge-g)
			//if reduction > a.gauge, remove it completely
			if g < a.gauge {
				result.next = append(result.next, aura{
					Ele:   a.Ele,
					gauge: a.gauge - g,
					rate:  a.rate,
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

type dmgResult struct {
	damage float64
	isCrit bool
}

func calcDmg(ds Snapshot, log *zap.SugaredLogger) dmgResult {

	result := dmgResult{}

	st := EleToDmgP(ds.Element)
	ds.DmgBonus += ds.Stats[st]

	log.Debugw("\t\tcalc", "base atk", ds.BaseAtk, "flat +", ds.Stats[ATK], "% +", ds.Stats[ATKP], "bonus dmg", ds.DmgBonus, "mul", ds.Mult)
	//calculate attack or def
	var a float64
	if ds.UseDef {
		a = ds.BaseDef*(1+ds.Stats[DEFP]) + ds.Stats[DEF]
	} else {
		a = ds.BaseAtk*(1+ds.Stats[ATKP]) + ds.Stats[ATK]
	}

	base := ds.Mult*a + ds.FlatDmg
	damage := base * (1 + ds.DmgBonus)

	log.Debugw("\t\tcalc", "total atk", a, "base dmg", base, "dmg + bonus", damage)

	//make sure 0 <= cr <= 1
	if ds.Stats[CR] < 0 {
		ds.Stats[CR] = 0
	}
	if ds.Stats[CR] > 1 {
		ds.Stats[CR] = 1
	}

	log.Debugw("\t\tcalc", "cr", ds.Stats[CR], "cd", ds.Stats[CD], "def adj", ds.DefMod, "res adj", ds.ResMod[ds.Element], "char lvl", ds.CharLvl, "target lvl", ds.TargetLvl, "target res", ds.TargetRes[ds.Element])

	defmod := float64(ds.CharLvl+100) / (float64(ds.CharLvl+100) + float64(ds.TargetLvl+100)*(1-ds.DefMod))
	//apply def mod
	damage = damage * defmod
	//add up the resist mods
	//apply resist mod
	res := ds.TargetRes[ds.Element] + ds.ResMod[ds.Element]
	resmod := 1 - res/2
	if res >= 0 && res < 0.75 {
		resmod = 1 - res
	} else if res > 0.75 {
		resmod = 1 / (4*res + 1)
	}
	damage = damage * resmod

	//apply other multiplier bonus
	if ds.OtherMult > 0 {
		damage = damage * ds.OtherMult
	}
	log.Debugw("\t\tcalc", "def mod", defmod, "res", res, "res mod", resmod, "pre crit damage", damage, "melt/vape", ds.IsMeltVape)

	//check melt/vape
	if ds.IsMeltVape {
		log.Debugw("\t\tcalc", "react mult", ds.ReactMult, "react bonus", ds.ReactBonus, "pre react damage", damage)
		damage = damage * (ds.ReactMult + ds.ReactBonus)
		log.Debugw("\t\tcalc", "pre crit (post react) damage", damage)
	}

	//check if crit
	if rand.Float64() <= ds.Stats[CR] || ds.HitWeakPoint {
		log.Debugf("\t\tdamage is crit!")
		damage = damage * (1 + ds.Stats[CD])
		result.isCrit = true
	}

	result.damage = damage

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

type Snapshot struct {
	CharName    string      //name of the character triggering the damage
	Abil        string      //name of ability triggering the damage
	AbilType    ActionType  //type of ability triggering the damage
	WeaponClass WeaponClass //b.c. Gladiators...

	HitWeakPoint bool

	TargetLvl int64
	TargetRes map[EleType]float64

	Mult      float64 //ability multiplier. could set to 0 from initial Mona dmg
	Element   EleType //element of ability
	UseDef    bool    //default false
	FlatDmg   float64 //flat dmg; so far only zhongli
	OtherMult float64 //so far just for xingqiu C4

	Stats        map[StatType]float64 //total character stats including from artifact, bonuses, etc...
	ExtraStatMod map[StatType]float64
	BaseAtk      float64 //base attack used in calc
	BaseDef      float64 //base def used in calc
	DmgBonus     float64 //total damage bonus, including appropriate ele%, etc..
	CharLvl      int64
	DefMod       float64
	ResMod       map[EleType]float64

	//reaction stuff
	ApplyAura     bool    //if aura should be applied; false if under ICD
	AuraGauge     float64 //1 2 or 4
	AuraDecayRate string  //A, B, or C

	//these are calculated fields
	WillReact    bool //true if this will react
	ReactionType ReactionType

	IsMeltVape bool    //trigger melt/vape
	ReactMult  float64 //reaction multiplier for melt/vape
	ReactBonus float64 //reaction bonus %+ such as witch; should be 0 and only affected by hooks
}

func (s *Snapshot) Clone() Snapshot {
	c := Snapshot{}
	c = *s
	c.ResMod = make(map[EleType]float64)
	c.TargetRes = make(map[EleType]float64)
	c.ExtraStatMod = make(map[StatType]float64)
	return c
}
