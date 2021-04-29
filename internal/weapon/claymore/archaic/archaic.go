package archaic

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("prototype archaic", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	atk := 1.8 + float64(r)*.6
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
		if s.F-last < 900 {
			return false
		}
		if s.Rand.Float64() > .5 {
			return false
		}
		s.Log.Infof("[%v] archaic proc'd", s.Frame())

		//add a new action that deals % dmg immediately
		d := c.Snapshot("Prototype Archaic Proc", combat.ActionSpecialProc, combat.Physical, combat.WeakDurability)
		d.Mult = atk
		s.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Prototype Archaic proc dealt %.0f damage [%v]", damage, str)
		}, fmt.Sprintf("Prototype Archaic Proc %v", c.Name()), 1)

		last = s.F

		return false
	}, "prototype-archaic", combat.PostDamageHook)

}
