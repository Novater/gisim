package skyward

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("skyrider greatsword", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	bonus := 0.05 + 0.01*float64(r)
	stacks := 0
	dur := 0
	lock := 0

	//counter to reduce duration
	s.AddEffect(func(s *combat.Sim) bool {
		//tick down cd
		dur--
		if dur < 0 {
			dur = 0
			stacks = 0 //stacks all gone if duration is gone
		}
		lock--
		if lock < 0 {
			lock = 0
		}
		return false
	}, "skyrider greatsword")

	//post snap shot to increase stacks
	s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor != c.Name() {
			return false
		}
		if ds.AbilType != combat.ActionAttack && ds.AbilType != combat.ActionCharge {
			return false
		}
		if lock > 0 {
			return false
		}
		//every exectuion, add 1 stack, to a max of 3, reset cd to 10 seconds
		stacks++
		s.Log.Debugf("\t skyrider greatsword adding stacks: %v", stacks)
		if stacks > 3 {
			stacks = 3
		}
		dur = 360 //6 seconds
		lock = 30 // can only trigger once every 18 frames
		return false
	}, "skyrider greatsword", combat.PostDamageHook)

	s.AddSnapshotHook(func(snap *combat.Snapshot) bool {
		if snap.Actor != c.Name() {
			return false
		}

		snap.Stats[combat.ATKP] += bonus
		s.Log.Debugf("\t skyrider greatsword adding attack: %v", bonus)

		return false
	}, "skyrider greatsword", combat.PostSnapshot)

}
