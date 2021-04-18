package skyward

import (
	"github.com/srliao/gisim/internal/rotation"
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("Song of Broken Pines", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	m := make(map[combat.StatType]float64)
	m[combat.ATKP] = 0.12 + 0.04*float64(r)
	c.AddMod("Song of Broken Pines Stats", m)

	bs := 0.09 + 0.03*float64(r)
	ba := 0.15 + 0.05*float64(r)

	stacks := 0
	isActive := false
	cd := 0
	dur := 0
	lock := 0

	//post snap shot to increase stacks
	s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor != c.Name() {
			return false
		}
		if ds.AbilType != rotation.ActionAttack && ds.AbilType != rotation.ActionCharge {
			return false
		}
		if lock > 0 {
			return false
		}
		if cd > 0 {
			return false //no new stacks while on 20s CD
		}
		//every exectuion, add 1 stack, to a max of 4
		stacks++
		s.Log.Debugf("\t broken pines adding stacks: %v", stacks)
		if stacks == 4 {
			//trigger the effect
			s.Log.Debugf("\t broken pines at 4 stacks, effect triggered")
			stacks = 0
			cd = 1200 //Once this effect is triggered, you will not gain Sigils of Whispers for 20s. Of the many effects of the "Millennial Movement," buffs of the same type will not stack.
			isActive = true
			dur = 720 //all nearby party members will obtain the "Millennial Movement: Banner-Hymn" effect for 12s.
		}
		lock = 18 // This effect can be triggered once every 0.3s.
		return false
	}, "broken pines stacks", combat.PostDamageHook)

	//counter to reduce duration
	s.AddEffect(func(s *combat.Sim) bool {
		//tick down cd
		dur--
		if dur < 0 {
			dur = 0
			isActive = false
		}
		cd--
		if cd < 0 {
			cd = 0
		}
		lock--
		if lock < 0 {
			lock = 0
		}
		return false
	}, "broken pines dur")

	s.AddSnapshotHook(func(snap *combat.Snapshot) bool {

		if !isActive {
			return false
		}

		snap.Stats[combat.ATKP] += ba
		snap.Stats[combat.AtkSpd] += bs

		s.Log.Debugf("\t skyrider greatsword adding atkp: %v atkspd: %v", ba, bs)

		return false
	}, "broken pines stats", combat.PostSnapshot)

}
