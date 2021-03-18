package ganyu

import (
	"fmt"

	"github.com/srliao/gisim/internal/pkg/combat"
	"go.uber.org/zap"
)

func init() {
	combat.RegisterCharFunc("Ganyu", New)
}

func New(s *combat.Sim) *combat.Character {
	c := &combat.Character{}
	c.ChargeAttack = charge(c, s.Log)
	c.Burst = burst(c, s.Log)
	c.Skill = skill(c, s.Log)
	c.MaxEnergy = 60
	c.Energy = 60

	return c
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
		//snap shot stats at cast time here
		d := c.Snapshot(combat.Cryo)
		d.Abil = "Celestial Shower"
		d.AbilType = combat.ActionTypeBurst
		d.Mult = shower[c.Profile.TalentLevel[combat.ActionTypeBurst]-1]
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

		return 122
	}
}

func skill(c *combat.Character, log *zap.SugaredLogger) combat.AbilFunc {
	return func(s *combat.Sim) int {
		//snap shot stats at cast time here
		d := c.Snapshot(combat.Cryo)
		d.Abil = "Ice Lotus"
		d.AbilType = combat.ActionTypeSkill
		d.Mult = lotus[c.Profile.TalentLevel[combat.ActionTypeSkill]-1]
		d.ApplyAura = true
		d.AuraGauge = 1
		d.AuraUnit = "A"

		tick := 0
		flower := func(s *combat.Sim) bool {
			if tick < 6*60 {
				return false
			}
			//do damage
			damage := s.ApplyDamage(d)
			zap.S().Infof("[%v]: Ganyu ice lotus (tick) dealt %.0f damage", combat.PrintFrames(s.Frame), damage)
			tick++
			return false
		}
		s.AddAction(flower, fmt.Sprintf("%v-Ganyu-Skill", s.Frame))
		//add cooldown to sim
		c.Cooldown["cd-skill"] = 15 * 60

		return 30
	}
}
