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
	b.Energy = 80
	b.MaxEnergy = 80

	a4 := make(map[combat.StatType]float64)
	a4[combat.HydroP] = 0.2
	b.AddMod("Xingqiu A4", a4)

	return &b, nil
}

func (x *bennett) Skill(p map[string]interface{}) int {
	if _, ok := x.CD[common.SkillCD]; ok {
		x.S.Log.Debugf("\tBennett skill still on CD; skipping")
		return 0
	}

	hold := 0
	if v, ok := p["Hold"]; ok {
		if h, isInt := v.(int); isInt {
			hold = int(h)
		}
	}
	if hold > 0 {
		x.S.Log.Debugf("\tBennet using skill, hold level = %v", hold)
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
		d.OtherMult = 1 //not implemented for now
	}
	d.ApplyAura = true
	d.AuraBase = combat.WeakAuraBase
	d.AuraUnits = 1
	d.Mult = rainscreen[0][lvl]
	d2 := d.Clone()
	d2.ApplyAura = false
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
	}, fmt.Sprintf("%v-Xiangling-Skill-Orb", x.S.Frame()))

	//should last 15s, cd 21s
	x.CD[common.SkillCD] = 21 * 60
	return 77
}

func (b *bennett) Burst(p map[string]interface{}) int {
	return 0
}
