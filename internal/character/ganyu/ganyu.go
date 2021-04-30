package ganyu

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("ganyu", NewChar)
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
	c.Weapon.Class = combat.WeaponClassBow
	c.a2()

	if c.Base.Cons >= 1 {
		c.c1()
	}

	return &c, nil
}

func (c *char) c1() {
	s := c.S
	s.Log.Debugf("\tactivating Ganyu C1")

	s.AddSnapshotHook(func(snap *combat.Snapshot) bool {
		if snap.Actor != c.Base.Name {
			return false
		}
		if snap.Abil != "Frost Flake Arrow" {
			return false
		}

		//if c1, increase character energy by 2, unaffected by ER; assume assuming arrow always hits here
		c.Energy += 2
		if c.Energy > c.MaxEnergy {
			c.Energy = c.MaxEnergy
		}
		s.Log.Debugf("\t Ganyu C1 refunding 2 energy; current energy %v", c.Energy)
		//also add c1 debuff to target
		s.Target.AddResMod("ganyu-c1", combat.ResistMod{
			Ele:      combat.Cryo,
			Value:    -0.15,
			Duration: 5 * 60,
		})

		return false
	}, "ganyu-c1", combat.PostDamageHook)
}

func (c *char) a2() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor != c.Base.Name {
			return false
		}
		if ds.AbilType != combat.ActionAim {
			return false
		}
		if c.CD["A2"] <= c.S.F {
			return false
		}
		ds.Stats[combat.CR] += 0.2
		c.S.Log.Debugf("\t applying Ganyu a2, new crit %v", ds.Stats[combat.CR])

		return false
	}, "ganyu-a2", combat.PreDamageHook)
}

func (c *char) ActionFrames(a combat.ActionType, p int) int {
	switch a {
	case combat.ActionAttack:
		f := 0
		switch c.NormalCounter {
		//TODO: need to add atkspd mod
		case 0:
			f = 18 //frames from keqing lib
		case 1:
			f = 43 - 18
		case 2:
			f = 73 - 43
		case 3:
			f = 117 - 73
		case 4:
			f = 153 - 117
		case 5:
			f = 190 - 153
		}
		f = int(float64(f) / (1 + c.Stats[combat.AtkSpd]))
		return f
	case combat.ActionAim:
		return 115 //frames from keqing lib
	case combat.ActionSkill:
		return 30 //ok
	case combat.ActionBurst:
		return 122 //ok
	default:
		c.S.Log.Warnf("%v: unknown action, frames invalid", a)
		return 0
	}
}

func (c *char) Attack(p int) int {
	d := c.Snapshot("Normal", combat.ActionAttack, combat.Physical, 0)
	d.Mult = attack[c.NormalCounter][c.TalentLvlAttack()]
	f := c.ActionFrames(combat.ActionAttack, p)
	delay := f + 40 //TODO: frames

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t %v normal %v dealt %.0f damage, dur %v [%v]", c.Base.Name, c.NormalCounter, damage, d.Durability, str)
	}, fmt.Sprintf("%v-Normal-%v", c.Base.Name, c.NormalCounter), delay)

	c.NormalResetTimer = 150
	c.NormalCounter++
	if c.NormalCounter == 6 {
		c.NormalCounter = 0
		c.NormalResetTimer = 0
	}

	return f
}

func (c *char) Aimed(p int) int {
	f := c.Snapshot("Frost Flake Arrow", combat.ActionAim, combat.Cryo, combat.WeakDurability)
	f.HitWeakPoint = true
	f.Mult = ffa[c.TalentLvlAttack()]

	b := c.Snapshot("Frost Flake Bloom", combat.ActionAim, combat.Cryo, combat.WeakDurability)
	b.Mult = ffb[c.TalentLvlAttack()]

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(f)
		s.Log.Infof("\t Ganyu frost arrow dealt %.0f damage [%v]", damage, str)
		//apply A2 on hit
		c.CD["A2"] = c.S.F + 5*60
	}, "Ganyu-Aimed-FFA", 20+137)

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(b)
		s.Log.Infof("\t Ganyu frost flake bloom dealt %.0f damage [%v]", damage, str)
		//apply A2 on hit
		c.CD["A2"] = c.S.F + 5*60
	}, "Ganyu-Aimed-FFB", 20+20+137)

	return c.ActionFrames(combat.ActionAim, p)
}

func (c *char) Skill(p int) int {
	//if c2, check if either cd is cooldown
	charge := ""
	c2onCD := c.CD["skill-cd-2"] > c.S.F
	onCD := c.CD[charge] > c.S.F

	if c.Base.Cons >= 2 {
		if !c2onCD {
			charge = "skill-cd-2"
		}
	}
	if !onCD {
		charge = "skill-cd"
	}

	if charge == "" {
		c.S.Log.Debugf("\tGanyu skill still on CD; skipping")
		return 0
	}

	//snap shot stats at cast time here
	d := c.Snapshot("Ice Lotus", combat.ActionSkill, combat.Cryo, combat.WeakDurability)
	d.Mult = lotus[c.TalentLvlSkill()]

	//we get the orbs right away
	c.S.AddEnergyParticles("Ganyu", 2, combat.Cryo, 90) //90s travel time

	//flower damage is after 6 seconds
	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Ganyu ice lotus dealt %.0f damage [%v]", damage, str)
	}, "Ganyu Flower", 6*60)

	//add cooldown to sim
	c.CD[charge] = c.S.F + 10*60

	return 30
}

func (c *char) Burst(p int) int {
	//check if on cd first
	if c.CD[combat.BurstCD] > c.S.F {
		c.S.Log.Debugf("\tGanyu burst still on CD; skipping")
		return 0
	}
	//check if sufficient energy
	if c.Energy < c.MaxEnergy {
		c.S.Log.Debugf("\tGanyu burst - insufficent energy, current: %v", c.Energy)
		return 0
	}
	//snap shot stats at cast time here
	d := c.Snapshot("Celestial Shower", combat.ActionBurst, combat.Cryo, combat.WeakDurability)
	d.Mult = shower[c.TalentLvlBurst()]

	for delay := 120; delay <= 900; delay += 60 {
		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Ganyu burst (tick) dealt %.0f damage [%v]", damage, str)
		}, "Ganyu Burst", delay)
	}

	//add cooldown to sim
	c.CD[combat.BurstCD] = c.S.F + 15*60
	//use up energy
	c.Energy = 0

	return 122
}

func (c *char) ActionReady(a combat.ActionType) bool {
	switch a {
	case combat.ActionBurst:
		if c.Energy != c.MaxEnergy {
			return false
		}
		return c.CD[combat.BurstCD] <= c.S.F
	case combat.ActionSkill:
		skillReady := c.CD[combat.SkillCD] <= c.S.F
		//if skill ready return true regardless of c2
		if skillReady {
			return true
		}
		//other wise skill-cd is there, we check c2
		if c.Base.Cons >= 2 {
			return c.CD["skill-cd2"] <= c.S.F
		}
		return false
	}
	return true
}

func (g *char) Tick() {
	g.CharacterTemplate.Tick()
}
