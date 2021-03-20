package combat

import (
	"math/rand"

	"go.uber.org/zap"
)

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

	var reactDamage float64
	//in general, transformative reaction does not change the snapshot
	//they will only trigger a sep damage calc
	if ds.ApplyAura {
		next := make(map[EleType]aura)
		//if target has no aura then we apply the current aura and carry on with damage calc
		//since no other aura, no reaction will occur
		if len(target.Auras) == 0 {
			dur := auraDur(ds.AuraUnit, ds.AuraGauge)
			next[ds.Element] = aura{
				gauge:    ds.AuraGauge,
				unit:     ds.AuraUnit,
				duration: dur,
			}
			s.Log.Debugf("\t%v applied (new). unit: %v. duration: %v", ds.Element, ds.AuraUnit, dur)

		} else {
			//if target has more than one aura then it gets complicated....
			//this should only happen on targets with electrocharged
			//since electro will only result in transformative, that should just trigger a separate damage calc
			//the main ds still carries through to the hydro portion
			for ele, a := range target.Auras {
				switch {
				//no reaction
				case ele == ds.Element:
					dur := auraDur(a.unit, ds.AuraGauge)
					next[ds.Element] = aura{
						gauge:    ds.AuraGauge,
						unit:     a.unit,
						duration: dur,
					}
					//refresh duration
					s.Log.Debugf("\t%v refreshed. unit: %v. new duration: %v", ds.Element, a.unit, dur)
				//these following reactions are transformative so we calculate separate damage, update gauge, and add to total
				//overload
				case ele == Pyro && ds.Element == Electro:
					for k, f := range s.effects[PreOverload] {
						if f(&ds) {
							s.Log.Debugf("[%v] effect (pre overload) %v expired", s.Frame(), k)
							delete(s.effects[PreOverload], k)
						}
					}
					reactDamage += calcOverload(ds)
				case ele == Electro && ds.Element == Pyro:
					for k, f := range s.effects[PreOverload] {
						if f(&ds) {
							s.Log.Debugf("[%v] effect (pre overload) %v expired", s.Frame(), k)
							delete(s.effects[PreOverload], k)
						}
					}
				//superconduct
				case ele == Electro && ds.Element == Cryo:
				case ele == Cryo && ds.Element == Electro:
				case ele == Frozen && ds.Element == Electro:
				//freeze
				case ele == Cryo && ds.Element == Hydro:
				case ele == Hydro && ds.Element == Cryo:
				//these following reactions are multipliers so we just modify snapshot and update the gauges
				//melt
				case ele == Pyro && ds.Element == Cryo:
				case ele == Cryo && ds.Element == Pyro:
				case ele == Frozen && ds.Element == Pyro:
				//vape
				case ele == Pyro && ds.Element == Hydro:
				case ele == Hydro && ds.Element == Pyro:
				//special reactions??
				//crystal
				case ele != Geo && ds.Element == Geo:
				//swirl
				case ele != Anemo && ds.Element == Anemo:
				//electrocharged is in it's own class
				case ele == Electro && ds.Element == Hydro:
				case ele == Hydro && ds.Element == Electro:
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
	damage += reactDamage
	s.Target.Damage += damage

	for k, f := range s.effects[PostDamageHook] {
		if f(&ds) {
			s.Log.Debugf("[%v] effect (pre damage) %v expired", s.Frame(), k)
			delete(s.effects[PostDamageHook], k)
		}
	}

	return damage
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
				s.Log.Debugw("\tapply aura", "aura", a, "existing ele", ele, "next ele", ds.Element)
			} else {
				s.Log.Debugf("\tnot implemented!!!")
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
			s.Log.Debugf("\t%v refreshed. unit: %v. new duration: %v", ds.Element, a.unit, next.duration)
			e.Auras[ds.Element] = next
		} else {
			//apply reaction
			//The length of the freeze is based on the lowest remaining duration of the two elements applied.
			s.Log.Debugf("\tnot implemented!!!")
		}
	} else {
		next := aura{
			gauge:    ds.AuraGauge,
			unit:     ds.AuraUnit,
			duration: auraDur(ds.AuraUnit, ds.AuraGauge),
		}
		s.Log.Debugf("%v applied (new). unit: %v. duration: %v", ds.Element, next.unit, next.duration)
		e.Auras[ds.Element] = next
	}
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
	log.Debugw("\t\tcalc", "def mod", defmod, "res", res, "res mod", resmod, "pre crit damage", damage)

	//check if crit
	if rand.Float64() <= ds.Stats[CR] || ds.HitWeakPoint {
		log.Debugf("\t\tdamage is crit!")
		damage = damage * (1 + ds.Stats[CD])
	}

	return damage
}

func calcOverload(ds Snapshot) float64 {
	//4 * ( 1 + 6.66 * EM / (1400+ EM) + react bonus) * lvl mult * res mult

	em := ds.Stats[EM]
	//lvl bonus is clearly a line of best fit...
	//need some datamining here
	cl := float64(ds.CharLvl)
	lvlm := 0.0002325*cl*cl*cl + 0.05547*cl*cl - 0.2523*cl + 14.74
	//resist is always fire resist b/c overload does fire dmg
	res := ds.TargetRes[Pyro] + ds.ResMod[Pyro]
	resmod := 1 - res/2
	if res >= 0 && res < 0.75 {
		resmod = 1 - res
	} else if res > 0.75 {
		resmod = 1 / (4*res + 1)
	}

	damage := 4 * (1 + ((6.66 * em) / (1400 + em)) + ds.ReactBonus) * lvlm * resmod

	return damage
}

func calcSuperconductDamage(ds Snapshot) float64 {
	return 0
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
	ApplyAura bool    //if aura should be applied; false if under ICD
	AuraGauge float64 //1 2 or 4
	AuraUnit  string  //A, B, or C

	//these are calculated fields
	WillReact bool //true if this will react

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
