package paleflame

import (
	"github.com/srliao/gisim/internal/rotation"
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("Pale Flame", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	if count >= 2 {
		m := make(map[combat.StatType]float64)
		m[combat.PhyP] = 0.25
		c.AddMod("Pale Flame 2PC", m)
	}
	stacks := 0
	dur := 0
	lock := 0
	if count >= 4 {
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
		}, "Pale Flame 4PC")

		//post snap shot to increase stacks
		s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
			if ds.Actor != c.Name() {
				return false
			}
			if ds.AbilType != rotation.ActionSkill {
				return false
			}
			if lock > 0 {
				return false
			}
			//every exectuion, add 1 stack, to a max of 3, reset cd to 10 seconds
			stacks++
			s.Log.Debugf("\t Pale Flame 4 adding stacks: %v", stacks)
			if stacks > 3 {
				stacks = 3
			}
			dur = 420 //7 seconds
			lock = 18 // can only trigger once every 18 frames
			return false
		}, "Pale Flame 4PC", combat.PostDamageHook)

		s.AddSnapshotHook(func(snap *combat.Snapshot) bool {
			if snap.Actor != c.Name() {
				return false
			}

			atk := 0.06 * float64(stacks)
			snap.Stats[combat.ATKP] += atk
			s.Log.Debugf("\t Pale Flame 4 adding attack: %v", atk)

			if stacks == 3 {
				snap.Stats[combat.PhyP] += 0.25 //double the phys bonus
				s.Log.Debugf("\t Pale Flame 4 adding bonus phy %")
			}

			return false
		}, "Pale Flame 4PC", combat.PostSnapshot)
	}
	//add flat stat to char
}
