package mona

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("mona", NewChar)
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
	c.Energy = 80
	c.MaxEnergy = 80
	c.Weapon.Class = combat.WeaponClassCatalyst

	if c.Base.Cons == 6 {
		c.c6()
	}

	return &c, nil
}

func (c *char) ActionFrames(a combat.ActionType, p int) int {
	switch a {
	case combat.ActionAttack:
		f := 0
		switch c.NormalCounter {
		//TODO: need to add atkspd mod
		case 0:
			f = 11
		case 1:
			f = 33
		case 2:
			f = 60
		case 3:
			f = 60
		}
		f = int(float64(f) / (1 + c.Stats[combat.AtkSpd]))
		return f
	case combat.ActionCharge:
		return 100
	case combat.ActionSkill:
		if p == 0 {
			return 1
		}
		return 10
	case combat.ActionBurst:
		return 1
	default:
		c.S.Log.Warnf("%v: unknown action, frames invalid", a)
		return 0
	}
}

func (c *char) c6() {
}

func (c *char) Attack(p int) int {
	delay := []int{1, 2, 3, 4} //TODO: frames
	d := c.Snapshot("Normal", combat.ActionAttack, combat.Electro, 0)
	d.Mult = attack[c.NormalCounter][c.TalentLvlAttack()]
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
		s.Log.Infof("\t %v normal %v dealt %.0f damage [%v]", c.Base.Name, c.NormalCounter, damage, str)
	}, fmt.Sprintf("%v-Normal-%v", c.Base.Name, c.NormalCounter), delay[c.NormalCounter])

	c.NormalResetTimer = 70
	c.NormalCounter++
	if c.NormalCounter == 4 {
		c.NormalCounter = 0
		c.NormalResetTimer = 0
	}

	return f
}
