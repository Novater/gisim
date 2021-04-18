package diona

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Diona", NewChar)
}

type ningguang struct {
	*combat.CharacterTemplate
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	n := ningguang{}
	t, err := combat.NewTemplateChar(s, p)
	if err != nil {
		return nil, err
	}
	n.CharacterTemplate = t
	n.Energy = 40
	n.MaxEnergy = 40
	n.Weapon.Class = combat.WeaponClassClaymore

	return &n, nil
}

func (n *ningguang) Attack(p int) int {

	return 0
}
