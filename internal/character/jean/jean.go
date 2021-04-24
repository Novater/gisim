package jean

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("jean", NewChar)
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
	c.Weapon.Class = combat.WeaponClassSword

	if c.Base.Cons >= 2 {
		c.c2()
	}

	return &c, nil
}

func (c *char) Attack(p int) int {
	frames := []int{29, 21, 40, 45, 31} //TODO: frames
	delay := []int{40, 40, 40, 40, 40}  //TODO: frames
	return c.CharacterTemplate.AttackHelperSingle(frames, delay, auto)
}

func (c *char) Skill(p int) int {
	if c.CD[combat.SkillCD] > c.S.F {
		c.S.Log.Debugf("\t Jean skill still on CD; skipping")
		return 0
	}
	//hold for p up to 5 seconds
	if p > 5 {
		p = 5
	}

	d := c.Snapshot("Gale Blade", combat.ActionSkill, combat.Anemo, combat.MedDurability)
	d.Mult = skill[c.TalentLvlSkill()]

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Jean skill on cast dealt %.0f damage [%v]", damage, str)
	}, "Jean-Gale-Blade", 0+p*60) //TODO: damage delay on initial cast

	c.CD[combat.SkillCD] = c.S.F + 360 //6 seconds
	return 1 + p*60                    //TODO: frames, + p for holding
}

func (c *char) Burst(p int) int {
	if c.CD[combat.BurstCD] > c.S.F {
		c.S.Log.Debugf("\t Jean burst still on CD; skipping")
		return 0
	}
	//p is the number of times enemy enters or exits the field

	d := c.Snapshot("Dandelion Breeze", combat.ActionBurst, combat.Anemo, combat.MedDurability)
	d.Mult = burst[c.TalentLvlBurst()]

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Jean burst on cast dealt %.0f damage [%v]", damage, str)
	}, "Jean-Dandelion-Breeze", 0) //TODO: damage delay on initial cast

	for i := 0; i < p; i++ {
		d := c.Snapshot("Dandelion Breeze", combat.ActionBurst, combat.Anemo, combat.MedDurability)
		d.Mult = burstEnter[c.TalentLvlBurst()]
		t := i + 1
		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Jean burst on entry/exit %v out of %v dealt %.0f damage [%v]", t, p, damage, str)
		}, fmt.Sprintf("Jean-Dandelion-Breeze-Entry-%v", i), 0) //TODO: damage delay on initial cast
	}

	c.S.Status["Jean Dandelion Field"] = c.S.F + 600 //TODO: how long does her field last?
	if c.Base.Cons >= 4 {
		//add debuff to target for ??? duration
		c.S.Target.AddResMod("Jean C4", combat.ResistMod{
			Duration: 600, //10 seconds
			Ele:      combat.Anemo,
			Value:    -0.4,
		})
	}

	c.CD[combat.SkillCD] = c.S.F + 1200 //6 seconds
	c.Energy = 16                       //jean a4
	return 1                            //TODO: frames, + p for holding
}

func (c *char) ReceiveParticle(p combat.Particle, isActive bool, partyCount int) {
	c.CharacterTemplate.ReceiveParticle(p, isActive, partyCount)
	if c.Base.Cons >= 2 {
		//only pop this if jean is active
		if !isActive {
			return
		}
		c.S.Status["Jean C2"] = c.S.F + 900
	}
}

func (c *char) c2() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if !c.S.StatusActive("Jean C2") {
			return false
		}
		//other wise add only atkspd since no one moves
		ds.Stats[combat.AtkSpd] += 0.15
		return false
	}, "jean-c2", combat.PostSnapshot)
}
