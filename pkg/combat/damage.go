package combat

import (
	"math/rand"

	"go.uber.org/zap"
)

//ApplyDamage applies damage to the target given a snapshot
func (s *Sim) ApplyDamage(ds Snapshot) float64 {

	target := s.Target

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
			for k, f := range s.snapshotHooks[PreReaction] {
				if f(&ds) {
					s.Log.Debugf("[%v] effect (pre reaction) %v expired", s.Frame(), k)
					delete(s.snapshotHooks[PreReaction], k)
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

	for k, f := range s.snapshotHooks[PreDamageHook] {
		s.Log.Debugf("\trunning pre damage hook: %v", k)
		if f(&ds) {
			s.Log.Debugf("[%v] effect (pre damage) %v expired", s.Frame(), k)
			delete(s.snapshotHooks[PreDamageHook], k)
		}
	}

	//for each reaction damage to occur -> call any pre reaction hooks
	//we can have multiple reaction so snapshot should be made a copy
	//of for each

	//determine type of reaction

	dr := calcDmg(ds, *s.Target, s.Log)
	s.Target.Damage += dr.damage
	s.Target.DamageDetails[ds.CharName][ds.Abil] += dr.damage

	if dr.isCrit {
		for k, f := range s.snapshotHooks[OnCritDamage] {
			s.Log.Debugf("\trunning on crit hook: %v", k)
			if f(&ds) {
				s.Log.Debugf("[%v] effect (on crit dmg) %v expired", s.Frame(), k)
				delete(s.snapshotHooks[PostDamageHook], k)
			}
		}
	}

	for k, f := range s.snapshotHooks[PostDamageHook] {
		s.Log.Debugf("\trunning post damage hook: %v", k)
		if f(&ds) {
			s.Log.Debugf("[%v] effect (pre damage) %v expired", s.Frame(), k)
			delete(s.snapshotHooks[PostDamageHook], k)
		}
	}

	//apply reaction damage now! not sure if this timing is right though; maybe we can add this to the next frame as a tick instead?
	if ds.WillReact {
		s.applyReactionDamage(ds, *s.Target)
		for k, f := range s.snapshotHooks[PostReaction] {
			s.Log.Debugf("\trunning post reaction hook: %v", k)
			if f(&ds) {
				s.Log.Debugf("[%v] effect (pre reaction) %v expired", s.Frame(), k)
				delete(s.snapshotHooks[PostReaction], k)
			}
		}
	}

	return dr.damage
}

type dmgResult struct {
	damage float64
	isCrit bool
}

func calcDmg(ds Snapshot, target Enemy, log *zap.SugaredLogger) dmgResult {

	result := dmgResult{}

	st := EleToDmgP(ds.Element)
	ds.DmgBonus += ds.Stats[st]

	log.Debugw("\t\tcalc", "base atk", ds.BaseAtk, "flat +", ds.Stats[ATK], "% +", ds.Stats[ATKP], "ele", st, "ele %", ds.Stats[st], "bonus dmg", ds.DmgBonus, "mul", ds.Mult)
	//calculate using attack or def
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

	log.Debugw("\t\tcalc", "cr", ds.Stats[CR], "cd", ds.Stats[CD], "def adj", ds.DefMod, "res adj", target.ResMod[ds.Element], "char lvl", ds.CharLvl, "target lvl", target.Level, "target res", target.Resist[ds.Element])

	defmod := float64(ds.CharLvl+100) / (float64(ds.CharLvl+100) + float64(target.Level+100)*(1-ds.DefMod))
	//apply def mod
	damage = damage * defmod
	//add up the resist mods
	var rm float64
	for _, v := range target.ResMod[ds.Element] {
		rm += v
	}
	//apply resist mod
	res := target.Resist[ds.Element] + rm

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
