package xingqiu

import (
	"fmt"

	"github.com/srliao/gisim/internal/character/common"
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Xingqiu", NewChar)
}

type xingqiu struct {
	*common.TemplateChar
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	x := xingqiu{}
	t, err := common.New(s, p)
	if err != nil {
		return nil, err
	}
	x.TemplateChar = t
	x.Energy = 80
	x.MaxEnergy = 80
	x.Profile.WeaponClass = combat.WeaponClassSword

	a4 := make(map[combat.StatType]float64)
	a4[combat.HydroP] = 0.2
	x.AddMod("Xingqiu A4", a4)

	//c2
	if x.Profile.Constellation >= 2 {
		s.Log.Debugf("\tactivating Xingqiu C2")

		s.AddHook(func(snap *combat.Snapshot) bool {
			//check if c1 debuff is on, if so, reduce resist by -0.15
			if _, ok := s.Target.Status["xingqiu-c2"]; ok {
				s.Log.Debugf("\t[%v]: applying Xingqiu C2 hydro debuff", s.Frame())
				snap.ResMod[combat.Hydro] -= 0.15
			}
			return false
		}, "xingqiu-c2", combat.PreDamageHook)
	}

	/** c6
	Activating 2 of Guhua Sword: Raincutter's sword rain attacks greatly increases the DMG of the third.
	Xingqiu regenerates 3 Energy when sword rain attacks hit opponents.
	**/

	return &x, nil
}

func (x *xingqiu) Skill(p map[string]interface{}) int {
	//applies wet to self 30 frame after cast
	if _, ok := x.CD[common.SkillCD]; ok {
		x.S.Log.Debugf("\tXingqiu skill still on CD; skipping")
		return 0
	}
	d := x.Snapshot("Guhua Sword: Fatal Rainscreen", combat.ActionTypeSkill, combat.Hydro)
	lvl := x.Profile.TalentLevel[combat.ActionTypeSkill] - 1
	if x.Profile.Constellation >= 5 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	if x.Profile.Constellation >= 4 {
		//check if ult is up, if so increase multiplier
		if _, ok := x.S.Status["Xingqiu-Burst"]; ok {
			d.OtherMult = 1.5
		}
	}
	d.ApplyAura = true
	d.AuraBase = combat.WeakAuraBase
	d.AuraUnits = 1
	d.Mult = rainscreen[0][lvl]
	d2 := d.Clone()
	d2.Mult = rainscreen[1][lvl]

	tick := 0
	x.S.AddAction(func(s *combat.Sim) bool {
		tick++
		if tick < 50 {
			return false
		}
		damage := s.ApplyDamage(d)
		s.Log.Infof("[%v]: Xingqiu skill hit 1 dealt %.0f damage", s.Frame(), damage)
		return true
	}, fmt.Sprintf("%v-Xingqiu-Skill-1", x.S.Frame()))

	x.S.AddAction(func(s *combat.Sim) bool {
		tick++
		if tick < 51 {
			return false
		}
		damage := s.ApplyDamage(d2)
		s.Log.Infof("[%v]: Xingqiu skill hit 2 dealt %.0f damage", s.Frame(), damage)
		return true
	}, fmt.Sprintf("%v-Xingqiu-Skill-2", x.S.Frame()))

	orbDelay := 0
	x.S.AddAction(func(s *combat.Sim) bool {
		if orbDelay < 60+60 {
			orbDelay++
			return false
		}
		s.GenerateOrb(5, combat.Hydro, false)
		return true
	}, fmt.Sprintf("%v-Xingqiu-Skill-Orb", x.S.Frame()))

	//should last 15s, cd 21s
	x.CD[common.SkillCD] = 21 * 60
	return 77
}

func (x *xingqiu) Burst(p map[string]interface{}) int {
	//apply hydro every 3rd hit
	//triggered on normal attack
	//not sure what ICD is
	if _, ok := x.CD[common.BurstCD]; ok {
		x.S.Log.Debugf("\tXingqiu skill still on CD; skipping")
		return 0
	}

	/** c2
	Extends the duration of Guhua Sword: Raincutter by 3s.
	Decreases the Hydro RES of opponents hit by sword rain attacks by 15% for 4s.
	**/
	dur := 15
	if x.Profile.Constellation >= 2 {
		dur += 3
	}
	dur = dur * 60
	x.S.Status["Xingqiu-Burst"] = dur

	burstCounter := 0
	x.S.AddHook(func(ds *combat.Snapshot) bool {
		dur--
		if dur < 0 {
			return true
		}

		//check if off ICD
		if _, ok := x.CD["Xingqiu-Burst-ICD"]; !ok {
			return false
		}

		//check if normal attack
		if ds.AbilType != combat.ActionTypeAttack && ds.AbilType != combat.ActionTypeChargedAttack {
			return false
		}

		d := x.Snapshot("Guhua Sword: Raincutter", combat.ActionTypeBurst, combat.Hydro)
		lvl := x.Profile.TalentLevel[combat.ActionTypeBurst] - 1
		if x.Profile.Constellation >= 3 {
			lvl += 3
			if lvl > 14 {
				lvl = 14
			}
		}
		d.Mult = burst[lvl]

		//apply aura every 3rd hit -> hit 0, 3, 6, etc...
		if burstCounter%3 == 0 {
			d.ApplyAura = true
			d.AuraBase = combat.WeakAuraBase
			d.AuraUnits = 1
		}

		x.S.AddAction(func(s *combat.Sim) bool {
			s.ApplyDamage(d)
			//add hydro debuff for 4s
			if x.Profile.Constellation >= 2 {
				s.Target.Status["xingqiu-c2"] = 4 * 60
			}
			return true
		}, fmt.Sprintf("%v-Xingqiu-Burst-Proc", x.S.Frame()))

		//TODO: how long is his ICD?
		x.CD["Xingqiu-Burst-ICD"] = 1
		burstCounter++

		return false
	}, "Xingqiu-Burst", combat.PostDamageHook)

	//remove the hook if off CD
	x.S.AddAction(func(s *combat.Sim) bool {
		_, ok := s.Status["Xingqiu-Burst"]
		if !ok {
			s.RemoveHook("Xingqiu-Burst", combat.PostDamageHook)
		}
		return !ok
	}, fmt.Sprintf("%v-Xingqiu-Burst", x.S.Frame()))

	x.CD[common.BurstCD] = 20 * 60

	return 0
}
