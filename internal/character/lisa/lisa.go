package lisa

import (
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
	c.Weapon.Class = combat.WeaponClassClaymore

	return &c, nil
}

func (c *char) Attack(p int) int {

	return 0
}

func (c *char) Skill(p int) int {
	if c.CD[combat.SkillCD] > c.S.F {
		c.S.Log.Debugf("\t Lisa skill still on CD; skipping")
		return 0
	}

	c.CD[combat.SkillCD] = c.S.F + 360
	return 1 //TODO: frames
}

func (c *char) Burst(p int) int {

	c.CD[combat.SkillCD] = c.S.F + 900
	return 1 //TODO: frames
}
