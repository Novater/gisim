package blacktassel

import "github.com/srliao/gisim/internal/pkg/combat"

func init() {
	combat.RegisterWeaponFunc("Black Tassel", weapon)
}

func weapon(c *combat.Char, s *combat.Sim, r int) {
	//add on hit effect to sim?

}
