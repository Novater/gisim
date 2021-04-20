package kaeya

import (
	"fmt"

	"github.com/srliao/gisim/internal/rotation"
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("kaeya", NewChar)
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
	c.Weapon.Class = combat.WeaponClassClaymore

	c.a4()

	if p.Base.Cons >= 1 {
		c.c1()
	}
	if p.Base.Cons == 6 {
		c.c6()
	}

	return &c, nil
}

func (c *char) c1() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor != c.Name() {
			return false
		}
		if ds.AbilType != rotation.ActionAttack && ds.AbilType != rotation.ActionCharge {
			return false
		}
		if c.S.TargetAura.E() != combat.Cryo && c.S.TargetAura.E() != combat.Frozen {
			return false
		}

		ds.Stats[combat.CR] += .2
		return false
	}, "kaeya-c1", combat.PreDamageHook)
}

func (c *char) c6() {
	c.S.AddEventHook(func(s *combat.Sim) bool {
		//return 15 energy
		if s.ActiveChar != c.Name() {
			return false
		}
		c.Energy += 15
		if c.Energy > c.MaxEnergy {
			c.Energy = c.MaxEnergy
		}
		return false
	}, "kaeya-c6", combat.PostBurstHook)
}

func (c *char) Attack(p int) int {

	frames := []int{29, 21, 40, 45, 31}
	delay := []int{40, 40, 40, 40, 40}
	return c.CharacterTemplate.AttackHelperSingle(frames, delay, auto)
}

func (c *char) a4() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor != c.Name() {
			return false
		}
		if ds.Abil != "Frostgnaw" {
			return false
		}
		if !c.S.GlobalFlags.ReactionDidOccur {
			return false
		}
		if c.S.GlobalFlags.ReactionType != combat.Freeze {
			return false
		}

		//TODO: energy count
		c.S.AddEnergyParticles("Kaeya", 2, combat.Cryo, 100)

		return false
	}, "kaeya-a4", combat.PreReaction)
}

func (c *char) Skill(p int) int {
	if c.CD[combat.SkillCD] > c.S.F {
		c.S.Log.Debugf("\t Kaeya skill still on CD; skipping")
		return 0
	}
	d := c.Snapshot("Frostgnaw", rotation.ActionSkill, combat.Cryo, combat.MedDurability)
	d.Mult = 1 //TODO: multiplier

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Kaeya skill hit dealt %.0f damage [%v]", damage, str)
	}, "Kaeya-Skill", 1) //TODO: frames

	//TODO: energy count
	c.S.AddEnergyParticles("Kaeya", 2, combat.Cryo, 100)

	c.CD[combat.SkillCD] = c.S.F + 360
	return 1 //TODO: frames
}

func (c *char) Burst(p int) int {

	d := c.Snapshot("Frostgnaw", rotation.ActionSkill, combat.Cryo, combat.MedDurability)
	d.Mult = 1 //TODO: multiplier

	max := 5 //TODO: number of hits, also C2 ignored
	for i := 0; i < max; i++ {
		t := i + 1
		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Kaeya burst hit %v dealt %.0f damage [%v]", t, damage, str)
		}, fmt.Sprintf("Kaeya-Burst-Hit-%v", t), i*5) //TODO: frames
	}

	c.CD[combat.SkillCD] = c.S.F + 900
	return 1 //TODO: frames
}
