package klee

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("klee", NewChar)
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
	c.Energy = 60
	c.MaxEnergy = 60
	c.Weapon.Class = combat.WeaponClassCatalyst

	c.a2()
	c.a4()

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
		}
		f = int(float64(f) / (1 + c.Stats[combat.AtkSpd]))
		return f
	case combat.ActionSkill:
		return 1
	case combat.ActionBurst:
		return 1
	default:
		c.S.Log.Warnf("%v: unknown action, frames invalid", a)
		return 0
	}
}

func (c *char) a2() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor != c.Base.Name {
			return false
		}
		if ds.AbilType != combat.ActionAttack && ds.AbilType != combat.ActionSkill {
			return false
		}
		if c.S.Rand.Float64() <= 0.5 {
			c.Tags["explosive spark"] = 1
		}
		return false
	}, "klee-a2", combat.PostDamageHook)
}

func (c *char) a4() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor != c.Base.Name {
			return false
		}
		if ds.AbilType != combat.ActionCharge {
			return false
		}
		for _, c := range c.S.Chars {
			c.AddEnergy(2)
		}
		return false
	}, "klee-a2", combat.OnCritDamage)
}

func (c *char) Attack(p int) int {
	delay := []int{1, 2, 3} //TODO: frames
	d := c.Snapshot("Normal", combat.ActionAttack, combat.Pyro, 0)
	d.Mult = attack[c.NormalCounter][c.TalentLvlAttack()]
	d.IsHeavyAttack = true
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
	if c.NormalCounter == 3 {
		c.NormalCounter = 0
		c.NormalResetTimer = 0
	}

	return f
}

//p is the number of mines that hit, up to ??
func (c *char) Skill(p int) int {

	return c.ActionFrames(combat.ActionSkill, p)
}
