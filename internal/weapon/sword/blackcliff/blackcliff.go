package blackcliff

import "github.com/srliao/gisim/pkg/combat"

func init() {
	combat.RegisterWeaponFunc("blackcliff longsword", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	//nothing ever dies...
}
