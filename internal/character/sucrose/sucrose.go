package sucrose

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("sucrose", NewChar)
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

func (c *char) a2() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if c.Tags["a2-pyro"] >= c.S.F && ds.ActorEle == combat.Pyro {
			ds.Stats[combat.EM] += 50
			return false
		}
		if c.Tags["a2-cryo"] >= c.S.F && ds.ActorEle == combat.Cryo {
			ds.Stats[combat.EM] += 50
			return false
		}
		if c.Tags["a2-hydro"] >= c.S.F && ds.ActorEle == combat.Hydro {
			ds.Stats[combat.EM] += 50
			return false
		}
		if c.Tags["a2-electro"] >= c.S.F && ds.ActorEle == combat.Electro {
			ds.Stats[combat.EM] += 50
			return false
		}
		return false
	}, "sucrose-a2-buff", combat.PostSnapshot)

	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor != c.Base.Name {
			return false
		}
		switch c.S.GlobalFlags.ReactionType {
		case combat.SwirlCryo:
			c.Tags["a2-cryo"] = c.S.F + 480
		case combat.SwirlElectro:
			c.Tags["a2-electro"] = c.S.F + 480
		case combat.SwirlHydro:
			c.Tags["a2-hydro"] = c.S.F + 480
		case combat.SwirlPyro:
			c.Tags["a2-pyro"] = c.S.F + 480
		}
		return false
	}, "sucrose-a2-buff", combat.PostReaction)
}

func (c *char) a4() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor == c.Base.Name {
			return false
		}
		if c.Tags["a4"] >= c.S.F {
			ds.Stats[combat.EM] += c.Stats[combat.EM] * 0.2
		}
		return false
	}, "sucrose-a2-buff", combat.PostSnapshot)
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

func (c *char) Burst(p int) int {
	//tag a4

	return 0
}
