package keqing

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("keqing", NewChar)
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

	return &c, nil
}

func (c *char) Skill(p int) int {
	//p is number of frames to wait before activating E, default is right away
	//if p = -1 then never activate it
	//if the intention is to detonate via charge attack then make sure p = -1
	//otherwise it'll ignore any charge attacks

	return 1 //TODO: frames
}
