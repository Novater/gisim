package skyward

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("Skyward Pride", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	//add passive crit, atk speed not sure how to do right now??
	//looks like jsut reduce the frames of normal attacks by 1 + 12%
	m := make(map[combat.StatType]float64)

	bonus := 0.08
	dmg := .8
	switch r {
	case 2:
		dmg = 1
		bonus = 0.1
	case 3:
		dmg = 1.2
		bonus = 0.12
	case 4:
		dmg = 1.4
		bonus = 0.14
	case 5:
		dmg = 1.6
		bonus = 0.16
	}
	m[combat.DmgP] = bonus

	c.AddMod("Skyward Pride Stats", m)

	counter := 0

	s.AddEventHook(func(s *combat.Sim) bool {
		//check if char is correct?
		if s.ActiveChar != c.Name() {
			return false
		}
		//20s timer
		s.Status["Skyward Pride Proc"] = 20 * 60
		counter = 0

		return false
	}, "skyward-pride-proc", combat.PostBurstHook)

	//add on hit effect to sim?
	s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		//check if char is correct?
		if ds.CharName != c.Name() {
			return false
		}
		//check if this is normal or charged
		if ds.AbilType != combat.ActionTypeAttack && ds.AbilType != combat.ActionTypeChargedAttack {
			return false
		}
		//check if buff is active
		if _, ok := s.Status["Skyward Pride Proc"]; !ok {
			return false
		}
		//check if already done 8 hits
		if counter > 8 {
			return false
		}

		counter++
		//add a new action that deals % dmg immediately
		d := c.Snapshot("Skyward Pride Proc", combat.ActionTypeSpecialProc, combat.Physical)
		d.Mult = dmg
		s.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(d)
			s.Log.Infof("\t Skyward Pride proc dealt %.0f damage", damage)
		}, fmt.Sprintf("Skyward Pride Proc (hit %v) %v", counter, c.Name()), 1)

		return false
	}, "skyward-spine-proc", combat.PostDamageHook)
}
