package chongyun

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("chongyun", NewChar)
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

	c.skillHookInfuse()
	c.a2()

	return &c, nil
}

func (c *char) Skill(p int) int {
	if c.CD[combat.SkillCD] > c.S.F {
		c.S.Log.Debugf("\t Chongyun skill still on CD; skipping")
		return 0
	}
	//TODO: check durability
	d := c.Snapshot("Spirit Blade: Chonghua's Layered Frost", combat.ActionSkill, combat.Cryo, combat.MedDurability)
	//TODO: multiplier
	d.Mult = 1

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Chongyun skill hit dealt %.0f damage [%v]", damage, str)
	}, "Chongyun-Skill", 1) //TODO: frames

	//TODO: energy count
	c.S.AddEnergyParticles("Chongyun", 2, combat.Cryo, 100)

	//a4
	c.S.AddTask(func(s *combat.Sim) {
		//TODO: check durability
		d := c.Snapshot("Spirit Blade: Chonghua's Layered Frost (A4)", combat.ActionSpecialProc, combat.Cryo, combat.MedDurability)
		//TODO: mult
		d.Mult = 1
		damage, str := s.ApplyDamage(d)
		//add res mod after dmg
		s.Target.AddResMod("Chongyun A4", combat.ResistMod{
			Duration: 480, //10 seconds
			Ele:      combat.Cryo,
			Value:    -0.10,
		})
		s.Log.Infof("\t Chongyun A4 proc dealt %.0f damage [%v]", damage, str)
	}, "Chongyun-Skill", 1+600) //TODO: frames

	c.CD[combat.SkillCD] = c.S.F + 900
	return 0 //TODO: frame count
}

func (c *char) skillHookInfuse() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if !c.S.StatusActive("Chongyun-Skill-Infuse") {
			return false
		}
		active := c.S.Chars[c.S.ActiveIndex].WeaponClass()
		if active != combat.WeaponClassClaymore && active != combat.WeaponClassSpear && active != combat.WeaponClassSword {
			return false
		}
		if ds.AbilType != combat.ActionAttack && ds.AbilType != combat.ActionCharge && ds.AbilType != combat.ActionHighPlunge && ds.AbilType != combat.ActionLowPlunge {
			return false
		}
		ds.Element = combat.Cryo
		return false
	}, "Chongyun-Skill-Infuse", combat.PostSnapshot)
}

func (c *char) a2() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if !c.S.StatusActive("Chongyun-Skill-Field") {
			return false
		}
		active := c.S.Chars[c.S.ActiveIndex].WeaponClass()
		if active != combat.WeaponClassClaymore && active != combat.WeaponClassSpear && active != combat.WeaponClassSword {
			return false
		}
		//TODO: check if this works correctly
		ds.Stats[combat.AtkSpd] += 0.15
		return false
	}, "Chongyun-Skill-Field", combat.PostSnapshot)
}

func (c *char) Burst(p int) int {
	delay := []int{1, 2, 3} //TODO: frames

	for i, v := range delay {
		//TODO: durability
		d := c.Snapshot(
			"Spirit Blade: Cloud-Parting Star",
			combat.ActionBurst,
			combat.Cryo,
			combat.MedDurability,
		)
		//TODO:  mult
		d.Mult = 1
		t := i + 1
		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Chongyun burst hit %v dealt %.0f damage [%v]", t, damage, str)
		}, fmt.Sprintf("Chongyun-Burst-Hit-%v", t), v)
	}

	c.CD[combat.BurstCD] = c.S.F + 900
	c.Energy = 0
	return 1 //TODO: frames
}
