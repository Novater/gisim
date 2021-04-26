package lisa

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("lisa", NewChar)
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
	c.S.AddEventHook(func(s *combat.Sim) bool {
		if c.S.ActiveChar == c.Base.Name {
			//swapped to lisa
			c.Tags["stack"] = 3
		}
		return false
	}, "lisa-c6", combat.PostSwapHook)
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

func (c *char) ChargeAttack(p int) int {
	d := c.Snapshot("Charge Attack", combat.ActionCharge, combat.Electro, combat.WeakDurability)
	d.Mult = charge[c.TalentLvlAttack()]
	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Lisa charge attack dealt %.0f damage [%v]", damage, str)
	}, "Lisa-Charge-Attack", 1) //TODO: frames

	//a2 add a stack
	if c.Tags["stack"] < 3 {
		c.Tags["stack"]++
	}

	//c1 adds energy (just 1 for now)
	if c.Base.Cons > 0 {
		c.Energy += 2
		if c.Energy > c.MaxEnergy {
			c.Energy = c.MaxEnergy
		}
	}

	return c.ActionFrames(combat.ActionCharge, p)
}

//p = 0 for no hold, p = 1 for hold
func (c *char) Skill(p int) int {
	if c.CD[combat.SkillCD] > c.S.F {
		c.S.Log.Debugf("\t Lisa skill still on CD; skipping")
		return 0
	}
	if p == 0 {
		return c.skillHold()
	}
	return c.skillPress()
}

//TODO: how long do stacks last?
func (c *char) skillPress() int {
	d := c.Snapshot("Violet Arc", combat.ActionSkill, combat.Electro, combat.WeakDurability)
	d.Mult = skillPress[c.TalentLvlSkill()]
	//add 1 stack if less than 3
	if c.Tags["stack"] < 3 {
		c.Tags["stack"]++
	}
	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Lisa skill dealt %.0f damage [%v]", damage, str)

	}, "Lisa-Skill-0", 1) //TODO: frames

	c.CD[combat.SkillCD] = c.S.F + 60
	return c.ActionFrames(combat.ActionSkill, 0)
}

func (c *char) skillHold() int {
	d := c.Snapshot("Violet Arc", combat.ActionSkill, combat.Electro, combat.WeakDurability)
	d.Mult = skillHold[c.TalentLvlSkill()][c.Tags["stack"]]
	c.Tags["stack"] = 0 //consume all stacks

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Lisa skill dealt %.0f damage [%v]", damage, str)
	}, "Lisa-Skill-0", 1) //TODO: frames

	f := c.ActionFrames(combat.ActionSkill, 1)
	//c2 add defense? no interruptions either way
	if c.Base.Cons >= 2 {
		//increase def for the duration of this abil in however many frames
		val := make(map[combat.StatType]float64)
		val[combat.DEFP] = 0.25
		c.AddMod("c2", val)
		//add hook to remove this
		c.AddTask(func() {
			c.RemoveMod("c24")
		}, "c2", c.S.F+f)
	}

	c.CD[combat.SkillCD] = c.S.F + 960
	return f
}

func (c *char) Burst(p int) int {
	if c.CD[combat.BurstCD] > c.S.F {
		c.S.Log.Debugf("\t Lisa burst still on CD; skipping")
		return 0
	}

	for i := 0; i < 15; i++ { //TODO: how many hits?
		d := c.Snapshot("Lightning Rose", combat.ActionBurst, combat.Electro, combat.WeakDurability)
		d.Mult = burst[c.TalentLvlBurst()]
		t := i + 1
		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Lisa burst hit %v dealt %.0f damage [%v]", t, damage, str)
		}, fmt.Sprintf("Lisa-Skill-%v", t), i*60) //TODO: frames
	}

	c.Energy = 0
	c.CD[combat.BurstCD] = c.S.F + 1200
	return 1 //TODO: frames
}
