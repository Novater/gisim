package sacrificialsword

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("Sacrificial Sword", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {

	prob := 0.4
	cd := 30 * 60
	switch r {
	case 2:
		prob = .5
		cd = 26 * 60
	case 3:
		prob = .6
		cd = 22 * 60
	case 4:
		prob = .7
		cd = 19 * 60
	case 5:
		prob = .8
		cd = 16 * 60
	}
	//add on crit effect
	s.AddSnapshotHook(func(snap *combat.Snapshot) bool {
		//check if char is correct?
		if snap.CharName != c.Name() {
			return false
		}
		if snap.AbilType != combat.ActionTypeSkill {
			return false
		}
		if s.StatusActive("Sacrificial Sword Proc") {
			return false
		}
		if s.Rand.Float64() > prob {
			return false
		}
		s.Log.Debugf("\t Sacrificial Sword proc triggered")

		c.ResetActionCooldown(combat.ActionTypeSkill)

		s.Status["Sacrificial Sword Proc"] = s.F + cd
		return false
	}, "sacrificial-sword-proc", combat.PostDamageHook)
}
