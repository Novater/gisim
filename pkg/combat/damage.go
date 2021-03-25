package combat

import (
	"math/rand"

	"go.uber.org/zap"
)

//ApplyDamage applies damage to the target given a snapshot
func (s *Sim) ApplyDamage(ds Snapshot) float64 {

	target := s.Target

	ds.TargetLvl = target.Level
	ds.TargetRes = target.Resist

	s.Log.Debugf("[%v] %v - %v triggered dmg", s.Frame(), ds.CharName, ds.Abil)
	s.Log.Debugw("\ttarget", "auras", target.Auras)

	//in general, transformative reaction does not change the snapshot
	//they will only trigger a sep damage calc
	if ds.ApplyAura {
		r := s.checkReact(ds)
		s.Log.Debugw("reaction result", "r", r, "source", ds.Element)
		if r.DidReact {
			ds.WillReact = true
			ds.ReactionType = r.Type
			//handle pre reaction
			for k, f := range s.combatHooks[PreReaction] {
				if f(&ds) {
					s.Log.Debugf("[%v] effect (pre reaction) %v expired", s.Frame(), k)
					delete(s.combatHooks[PreReaction], k)
				}
			}

			//either adjust damage snap, adjust stats, or add effect to deal damage after initial damage
			switch r.Type {
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
			case Superconduct:
				s.Status["Superconduct"] = 12 * 60 //add a debuff for superconduct
			case ElectroCharged:
				target.Status["electrocharge icd"] = 60
			}
		}
		target.Auras = r.Next
	}

	for k, f := range s.combatHooks[PreDamageHook] {
		s.Log.Debugf("\trunning pre damage hook: %v", k)
		if f(&ds) {
			s.Log.Debugf("[%v] effect (pre damage) %v expired", s.Frame(), k)
			delete(s.combatHooks[PreDamageHook], k)
		}
	}

	//for each reaction damage to occur -> call any pre reaction hooks
	//we can have multiple reaction so snapshot should be made a copy
	//of for each

	//determine type of reaction

	dr := calcDmg(ds, s.Log)
	s.Target.Damage += dr.damage
	s.Target.DamageDetails[ds.CharName][ds.Abil] += dr.damage

	if dr.isCrit {
		for k, f := range s.combatHooks[OnCritDamage] {
			s.Log.Debugf("\trunning on crit hook: %v", k)
			if f(&ds) {
				s.Log.Debugf("[%v] effect (on crit dmg) %v expired", s.Frame(), k)
				delete(s.combatHooks[PostDamageHook], k)
			}
		}
	}

	for k, f := range s.combatHooks[PostDamageHook] {
		s.Log.Debugf("\trunning post damage hook: %v", k)
		if f(&ds) {
			s.Log.Debugf("[%v] effect (pre damage) %v expired", s.Frame(), k)
			delete(s.combatHooks[PostDamageHook], k)
		}
	}

	//apply reaction damage now! not sure if this timing is right though; maybe we can add this to the next frame as a tick instead?
	if ds.WillReact {
		s.applyReactionDamage(ds)
		for k, f := range s.combatHooks[PostReaction] {
			s.Log.Debugf("\trunning post reaction hook: %v", k)
			if f(&ds) {
				s.Log.Debugf("[%v] effect (pre reaction) %v expired", s.Frame(), k)
				delete(s.combatHooks[PostReaction], k)
			}
		}
	}

	return dr.damage
}

type dmgResult struct {
	damage float64
	isCrit bool
}

func calcDmg(ds Snapshot, log *zap.SugaredLogger) dmgResult {

	result := dmgResult{}

	st := EleToDmgP(ds.Element)
	ds.DmgBonus += ds.Stats[st]

	log.Debugw("\t\tcalc", "base atk", ds.BaseAtk, "flat +", ds.Stats[ATK], "% +", ds.Stats[ATKP], "ele", st, "ele %", ds.Stats[st], "bonus dmg", ds.DmgBonus, "mul", ds.Mult)
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
	log.Debugw("\t\tcalc", "def mod", defmod, "res", res, "res mod", resmod)
	log.Debugw("\t\tcalc", "pre crit damage", damage, "dmg if crit", damage*(1+ds.Stats[CD]), "melt/vape", ds.IsMeltVape)

	//check melt/vape
	if ds.IsMeltVape {
		//calculate em bonus
		em := ds.Stats[EM]
		emBonus := (2.78 * em) / (1400 + em)
		log.Debugw("\t\tcalc", "react mult", ds.ReactMult, "em", em, "em bonus", emBonus, "react bonus", ds.ReactBonus, "pre react damage", damage)
		damage = damage * (ds.ReactMult * (1 + emBonus + ds.ReactBonus))
		log.Debugw("\t\tcalc", "pre crit (post react) damage", damage, "pre react if crit", damage*(1+ds.Stats[CD]))
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
	ApplyAura     bool  //if aura should be applied; false if under ICD
	AuraBase      int64 //unit base
	AuraUnits     int64 //number of units
	IsHeavyAttack bool

	//these are calculated fields
	WillReact bool //true if this will react
	//these two fields will only work if only reaction vs one element?!
	ReactionType ReactionType
	ReactedTo    EleType //NOT IMPLEMENTED

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
