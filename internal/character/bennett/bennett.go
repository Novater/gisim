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
	b.Profile.WeaponClass = combat.WeaponClassSword

	a4 := make(map[combat.StatType]float64)
	a4[combat.HydroP] = 0.2
	b.AddMod("Xingqiu A4", a4)

	return &b, nil
}

func (b *bennett) Attack(p map[string]interface{}) int {
	//register action depending on number in chain
	//3 and 4 need to be registered as multi action
	d := b.Snapshot("Normal", combat.ActionTypeAttack, combat.Physical)
	//figure out which hit it is
	var hits []float64
	reset := false
	frames := 21 //first hit
	n := 1
	switch b.NormalCounter {
	case 1:
		hits = n2
		frames = 48 - 21
		n = 2
	case 2:
		hits = n3
		frames = 74 - 48
		n = 3
	case 3:
		hits = n4
		frames = 114 - 74
		n = 4
	case 4:
		hits = n5
		frames = 180 - 114
		n = 5
		reset = true
	default:
		hits = n1
	}
	b.NormalCounter++
	//apply attack speed
	frames = int(float64(frames) / (1 + b.Stats[combat.AtkSpd]))
	d.Mult = hits[b.Profile.TalentLevel[combat.ActionTypeAttack]-1]
	b.S.AddAction(func(s *combat.Sim) bool {
		//no delay for now? realistically the hits should have delay
		//i guess this screws up elemental application if/when infused
		damage := s.ApplyDamage(d)
		s.Log.Infof("[%v]: Bennett normal %v dealt %.0f damage", s.Frame(), n, damage)
		return true
	}, fmt.Sprintf("%v-Bennett-Normal-%v", b.S.Frame(), n))

	//add a 75 frame attackcounter reset
	b.NormalResetTimer = 70

	if reset {
		b.NormalResetTimer = 0
	}
	//return animation cd
	//this also depends on which hit in the chain this is
	return frames
}

func (b *bennett) ChargeAttackStam() float64 {
	return 20
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
			if delay < 26 { //TODO: fix delay
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
			if delay < 89 {
				return false
			}
			damage := s.ApplyDamage(d1)
			s.Log.Infof("[%v]: Bennett skill (hold 1 - hit 1) dealt %.0f damage", s.Frame(), damage)
			return true
		}, fmt.Sprintf("%v-Bennett-Skill-Hold-1-1", b.S.Frame()))
		d2 := b.Snapshot("Passion Overload", combat.ActionTypeSkill, combat.Pyro)
		d2.ApplyAura = true
		d2.AuraBase = combat.MedAuraBase
		d2.AuraUnits = 2
		d2.Mult = skill1[1][lvl]
		delay2 := 0
		b.S.AddAction(func(s *combat.Sim) bool {
			delay2++
			if delay2 < 115 {
				return false
			}
			damage := s.ApplyDamage(d2)
			s.Log.Infof("[%v]: Bennett skill (hold 1 - hit 2) dealt %.0f damage", s.Frame(), damage)
			return true
		}, fmt.Sprintf("%v-Bennett-Skill-Hold-1-2", b.S.Frame()))
	case 2:
		d1 := b.Snapshot("Passion Overload", combat.ActionTypeSkill, combat.Pyro)
		d1.ApplyAura = true
		d1.AuraBase = combat.MedAuraBase
		d1.AuraUnits = 2
		d1.Mult = skill2[0][lvl]
		delay := 0
		b.S.AddAction(func(s *combat.Sim) bool {
			delay++
			if delay < 136 {
				return false
			}
			damage := s.ApplyDamage(d1)
			s.Log.Infof("[%v]: Bennett skill (hold 2 - hit 1) dealt %.0f damage", s.Frame(), damage)
			return true
		}, fmt.Sprintf("%v-Bennett-Skill-Hold-2-1", b.S.Frame()))
		d2 := b.Snapshot("Passion Overload", combat.ActionTypeSkill, combat.Pyro)
		d2.ApplyAura = true
		d2.AuraBase = combat.MedAuraBase
		d2.AuraUnits = 2
		d2.Mult = skill2[1][lvl]
		delay2 := 0
		b.S.AddAction(func(s *combat.Sim) bool {
			delay2++
			if delay2 < 154 {
				return false
			}
			damage := s.ApplyDamage(d2)
			s.Log.Infof("[%v]: Bennett skill (hold 2 - hit 2) dealt %.0f damage", s.Frame(), damage)
			return true
		}, fmt.Sprintf("%v-Bennett-Skill-Hold-2-2", b.S.Frame()))
		d3 := b.Snapshot("Passion Overload", combat.ActionTypeSkill, combat.Pyro)
		d3.ApplyAura = true
		d3.AuraBase = combat.MedAuraBase
		d3.AuraUnits = 2
		d3.Mult = explosion[lvl]
		delay3 := 0
		b.S.AddAction(func(s *combat.Sim) bool {
			delay3++
			if delay3 < 198 {
				return false
			}
			damage := s.ApplyDamage(d2)
			s.Log.Infof("[%v]: Bennett skill (hold 2 - hit 3) dealt %.0f damage", s.Frame(), damage)
			return true
		}, fmt.Sprintf("%v-Bennett-Skill-Hold-2-3", b.S.Frame()))
	}

	//A2
	reduction := 0.2
	//A4
	if _, ok := b.S.Status["Bennett Burst"]; ok {
		reduction += 0.5
	}

	switch hold {
	case 1:
		cd := int(7.5 * 60 * (1 - reduction))
		b.CD[common.SkillCD] = cd
		return 153 //not right?
	case 2:
		cd := int(10 * 60 * (1 - reduction))
		b.CD[common.SkillCD] = cd
		return 370 //too high as well
	}

	cd := int(5 * 60 * (1 - reduction))
	b.CD[common.SkillCD] = cd //should be 7.5 or 10
	return 52
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
		if burstDelay < 43 { //TODO: fix delay
			return false
		}
		damage := s.ApplyDamage(d)
		s.Log.Infof("[%v]: Bennett burst dealt %.0f damage", s.Frame(), damage)
		return true
	}, fmt.Sprintf("%v-Bennett-Burst", b.S.Frame()))

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

	//hook for buffs; active right away after cast
	atk := burstatk[lvl] * float64(b.Profile.BaseAtk+b.Profile.WeaponBaseAtk)
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

	//add a status
	b.S.Status["Bennett Burst"] = 12 * 60

	b.CD[common.BurstCD] = 15 * 60
	return 51 //todo fix field cast time
}
