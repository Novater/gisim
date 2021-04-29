package skywardspine

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("skyward spine", weapon)
}

func weapon(c combat.Character, s *combat.Sim, r int) {
	//add passive crit, atk speed not sure how to do right now??
	//looks like jsut reduce the frames of normal attacks by 1 + 12%
	m := make(map[combat.StatType]float64)
	m[combat.AtkSpd] = 0.12
	dmg := .4
	switch r {
	default:
		m[combat.CR] = 0.08
	case 2:
		m[combat.CR] = 0.10
		dmg = .55
	case 3:
		m[combat.CR] = 0.12
		dmg = .70
	case 4:
		m[combat.CR] = 0.14
		dmg = .85
	case 5:
		m[combat.CR] = 0.16
		dmg = .100
	}
	c.AddMod("Skyward-Spine-Stats", m)
	//add on hit effect to sim?
	s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		//check if char is correct?
		if ds.Actor != c.Name() {
			return false
		}
		//check if this is normal or charged
		if ds.AbilType != combat.ActionAttack && ds.AbilType != combat.ActionCharge {
			return false
		}
		//check if cd is up
		if !s.StatusActive("Skyward Spine Proc") {
			return false
		}

		//check 50/50 proc chance
		r := s.Rand.Intn(2)
		if r == 0 {
			return false
		}

		//add a new action that deals % dmg immediately
		d := c.Snapshot("Skyward Spine Proc", combat.ActionSpecialProc, combat.Physical, combat.WeakDurability)
		d.Mult = dmg
		s.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Skyward Spine proc dealt %.0f damage [%v]", damage, str)
		}, fmt.Sprintf("Skyward Spine Proc %v", c.Name()), 1)

		//trigger cd
		s.Status["Skyward Spine Proc"] = s.F + 2*60

		return false
	}, "skyward-spine-proc", combat.PostDamageHook)
}
