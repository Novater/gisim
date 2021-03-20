package combat

import (
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

	//tracking
	Auras  map[EleType]aura
	Status map[string]int //countdown to how long status last

	//stats
	Damage float64 //total damage received
}

type aura struct {
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
	for k, v := range e.Auras {
		//decay first, then delete if < -1
		v.gauge -= 1 / (gaugeMul(v.rate) * 60)
		if v.gauge < 0 {
			s.Log.Debugf("[%v] aura %v expired", s.Frame(), k)
			delete(e.Auras, k)
		}
		e.Auras[k] = v
	}
}

//ApplyDamage applies damage to the target given a snapshot
func (s *Sim) ApplyDamage(ds Snapshot) float64 {

	var damage float64

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
		willReact, react, next := s.checkReact(ds)
		if willReact {
			ds.WillReact = true
			ds.ReactionType = react
			//handle pre reaction
			for k, f := range s.effects[PreReaction] {
				if f(&ds) {
					s.Log.Debugf("[%v] effect (pre reaction) %v expired", s.Frame(), k)
					delete(s.effects[PreReaction], k)
				}
			}
			//either adjust damage snap, adjust stats, or add effect to deal damage after initial damage
			switch react {
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
		target.Auras = next
	}

	//for each reaction damage to occur -> call any pre reaction hooks
	//we can have multiple reaction so snapshot should be made a copy
	//of for each

	//determine type of reaction

	damage = calcDmg(ds, s.Log)
	s.Target.Damage += damage

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

	return damage
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

//checkReact checks if a reaction has occured. for now it's only handling what happens when target has 1
//element only and therefore only one reaction.
func (s *Sim) checkReact(ds Snapshot) (bool, ReactionType, map[EleType]aura) {
	target := s.Target
	next := make(map[EleType]aura)
	//if target has no aura then we apply the current aura and carry on with damage calc
	//since no other aura, no reaction will occur
	if len(target.Auras) == 0 {
		next[ds.Element] = aura{
			gauge: ds.AuraGauge,
			rate:  ds.AuraDecayRate,
		}
		s.Log.Debugf("\t%v applied (new). GU: %v Decay: %v", ds.Element, ds.AuraGauge, ds.AuraDecayRate)
		return false, NoReaction, next
	}

	//for now let's deal with single aura reactions (cause it's simpler)
	if len(target.Auras) == 1 {
		var ele EleType
		var a aura
		var react ReactionType = NoReaction
		for k, v := range target.Auras {
			ele = k
			a = v
		}
		//since element already exist, here we only top up the gauge to the extent the new gauge <= source
		if ele == ds.Element {
			g := math.Max(ds.AuraGauge, a.gauge)
			next[ds.Element] = aura{
				gauge: g,
				rate:  a.rate,
			}
			s.Log.Debugf("\t%v refreshed. old gu: %v new gu: %v. rate: %v", ds.Element, a.gauge, g, a.rate)
			return false, react, next
		}

		switch {
		//these following reactions are transformative so we calculate separate damage, update gauge, and add to total
		//overload
		case ele == Pyro && ds.Element == Electro:
			s.Log.Debugf("\t%v caused overload", ds.Element)
		case ele == Electro && ds.Element == Pyro:
		//superconduct
		case ele == Electro && ds.Element == Cryo:
		case ele == Cryo && ds.Element == Electro:
		case ele == Frozen && ds.Element == Electro:
		//freeze
		//Once freeze is triggered, an enemy will be afflicted by a frozen aura. This aura hides the hydro aura, and any elements
		//applied to a frozen enemy will react with cryo. Removing the frozen aura, either through melt or shatter, will also remove
		//cryo and expose the hydro aura, allowing any elemental sources to react with hydro again.
		//this is a pain to implement.... but so far looks like only childe will have this problem
		//as everyone else applies only 1A of hydro
		//there's a 1.25 multiplier to the source element

		case ele == Cryo && ds.Element == Hydro:
		case ele == Hydro && ds.Element == Cryo:
		//these following reactions are multipliers so we just modify snapshot and update the gauges
		//melt
		case ele == Pyro && ds.Element == Cryo:
			s.Log.Debugf("\t%v caused (weak) melt", ds.Element)
			//weak melt, source unit x 0.625, subtracted from target unit
			//IS IT ROUNDED??? probably not??
			g := ds.AuraGauge * 0.625
			s.Log.Debugw("\taura applied", "src ele", ds.Element, "src gu", ds.AuraGauge, "t ele", ele, "t gu", a.gauge, "red", g, "rem", a.gauge-g)
			//if reduction > a.gauge, remove it completely
			if g < a.gauge {
				next[ele] = aura{
					gauge: a.gauge - g,
					rate:  a.rate,
				}
			}
			react = Melt
		case ele == Cryo && ds.Element == Pyro:
			s.Log.Debugf("\t%v caused (strong) melt", ds.Element)
			g := ds.AuraGauge * 2.5
			s.Log.Debugw("\taura applied", "src ele", ds.Element, "src gu", ds.AuraGauge, "t ele", ele, "t gu", a.gauge, "red", g, "rem", a.gauge-g)
			//if reduction > a.gauge, remove it completely
			if g < a.gauge {
				next[ele] = aura{
					gauge: a.gauge - g,
					rate:  a.rate,
				}
			}
			react = Melt
		case ele == Frozen && ds.Element == Pyro:
			//NOT IMPLEMENTED, HOW TO DEAL WITH HYDRO REMAINING??
		//vape
		case ele == Pyro && ds.Element == Hydro:
			s.Log.Debugf("\t%v caused (strong) vaporize", ds.Element)
			g := ds.AuraGauge * 2.5
			s.Log.Debugw("\taura applied", "src ele", ds.Element, "src gu", ds.AuraGauge, "t ele", ele, "t gu", a.gauge, "red", g, "rem", a.gauge-g)
			//if reduction > a.gauge, remove it completely
			if g < a.gauge {
				next[ele] = aura{
					gauge: a.gauge - g,
					rate:  a.rate,
				}
			}
			react = Vaporize
		case ele == Hydro && ds.Element == Pyro:
			s.Log.Debugf("\t%v caused (weak) vaporize", ds.Element)
			g := ds.AuraGauge * 0.625
			s.Log.Debugw("\taura applied", "src ele", ds.Element, "src gu", ds.AuraGauge, "t ele", ele, "t gu", a.gauge, "red", g, "rem", a.gauge-g)
			//if reduction > a.gauge, remove it completely
			if g < a.gauge {
				next[ele] = aura{
					gauge: a.gauge - g,
					rate:  a.rate,
				}
			}
			react = Vaporize
		//special reactions??
		//crystal
		case ele != Geo && ds.Element == Geo:
			//ONLY GU IMPLEMENTED FOR NOW, NO DAMAGE??
		//swirl
		case ele != Anemo && ds.Element == Anemo:
		//electrocharged is in it's own class
		case ele == Electro && ds.Element == Hydro:
		case ele == Hydro && ds.Element == Electro:
			//ec seem to trigger every 1 second and on expiry, with a 0.5 second ICD of sort?
			//each tick consumes .4 gu?
			//seems kinda complicated? the underlying code must be simpler than this...
			//i wonder if electro + hydro just maintains it's aura and keeps ticking every 1s
			//and any reapplication just adds to the existing GU
			//then any reactions react to each aura separately and then uses up more GU
			//that would seem like a fairly simple way to implement and explain what we're seeing
			//would also apply to future dendro reactions
		}

		return true, react, next
	}

	//if target has more than one aura then it gets complicated....
	//this should only happen on targets with electrocharged
	//since electro will only result in transformative, that should just trigger a separate damage calc
	//the main ds still carries through to the hydro portion
	return false, NoReaction, next
}

func calcDmg(ds Snapshot, log *zap.SugaredLogger) float64 {

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
	}

	return damage
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
	CharName string     //name of the character triggering the damage
	Abil     string     //name of ability triggering the damage
	AbilType ActionType //type of ability triggering the damage

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
