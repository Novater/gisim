package bennett

import (
	"fmt"

	"github.com/srliao/gisim/internal/character/common"
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Bennett", NewChar)
}

type bennett struct {
	*common.TemplateChar
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	b := bennett{}
	t, err := common.New(s, p)

	if err != nil {
		return nil, err
	}
	b.TemplateChar = t
	b.Energy = 60
	b.MaxEnergy = 60

	a4 := make(map[combat.StatType]float64)
	a4[combat.HydroP] = 0.2
	b.AddMod("Xingqiu A4", a4)

	return &b, nil
}

func (b *bennett) Skill(p map[string]interface{}) int {
	if _, ok := b.CD[common.SkillCD]; ok {
		b.S.Log.Debugf("\tBennett skill still on CD; skipping")
		return 0
	}
	hold := 0
	if v, ok := p["Hold"]; ok {
		h, n := v.(int)
		if n {
			hold = h
		}
	}

	lvl := b.Profile.TalentLevel[combat.ActionTypeSkill] - 1
	if b.Profile.Constellation >= 3 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	switch hold {
	case 0:
		d := b.Snapshot("Passion Overload", combat.ActionTypeSkill, combat.Pyro)
		d.ApplyAura = true
		d.AuraBase = combat.MedAuraBase
		d.AuraUnits = 2
		d.Mult = skill[lvl]
		delay := 0
		b.S.AddAction(func(s *combat.Sim) bool {
			delay++
			if delay < 50 { //TODO: fix delay
				return false
			}
			damage := s.ApplyDamage(d)
			s.Log.Infof("[%v]: Bennett skill dealt %.0f damage", s.Frame(), damage)
			return true
		}, fmt.Sprintf("%v-Bennett-Skill", b.S.Frame()))
	case 1:
		d1 := b.Snapshot("Passion Overload", combat.ActionTypeSkill, combat.Pyro)
		d1.ApplyAura = true
		d1.AuraBase = combat.MedAuraBase
		d1.AuraUnits = 2
		d1.Mult = skill1[0][lvl]
		delay := 0
		b.S.AddAction(func(s *combat.Sim) bool {
			delay++
			if delay < 50 { //TODO: fix delay
				return false
			}
			damage := s.ApplyDamage(d1)
			s.Log.Infof("[%v]: Bennett skill dealt %.0f damage", s.Frame(), damage)
			return true
		}, fmt.Sprintf("%v-Bennett-Skill", b.S.Frame()))
		d2 := b.Snapshot("Passion Overload", combat.ActionTypeSkill, combat.Pyro)
		d2.ApplyAura = true
		d2.AuraBase = combat.MedAuraBase
		d2.AuraUnits = 2
		d2.Mult = skill1[1][lvl]
		delay2 := 0
		b.S.AddAction(func(s *combat.Sim) bool {
			delay2++
			if delay2 < 50 { //TODO: fix delay
				return false
			}
			damage := s.ApplyDamage(d2)
			s.Log.Infof("[%v]: Bennett skill dealt %.0f damage", s.Frame(), damage)
			return true
		}, fmt.Sprintf("%v-Bennett-Skill", b.S.Frame()))
	case 2:
		d1 := b.Snapshot("Passion Overload", combat.ActionTypeSkill, combat.Pyro)
		d1.ApplyAura = true
		d1.AuraBase = combat.MedAuraBase
		d1.AuraUnits = 2
		d1.Mult = skill1[0][lvl]
		delay := 0
		b.S.AddAction(func(s *combat.Sim) bool {
			delay++
			if delay < 50 { //TODO: fix delay
				return false
			}
			damage := s.ApplyDamage(d1)
			s.Log.Infof("[%v]: Bennett skill dealt %.0f damage", s.Frame(), damage)
			return true
		}, fmt.Sprintf("%v-Bennett-Skill", b.S.Frame()))
		d2 := b.Snapshot("Passion Overload", combat.ActionTypeSkill, combat.Pyro)
		d2.ApplyAura = true
		d2.AuraBase = combat.MedAuraBase
		d2.AuraUnits = 2
		d2.Mult = skill1[1][lvl]
		delay2 := 0
		b.S.AddAction(func(s *combat.Sim) bool {
			delay2++
			if delay2 < 50 { //TODO: fix delay
				return false
			}
			damage := s.ApplyDamage(d2)
			s.Log.Infof("[%v]: Bennett skill dealt %.0f damage", s.Frame(), damage)
			return true
		}, fmt.Sprintf("%v-Bennett-Skill", b.S.Frame()))
	}

	b.CD[common.SkillCD] = 5 * 60 //should be 7.5 or 10
	return 77
}

func (b *bennett) Burst(p map[string]interface{}) int {
	if _, ok := b.CD[common.BurstCD]; ok {
		b.S.Log.Debugf("\tBennett burst still on CD; skipping")
		return 0
	}
	d := b.Snapshot("Fantastic Voyage", combat.ActionTypeBurst, combat.Pyro)
	lvl := b.Profile.TalentLevel[combat.ActionTypeBurst] - 1
	if b.Profile.Constellation >= 5 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	d.ApplyAura = true
	d.AuraBase = combat.MedAuraBase
	d.AuraUnits = 2
	d.Mult = burst[lvl]
	burstDelay := 0
	b.S.AddAction(func(s *combat.Sim) bool {
		burstDelay++
		if burstDelay < 50 { //TODO: fix delay
			return false
		}
		damage := s.ApplyDamage(d)
		s.Log.Infof("[%v]: Bennett burst dealt %.0f damage", s.Frame(), damage)
		return true
	}, fmt.Sprintf("%v-Bennett-Burst", b.S.Frame()))

	atk := burstatk[lvl] * float64(b.Profile.BaseAtk+b.Profile.WeaponBaseAtk)
	//we can use the same delay for field buff
	b.S.AddHook(func(ds *combat.Snapshot) bool {
		if b.S.ActiveChar != ds.CharName {
			return false
		}
		//TODO: should have an HP check here but no one ever takes damage in this sim..
		ds.Stats[combat.ATK] += atk
		if b.Profile.Constellation == 6 {
			ok := ds.AbilType == combat.ActionTypeAttack || ds.AbilType == combat.ActionTypeChargedAttack
			ok = ok && (ds.WeaponClass == combat.WeaponClassSpear || ds.WeaponClass == combat.WeaponClassSword || ds.WeaponClass == combat.WeaponClassClaymore)
			if ok {
				ds.Element = combat.Pyro
				ds.Stats[combat.PyroP] += 0.15
			}
		}
		return false
	}, "Bennett-Burst-Field", combat.PreSnapshot)
	//add field effect timer
	ft := 0
	b.S.AddAction(func(s *combat.Sim) bool {
		ft++
		if ft < 12*60 { //12 seconds, first tick after 1 sec
			return false
		}
		//TODO: should heal here but no one takes damage...
		b.S.RemoveHook("Bennett-Burst-Field", combat.PreSnapshot)
		return true
	}, fmt.Sprintf("%v-Bennett-Burst-Field", b.S.Frame()))

	b.CD[common.BurstCD] = 15 * 60
	return 0 //todo fix field cast time
}
