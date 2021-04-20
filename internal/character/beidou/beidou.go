package beidou

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("beidou", NewChar)
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

/**
Counterattacking with Tidecaller at the precise moment when the character is hit grants the maximum DMG Bonus.

Gain the following effects for 10s after unleashing Tidecaller with its maximum DMG Bonus:
• DMG dealt by Normal and Charged Attacks is increased by 15%. ATK SPD of Normal and Charged Attacks is increased by 15%.
• Greatly reduced delay before unleashing Charged Attacks.

c1
When Stormbreaker is used:
Creates a shield that absorbs up to 16% of Beidou's Max HP for 15s.
This shield absorbs Electro DMG 250% more effectively.

c2
Stormbreaker's arc lightning can jump to 2 additional targets.

c3
Within 10s of taking DMG, Beidou's Normal Attacks gain 20% additional Electro DMG.

c6
During the duration of Stormbreaker, the Electro RES of surrounding opponents is decreased by 15%.
**/

func (c *char) Attack(p int) int {
	frames := []int{29, 21, 40, 45, 31}
	delay := []int{40, 40, 40, 40, 40}
	return c.CharacterTemplate.AttackHelperSingle(frames, delay, auto)
}

func (c *char) Skill(p int) int {
	if c.CD[combat.SkillCD] > c.S.F {
		c.S.Log.Debugf("\t Beidou skill still on CD; skipping")
		return 0
	}

	c.CD[combat.SkillCD] = c.S.F + 360
	return 1 //TODO: frames
}

func (c *char) Burst(p int) int {

	c.CD[combat.SkillCD] = c.S.F + 900
	return 1 //TODO: frames
}
