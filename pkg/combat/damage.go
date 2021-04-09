package combat

import (
	"math/rand"

	"go.uber.org/zap"
)

//ApplyDamage applies damage to the target given a snapshot
func (s *Sim) ApplyDamage(ds Snapshot) float64 {

	s.Log.Debugf("  [%v] %v - %v triggered dmg", s.Frame(), ds.Actor, ds.Abil)
	s.Log.Debugw("\t target", "auras", s.TargetAura, "ele applied", ds.Element, "dur applied", ds.Durability)

	//change multiplier if vape or melt
	if s.GlobalFlags.NextAttackMVMult > 1 {
		ds.IsMeltVape = true
		ds.ReactMult = s.GlobalFlags.NextAttackMVMult
	}

	//caculate damage
	s.executeSnapshotHooks(PreDamageHook, &ds)
	dr := calcDmg(ds, *s.Target, s.Rand, s.Log)
	s.Target.Damage += dr.damage
	s.Stats.DamageByChar[ds.Actor][ds.Abil] += dr.damage

	if dr.isCrit {
		s.executeSnapshotHooks(OnCritDamage, &ds)
	}
	s.executeSnapshotHooks(PostDamageHook, &ds)

	//apply new aura
	s.TargetAura = s.TargetAura.React(ds, s)

	s.Log.Debugw("\t target", "next aura", s.TargetAura)

	//check if reaction occured and call hooks
	if s.GlobalFlags.ReactionDidOccur {
		s.executeSnapshotHooks(PreReaction, &ds)
	}

	//add superconduct buff if triggered
	if s.GlobalFlags.NextAttackSuperconductTriggered {
		s.Target.AddResMod("Superconduct", ResistMod{
			Duration: 12 * 60,
			Ele:      Physical,
			Value:    -0.4,
		})
	}

	//apply reaction damage
	if s.GlobalFlags.NextAttackOverloadTriggered {
		s.applyReactionDamage(ds, Overload)
	}
	if s.GlobalFlags.NextAttackSuperconductTriggered {
		s.applyReactionDamage(ds, Superconduct)
	}
	if s.GlobalFlags.NextAttackShatterTriggered {
		s.applyReactionDamage(ds, Shatter)
	}

	if s.GlobalFlags.ReactionDidOccur {
		s.executeSnapshotHooks(PostReaction, &ds)
	}

	//reset reaction flags
	s.ResetReactionFlags()

	return dr.damage
}

func (s *Sim) ResetReactionFlags() {
	s.GlobalFlags.ReactionDidOccur = false
	s.GlobalFlags.ReactionType = ""
	s.GlobalFlags.NextAttackMVMult = 1 // melt vape multiplier
	s.GlobalFlags.NextAttackOverloadTriggered = false
	s.GlobalFlags.NextAttackSuperconductTriggered = false
	s.GlobalFlags.NextAttackShatterTriggered = false
}

type dmgResult struct {
	damage float64
	isCrit bool
}

func calcDmg(ds Snapshot, target Enemy, rand *rand.Rand, log *zap.SugaredLogger) dmgResult {

	result := dmgResult{}

	st := EleToDmgP(ds.Element)
	ds.DmgBonus += ds.Stats[st] + ds.Stats[DmgP]

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
	res := target.Resist(log)[ds.Element]

	log.Debugw("\t\tcalc", "cr", ds.Stats[CR], "cd", ds.Stats[CD], "def adj", ds.DefMod, "char lvl", ds.CharLvl, "target lvl", target.Level, "target res", res)

	defmod := float64(ds.CharLvl+100) / (float64(ds.CharLvl+100) + float64(target.Level+100)*(1-ds.DefMod))
	//apply def mod
	damage = damage * defmod
	//apply resist mod

	resmod := 1 - res/2
	if res >= 0 && res < 0.75 {
		resmod = 1 - res
	} else if res > 0.75 {
		resmod = 1 / (4*res + 1)
	}
	damage = damage * resmod

	//apply other multiplier bonus
	if ds.OtherMult > 0 {
		log.Debugw("\t\tcalc", "other mult", ds.OtherMult)
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
