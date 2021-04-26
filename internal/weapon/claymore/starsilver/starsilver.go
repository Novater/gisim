package starsilver

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("Snow-Tombed Starsilver", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	atk := .65 + float64(r)*.15
	atkUp := 1.6 + float64(r)*.4
	p := .5 + float64(r)*.1
	last := 0

	//add on crit effect
	s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		//check if char is correct?
		if ds.Actor != c.Name() {
			return false
		}
		if ds.AbilType != combat.ActionAttack && ds.AbilType != combat.ActionCharge {
			return false
		}
		if s.F-last < 600 {
			return false
		}
		if s.Rand.Float64() > p {
			return false
		}
		s.Log.Infof("[%v] starsilver proc'd", s.Frame())

		dmg := atk
		//check if affected by cryo
		if s.TargetAura.E() == combat.Cryo || s.TargetAura.E() == combat.Frozen {
			dmg = atkUp
		}

		//add a new action that deals % dmg immediately
		d := c.Snapshot("Starsilver Proc", combat.ActionSpecialProc, combat.Physical, combat.WeakDurability)
		d.Mult = dmg
		s.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Starsilver proc dealt %.0f damage [%v]", damage, str)
		}, fmt.Sprintf("Starsilver Proc %v", c.Name()), 1)

		last = s.F

		return false
	}, "starsilver", combat.PostDamageHook)

}
