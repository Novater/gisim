package skyward

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("Wolf's Gravestone", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {

	m := make(map[combat.StatType]float64)
	m[combat.ATKP] = 0.15 + 0.05*float64(r)
	c.AddMod("Wolf's Gravestone Stats", m)

	bonus := 0.3 + 0.1*float64(r)
	dur := 0
	lock := 0

	//counter to reduce duration
	s.AddEffect(func(s *combat.Sim) bool {
		//tick down cd
		dur--
		if dur < 0 {
			dur = 0
		}
		lock--
		if lock < 0 {
			lock = 0
		}
		return false
	}, "wolf's gravestone")

	//check hp
	s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor != c.Name() {
			return false
		}
		if lock > 0 {
			return false
		}
		if !s.Target.HPMode {
			return false //ignore as we not tracking HP
		}
		if s.Target.HP/s.Target.MaxHP > 0.3 {
			return false
		}

		dur = 720   //12 seconds
		lock = 1800 // can only occur once every 30s
		return false
	}, "wolf's gravestone", combat.PreDamageHook)

	s.AddSnapshotHook(func(snap *combat.Snapshot) bool {
		if dur <= 0 {
			return false
		}

		snap.Stats[combat.ATKP] += bonus
		s.Log.Debugf("\t wolf's gravestone adding attack: %v", bonus)

		return false
	}, "wolf's gravestone", combat.PostSnapshot)

}
