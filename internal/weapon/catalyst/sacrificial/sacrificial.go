package sacrificial

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("sacrificial fragments", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {

	prob := 0.3 + float64(r)*0.1
	cd := (34 - r*4) * 60
	//add on crit effect
	s.AddSnapshotHook(func(snap *combat.Snapshot) bool {
		//check if char is correct?
		if snap.Actor != c.Name() {
			return false
		}
		if snap.AbilType != combat.ActionSkill {
			return false
		}
		if s.StatusActive("sac frag") {
			return false
		}
		if c.ActionReady(combat.ActionSkill) {
			return false
		}
		if s.Rand.Float64() > prob {
			return false
		}
		s.Log.Infof("[%v] sacrificial frag proc'd", s.Frame())

		c.ResetActionCooldown(combat.ActionSkill)

		s.Status["sac frag"] = s.F + cd
		return false
	}, "sacrificial-frag-proc", combat.PostDamageHook)
}
