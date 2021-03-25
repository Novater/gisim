package diluc

import (
	"github.com/srliao/gisim/internal/character/common"
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Diluc", NewChar)
}

type diluc struct {
	*common.TemplateChar
	eCounter int
}

/**

Skills:
	A2: Diluc's Charged Attack Stamina Cost is decreased by 50%, and its duration is increased by 3s.

	A4: The Pyro Enchantment provided by Dawn lasts for 4s longer. Additionally, Diluc gains 20% Pyro DMG Bonus during the duration of this effect.

	C1: Diluc deals 15% more DMG to enemies whose HP is above 50%.

	C2: When Diluc takes DMG, his Base ATK increases by 10% and his ATK SPD increases by 5%. Lasts for 10s.
		This effect can stack up to 3 times and can only occur once every 1.5s.

	C4: Casting Searing Onslaught in sequence greatly increases damage dealt.
		Within 2s of using Searing Onslaught, casting the next Searing Onslaught in the combo deals 40% additional DMG. This effect lasts for the next 2s.

	C6: After casting Searing Onslaught, the next 2 Normal Attacks within the next 6s will have their DMG and ATK SPD increased by 30%.
		Additionally, Searing Onslaught will not interrupt the Normal Attack combo. <- what does this mean??

Checklist:
	- Frame count
	- Orb generation
	- Ascension bonus
	- Constellations

**/

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	d := diluc{}
	t, err := common.New(s, p)
	if err != nil {
		return nil, err
	}
	d.TemplateChar = t
	d.Energy = 60
	d.MaxEnergy = 60
	d.Profile.WeaponClass = combat.WeaponClassClaymore

	return &d, nil
}

func (d *diluc) Skill(p map[string]interface{}) int {
	return 0
}
