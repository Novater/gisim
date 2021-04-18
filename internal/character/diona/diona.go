package diona

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Diona", NewChar)
}

type diona struct {
	*combat.CharacterTemplate
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	d := diona{}
	t, err := combat.NewTemplateChar(s, p)
	if err != nil {
		return nil, err
	}
	d.CharacterTemplate = t
	d.Energy = 60
	d.MaxEnergy = 60
	d.Weapon.Class = combat.WeaponClassClaymore

	return &d, nil
}

func (d *diona) Attack(p int) int {

	frames := []int{29, 21, 40, 45, 31}
	delay := []int{40, 40, 40, 40, 40}
	return d.CharacterTemplate.AttackHelperSingle(frames, delay, auto)
}
