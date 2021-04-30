package combat

import (
	"math/rand"
	"strings"

	"go.uber.org/zap"
)

//ApplyDamage applies damage to the target given a snapshot
func (s *Sim) ApplyDamage(ds Snapshot) (float64, string) {
	var sb strings.Builder

	s.Log.Debugf("[%v] %v - %v triggered dmg", s.Frame(), ds.Actor, ds.Abil)
	s.Log.Debugw("\t target", "auras", s.TargetAura, "ele applied", ds.Element, "dur applied", ds.Durability)
	old := s.TargetAura
	//apply new aura
	s.TargetAura = s.TargetAura.React(ds, s)

	if old.E() != s.TargetAura.E() {
		s.Log.Infof("\t [%v] previous ele <%v>, next ele <%v>", s.Frame(), old.E(), s.TargetAura.E())
	}

	//change multiplier if vape or melt
	// if s.GlobalFlags.ReactionDidOccur {
	// 	if s.GlobalFlags.ReactionType == Melt || s.GlobalFlags.ReactionType == Vaporize {
	// 		s.Log.Debugf("hello? vape %v", s.GlobalFlags.NextAttackMVMult)
	// 		ds.IsMeltVape = true
	// 		ds.ReactMult = s.GlobalFlags.NextAttackMVMult
	// 	}
	// }
	if s.GlobalFlags.NextAttackMVMult > 1 {
		ds.IsMeltVape = true
		ds.ReactMult = s.GlobalFlags.NextAttackMVMult
	}

	//caculate damage
	s.executeSnapshotHooks(PreDamageHook, &ds)
	dr := calcDmg(ds, *s.Target, s.Rand, s.Log)
	s.Target.Damage += dr.damage
	s.Target.HP -= dr.damage
	if s.Stats.LogStats {
		s.Stats.DamageByChar[ds.Actor][ds.Abil] += dr.damage
	}

	if dr.isCrit {
		s.executeSnapshotHooks(OnCritDamage, &ds)
		sb.WriteString(" crit")
	}
	s.executeSnapshotHooks(PostDamageHook, &ds)

	//check if reaction occured and call hooks
	if s.GlobalFlags.ReactionDidOccur {
		s.executeSnapshotHooks(PreReaction, &ds)
		s.Log.Debugf("\t reaction %v occured", s.GlobalFlags.ReactionType)
		sb.WriteString(" ")
		sb.WriteString(string(s.GlobalFlags.ReactionType))

		if s.Stats.LogStats {
			s.Stats.ReactionsTriggered[s.GlobalFlags.ReactionType]++
		}
	}

	//apply reaction damage
	if s.GlobalFlags.ReactionDamageTriggered {
		//add superconduct buff if triggered
		if s.GlobalFlags.ReactionType == Superconduct {
			s.Target.AddResMod("Superconduct", ResistMod{
				Duration: 12 * 60,
				Ele:      Physical,
				Value:    -0.4,
			})
		}
		s.applyReactionDamage(ds, s.GlobalFlags.ReactionType)
	}

	if s.GlobalFlags.ReactionDidOccur {
		s.executeSnapshotHooks(PostReaction, &ds)
	}

	//reset reaction flags
	s.ResetReactionFlags()
	sb.WriteString(" ")
	return dr.damage, sb.String()
}

func (s *Sim) ResetReactionFlags() {
	s.GlobalFlags.ReactionDidOccur = false
	s.GlobalFlags.ReactionType = ""
	s.GlobalFlags.NextAttackMVMult = 1 // melt vape multiplier
}

type dmgResult struct {
	damage float64
	isCrit bool
}

func calcDmg(ds Snapshot, target Enemy, rand *rand.Rand, log *zap.SugaredLogger) dmgResult {

	result := dmgResult{}

	st := EleToDmgP(ds.Element)
	ds.DmgBonus += ds.Stats[st] + ds.Stats[DmgP]
	log.Debugf("\t\t -------------------------------------------------")
	log.Debugf("\t\t |Calculating damage for %v: %v (source frame: %v)", ds.Actor, ds.Abil, ds.SourceFrame)
	log.Debugw("\t\t |calc", "base atk", ds.BaseAtk, "flat +", ds.Stats[ATK], "% +", ds.Stats[ATKP], "ele", st, "ele %", ds.Stats[st], "bonus dmg", ds.DmgBonus, "mul", ds.Mult)
	//calculate using attack or def
	var a float64
	if ds.UseDef {
		a = ds.BaseDef*(1+ds.Stats[DEFP]) + ds.Stats[DEF]
	} else {
		a = ds.BaseAtk*(1+ds.Stats[ATKP]) + ds.Stats[ATK]
	}

	base := ds.Mult*a + ds.FlatDmg
	damage := base * (1 + ds.DmgBonus)

	log.Debugw("\t\t |calc", "total atk", a, "base dmg", base, "dmg + bonus", damage)

	//make sure 0 <= cr <= 1
	if ds.Stats[CR] < 0 {
		ds.Stats[CR] = 0
	}
	if ds.Stats[CR] > 1 {
		ds.Stats[CR] = 1
	}
	res := target.Resist(log)[ds.Element]

	log.Debugw("\t\t |calc", "cr", ds.Stats[CR], "cd", ds.Stats[CD], "def adj", ds.DefMod, "char lvl", ds.CharLvl, "target lvl", target.Level, "target res", res)

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
		log.Debugw("\t\t |calc", "other mult", ds.OtherMult)
		damage = damage * ds.OtherMult
	}
	log.Debugw("\t\t |calc", "def mod", defmod, "res", res, "res mod", resmod)
	log.Debugw("\t\t |calc", "pre crit damage", damage, "dmg if crit", damage*(1+ds.Stats[CD]), "melt/vape", ds.IsMeltVape)

	//check melt/vape
	if ds.IsMeltVape {
		//calculate em bonus
		em := ds.Stats[EM]
		emBonus := (2.78 * em) / (1400 + em)
		log.Debugw("\t\t |calc", "react mult", ds.ReactMult, "em", em, "em bonus", emBonus, "react bonus", ds.ReactBonus, "pre react damage", damage)
		damage = damage * (ds.ReactMult * (1 + emBonus + ds.ReactBonus))
		log.Debugw("\t\t |calc", "pre crit (post react) damage", damage, "pre react if crit", damage*(1+ds.Stats[CD]))
	}

	//check if crit
	if rand.Float64() <= ds.Stats[CR] || ds.HitWeakPoint {
		log.Debugf("\t\t |damage is crit!")
		damage = damage * (1 + ds.Stats[CD])
		result.isCrit = true
	}

	result.damage = damage

	log.Debugw("\t\t |Calculation result:", "damage", damage)
	log.Debugf("\t\t -------------------------------------------------")

	return result
}
