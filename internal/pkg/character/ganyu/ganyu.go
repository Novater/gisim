package ganyu

import (
	"fmt"

	"github.com/srliao/gisim/internal/pkg/combat"
	"go.uber.org/zap"
)

func init() {
	combat.RegisterCharFunc("Ganyu", New)
}

func New(s *combat.Sim, c *combat.Character) {
	c.ChargeAttack = charge(c, s.Log)
	c.Burst = burst(c, s.Log)
	c.Skill = skill(c, s.Log)
	c.MaxEnergy = 60
	c.Energy = 60

	if c.Profile.Constellation >= 1 {
		s.Log.Debugf("\tactivating Ganyu C1")

		s.AddEffect(func(snap *combat.Snapshot) bool {
			//check if c1 debuff is on, if so, reduce resist by -0.15
			if _, ok := s.Target.Status["ganyu-c1"]; ok {
				s.Log.Debugf("\t[%v]: applying Ganyu C1 cryo debuff", combat.PrintFrames(s.Frame))
				snap.ResMod[combat.Cryo] -= 0.15
			}
			return false
		}, "ganyu-c1", combat.PreDamageHook)

		s.AddEffect(func(snap *combat.Snapshot) bool {
			if snap.CharName == "Ganyu" && snap.Abil == "Frost Flake Arrow" {
				//if c1, increase character energy by 2, unaffected by ER; assume assuming arrow always hits here
				c.Energy += 2
				if c.Energy > c.MaxEnergy {
					c.Energy = c.MaxEnergy
				}
				s.Log.Debugf("\t[%v]: Ganyu C1 refunding 2 energy; current energy %v", combat.PrintFrames(s.Frame), c.Energy)
				//also add c1 debuff to target
				s.Target.Status["ganyu-c1"] = 5 * 60
			}
			return false
		}, "ganyu-c1", combat.PostDamageHook)
	}

}

func charge(c *combat.Character, log *zap.SugaredLogger) combat.AbilFunc {
	return func(s *combat.Sim) int {
		i := 0
		initial := func(s *combat.Sim) bool {
			if i < 20 {
				i++
				return false
			}
			//abil
			d := c.Snapshot(combat.Cryo)
			d.Abil = "Frost Flake Arrow"
			d.AbilType = combat.ActionTypeChargedAttack
			d.HitWeakPoint = true
			d.Mult = ffa[c.Profile.TalentLevel[combat.ActionTypeAttack]-1]
			d.AuraGauge = 1
			d.AuraUnit = "A"
			d.ApplyAura = true
			//if not ICD, apply aura
			if _, ok := c.Cooldown["ICD-charge"]; !ok {
				d.ApplyAura = true
			}
			//check if A4 talent is
			if _, ok := c.Cooldown["A2"]; ok {
				d.Stats[combat.CR] += 0.2
			}
			c.Cooldown["A2"] = 5 * 60
			//apply damage
			damage := s.ApplyDamage(d)
			log.Infof("[%v]: Ganyu frost arrow dealt %.0f damage", combat.PrintFrames(s.Frame), damage)
			return true
		}

		b := 0
		//apply second bloom w/ more travel time
		bloom := func(s *combat.Sim) bool {
			if b < 50 {
				b++
				return false
			}
			//abil
			d := c.Snapshot(combat.Cryo)
			d.Abil = "Frost Flake Bloom"
			d.AbilType = combat.ActionTypeChargedAttack
			d.Mult = ffb[c.Profile.TalentLevel[combat.ActionTypeAttack]-1]
			d.ApplyAura = true
			d.AuraGauge = 1
			d.AuraUnit = "A"
			//if not ICD, apply aura
			if _, ok := c.Cooldown["ICD-charge"]; !ok {
				d.ApplyAura = true
			}
			if _, ok := c.Cooldown["A2"]; ok {
				d.Stats[combat.CR] += 0.2
			}
			//apply damage
			damage := s.ApplyDamage(d)
			log.Infof("[%v]: Ganyu frost flake bloom dealt %.0f damage", combat.PrintFrames(s.Frame), damage)
			return true
		}
		s.AddAction(initial, fmt.Sprintf("%v-Ganyu-CA-FFA", s.Frame))
		s.AddAction(bloom, fmt.Sprintf("%v-Ganyu-CA-FFB", s.Frame))

		//return animation cd
		return 137
	}
}

func burst(c *combat.Character, log *zap.SugaredLogger) combat.AbilFunc {
	return func(s *combat.Sim) int {
		//check if on cd first
		if _, ok := c.Cooldown["burst-cd"]; ok {
			log.Debugf("\tGanyu burst still on CD; skipping")
			return 0
		}
		//check if sufficient energy
		if c.Energy < c.MaxEnergy {
			log.Debugf("\tGanyu burst - insufficent energy, current: %v", c.Energy)
			return 0
		}
		//snap shot stats at cast time here
		d := c.Snapshot(combat.Cryo)
		d.Abil = "Celestial Shower"
		d.AbilType = combat.ActionTypeBurst
		lvl := c.Profile.TalentLevel[combat.ActionTypeBurst] - 1
		if c.Profile.Constellation >= 3 {
			lvl += 3
			if lvl > 14 {
				lvl = 14
			}
		}
		d.Mult = shower[lvl]
		d.ApplyAura = true
		d.AuraGauge = 1
		d.AuraUnit = "A"

		//apply weapon stats here
		//burst should be instant
		//should add a hook to the unit, triggering damage every 1 sec
		//also add a field effect
		tick := 0
		storm := func(s *combat.Sim) bool {
			if tick > 900 {
				return true
			}
			//check if multiples of 60s; also add an initial delay of 120 frames
			if tick%60 != 0 || tick < 120 {
				tick++
				return false
			}
			//do damage
			damage := s.ApplyDamage(d)
			log.Infof("[%v]: Ganyu burst (tick) dealt %.0f damage", combat.PrintFrames(s.Frame), damage)
			tick++
			return false
		}
		s.AddAction(storm, fmt.Sprintf("%v-Ganyu-Burst", s.Frame))
		//add cooldown to sim
		c.Cooldown["burst-cd"] = 15 * 60
		//use up energy
		c.Energy = 0

		return 122
	}
}

func skill(c *combat.Character, log *zap.SugaredLogger) combat.AbilFunc {
	return func(s *combat.Sim) int {
		//if c2, check if either cd is cooldown
		charge := ""
		_, c2ok := c.Cooldown["skill-cd-2"]
		_, ok := c.Cooldown[charge]

		if c.Profile.Constellation >= 2 {
			if !c2ok {
				charge = "skill-cd-2"
			}
		}
		if !ok {
			charge = "skill-cd"
		}

		if charge == "" {
			log.Debugf("\tGanyu skill still on CD; skipping")
			return 0
		}

		//snap shot stats at cast time here
		d := c.Snapshot(combat.Cryo)
		d.Abil = "Ice Lotus"
		d.AbilType = combat.ActionTypeSkill
		lvl := c.Profile.TalentLevel[combat.ActionTypeSkill] - 1
		if c.Profile.Constellation >= 5 {
			lvl += 3
			if lvl > 14 {
				lvl = 14
			}
		}
		d.Mult = lotus[lvl]
		d.ApplyAura = true
		d.AuraGauge = 1
		d.AuraUnit = "A"

		//we get the orbs right away
		//add delayed orb for travel time
		orbDelay := 0
		s.AddAction(func(s *combat.Sim) bool {
			if orbDelay < 90 { //1.5 second to receive the org
				orbDelay++
				return false
			}
			s.GenerateOrb(2, combat.Cryo, false)
			return true
		}, fmt.Sprintf("%v-Ganyu-Skill-Orb", s.Frame))

		tick := 0
		flower := func(s *combat.Sim) bool {
			if tick < 6*60 {
				tick++
				return false
			}
			//do damage
			damage := s.ApplyDamage(d)
			log.Infof("[%v]: Ganyu ice lotus (tick) dealt %.0f damage", combat.PrintFrames(s.Frame), damage)
			tick++
			return true
		}
		s.AddAction(flower, fmt.Sprintf("%v-Ganyu-Skill", s.Frame))
		//add cooldown to sim
		c.Cooldown[charge] = 10 * 60

		return 30
	}
}
