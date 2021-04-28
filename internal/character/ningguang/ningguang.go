package ningguang

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("ningguang", NewChar)
}

type char struct {
	*combat.CharacterTemplate
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	c := char{}
	t, err := combat.NewTemplateChar(s, p)
	if err != nil {
		return nil, err
	}
	c.CharacterTemplate = t
	c.Energy = 40
	c.MaxEnergy = 40
	c.Weapon.Class = combat.WeaponClassClaymore

	return &c, nil
}

func (c *char) ActionFrames(a combat.ActionType, p int) int {
	switch a {
	case combat.ActionAttack:
		f := 10
		return int(float64(f) / (1 + c.Stats[combat.AtkSpd]))
	case combat.ActionCharge:
		return 100
	case combat.ActionSkill:
		return 10
	case combat.ActionBurst:
		return 1
	default:
		c.S.Log.Warnf("%v: unknown action, frames invalid", a)
		return 0
	}
}

func (c *char) Attack(p int) int {
	d := c.Snapshot("Normal", combat.ActionAttack, combat.Geo, combat.WeakDurability)
	d.Mult = attack[c.TalentLvlAttack()]
	f := c.ActionFrames(combat.ActionAttack, p)

	//apply every 3rd hit or every 2.5s
	//TODO: check if this code works
	if c.CD[combat.NormalICD] <= c.S.F {
		c.Tags[combat.NormalICD] = 1
		c.CD[combat.NormalICD] = c.S.F + 150 //2.5s
		d.Durability = combat.WeakDurability
	} else if c.Tags[combat.NormalICD] == 4 {
		c.Tags[combat.NormalICD] = 1
		d.Durability = combat.WeakDurability
	} else {
		c.Tags[combat.NormalICD]++
	}

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t %v normal dealt %.0f damage [%v]", c.Base.Name, damage, str)
	}, "Ningguang - Normal", 1) //TODO: frames

	count := c.Tags["jade"]
	if count != 7 {
		count++
		if count > 3 {
			count = 3
		}
		c.Tags["jade"] = count
	}

	return f
}

func (c *char) Charge(p int) int {
	d := c.Snapshot("Charge", combat.ActionAttack, combat.Geo, combat.WeakDurability)
	d.Mult = charge[c.TalentLvlAttack()]
	//apply every 3rd hit or every 2.5s
	//TODO: check if this code works
	if c.CD[combat.ChargedICD] <= c.S.F {
		c.Tags[combat.ChargedICD] = 1
		c.CD[combat.ChargedICD] = c.S.F + 150 //2.5s
		d.Durability = combat.WeakDurability
	} else if c.Tags[combat.ChargedICD] == 4 {
		c.Tags[combat.ChargedICD] = 1
		d.Durability = combat.WeakDurability
	} else {
		c.Tags[combat.ChargedICD]++
	}

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Ningguang charge dealt %.0f damage [%v]", damage, str)
	}, "Ningguang - Charge", 1) //TODO: frames

	j := c.Tags["jade"]
	for i := 0; i < j; i++ {
		d := c.Snapshot("Charge", combat.ActionAttack, combat.Geo, 0)
		d.Mult = jade[c.TalentLvlAttack()]
		t := i + 1
		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Ningguang charge (jades) dealt %.0f damage [%v]", damage, str)
		}, fmt.Sprintf("Ningguang - Jades #%v", t), 1) //TODO: frames
	}
	c.Tags["jade"] = 0

	return c.ActionFrames(combat.ActionCharge, p)
}

func (c *char) Skill(p int) int {
	return 0
}
