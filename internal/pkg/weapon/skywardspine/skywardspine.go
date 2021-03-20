package skywardspine

import (
	"fmt"
	"math/rand"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterWeaponFunc("Skyward Spine", weapon)
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
	s.AddEffect(func(snap *combat.Snapshot) bool {
		//check if char is correct?
		if snap.CharName != c.Name() {
			return false
		}
		//check if this is normal or charged
		if snap.AbilType != combat.ActionTypeAttack && snap.AbilType != combat.ActionTypeChargedAttack {
			return false
		}
		//check if cd is up
		if _, ok := s.Status["Skyward Spine Proc"]; ok {
			return false
		}
		//check 50/50 proc chance
		r := rand.Intn(2)
		if r == 0 {
			return false
		}

		//add a new action that deals % dmg immediately
		d := c.Snapshot(combat.Physical)
		d.Abil = "Skyward Spine Proc"
		d.AbilType = combat.ActionTypeWeaponProc
		d.Mult = dmg
		s.AddAction(func(s *combat.Sim) bool {
			damage := s.ApplyDamage(d)
			s.Log.Infof("[%v]: Skyward Spine proc dealt %.0f damage", s.Frame(), damage)
			return true
		}, fmt.Sprintf("Skyware-Spine-Proc-%v", c.Name()))
		//trigger cd
		s.Status["Skyward Spine Proc"] = 2 * 60

		return false
	}, "skyward-spine-proc", combat.PostDamageHook)
}
