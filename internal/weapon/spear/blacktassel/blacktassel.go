package blacktassel

import "github.com/srliao/gisim/pkg/combat"

func init() {
	combat.RegisterWeaponFunc("black tassel", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	//add on hit effect to sim?

}
