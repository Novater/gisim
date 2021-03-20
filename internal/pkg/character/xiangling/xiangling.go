package xiangling

import (
	"fmt"

	"github.com/srliao/gisim/internal/pkg/combat"
	"go.uber.org/zap"
)

func init() {
	combat.RegisterCharFunc("Xiangling", New)
}

func New(s *combat.Sim, c *combat.Character) {
	c.Attack = normal(c, s.Log)
	c.ChargeAttack = charge(c, s.Log)
	c.Burst = burst(c, s.Log)
	c.Skill = skill(c, s.Log)
	c.MaxEnergy = 80
	c.Energy = 80

	if c.Profile.Constellation >= 1 {
		s.Log.Debugf("\tactivating Xiangling C1")

		s.AddEffect(func(snap *combat.Snapshot) bool {
			//check if c1 debuff is on, if so, reduce resist by -0.15
			if _, ok := s.Target.Status["xiangling-c1"]; ok {
				s.Log.Debugf("\t[%v]: applying Xiangling C1 pyro debuff", combat.PrintFrames(s.Frame))
				snap.ResMod[combat.Pyro] -= 0.15
			}
			return false
		}, "xiangling-c1", combat.PreDamageHook)

		s.AddEffect(func(snap *combat.Snapshot) bool {
			if snap.CharName == "Xiangling" && snap.Abil == "Guoba" {
				// add c1 debuff to target
				s.Target.Status["xiangling-c1"] = 6 * 60
			}
			return false
		}, "xiangling-c1", combat.PostDamageHook)
	}

}

func normal(c *combat.Character, log *zap.SugaredLogger) combat.AbilFunc {
	return func(s *combat.Sim) int {
		//register action depending on number in chain
		//3 and 4 need to be registered as multi action
		d := c.Snapshot(combat.Physical)
		d.Abil = "Normal"
		d.AbilType = combat.ActionTypeAttack
		//figure out which hit it is
		var hits [][]float64
		reset := false
		frames := 26
		n := 1
		//hit one starts at 1955 end 2097
		//1480 to 1677, 1853, 2045
		switch c.NormalCounter {
		case 1:
			hits = n2
			frames = 41
			n = 2
		case 2:
			hits = n3
			frames = 66
			n = 3
		case 3:
			hits = n4
			frames = 49
			n = 4
		case 4:
			hits = n5
			frames = 17
			n = 5
			reset = true
		default:
			hits = n1
		}
		c.NormalCounter++
		//apply attack speed
		frames = int(float64(frames) / (1 + c.Stats[combat.AtkSpd]))
		for i, hit := range hits {
			d.Mult = hit[c.Profile.TalentLevel[combat.ActionTypeAttack]-1]
			x := i + 1
			s.AddAction(func(s *combat.Sim) bool {
				//no delay for now? realistically the hits should have delay but not sure if it actually makes a diff
				//since it doesnt apply any elements, only trigger weapon procs
				c := d.Clone()
				damage := s.ApplyDamage(c)
				log.Infof("[%v]: Xiangling normal %v (hit %v) dealt %.0f damage", combat.PrintFrames(s.Frame), n, x, damage)
				return true
			}, fmt.Sprintf("%v-Xiangling-Normal-%v-%v", s.Frame, n, i))
		}
		//if n = 5, add explosion for c2
		if c.Profile.Constellation >= 2 && n == 5 {
			tick := 0
			s.AddAction(func(s *combat.Sim) bool {
				tick++
				if tick < 2*60 {
					return false
				}
				c := d.Clone()
				c.Element = combat.Pyro
				damage := s.ApplyDamage(c)
				log.Infof("[%v]: Xiangling C2 explosion dealt %.0f damage", combat.PrintFrames(s.Frame), damage)
				return true
			}, fmt.Sprintf("%v-Xiangling-C2-Explosion", s.Frame))
		}
		//add a 75 frame attackcounter reset
		c.NormalResetTimer = 70

		if reset {
			c.NormalResetTimer = 0
		}
		//return animation cd
		//this also depends on which hit in the chain this is
		return frames
	}
}

func charge(c *combat.Character, log *zap.SugaredLogger) combat.AbilFunc {
	return func(s *combat.Sim) int {
		d := c.Snapshot(combat.Physical)
		d.Abil = "Charge Attack"
		d.AbilType = combat.ActionTypeChargedAttack

		s.AddAction(func(s *combat.Sim) bool {
			//no delay for now? realistically the hits should have delay but not sure if it actually makes a diff
			//since it doesnt apply any elements, only trigger weapon procs
			damage := s.ApplyDamage(d)
			log.Infof("[%v]: Xiangling charge attack dealt %.0f damage", combat.PrintFrames(s.Frame), damage)
			return true
		}, fmt.Sprintf("%v-Xiangling-Charge-Attack", s.Frame))
		c.NormalResetTimer = 0
		//return animation cd
		return 85
	}
}

func plunge(c *combat.Character, log *zap.SugaredLogger) combat.AbilFunc {
	return func(s *combat.Sim) int {
		log.Warnf("[%v]: Xiangling plunge attack not implemented", combat.PrintFrames(s.Frame))
		return 0
	}
}

//ult starts 720
//start at 790 first hit

func burst(c *combat.Character, log *zap.SugaredLogger) combat.AbilFunc {
	return func(s *combat.Sim) int {
		//check if on cd first
		if _, ok := c.Cooldown["burst-cd"]; ok {
			log.Debugf("\tXiangling skill still on CD; skipping")
			return 0
		}
		//check if sufficient energy
		if c.Energy < c.MaxEnergy {
			log.Debugf("\tXiangling burst - insufficent energy, current: %v", c.Energy)
			return 0
		}

		//starts 1390
		//first hit 1410
		//second 1440
		//third 1465
		//first rotation 1530
		//rotation 1610 / 1600
		//rotation 1690 / 1680
		//rotation 1770 / 1750
		//1850 / 1840
		//1930
		//2000??
		//2080??
		//2150
		//2230
		lvl := c.Profile.TalentLevel[combat.ActionTypeBurst] - 1
		if c.Profile.Constellation >= 5 {
			lvl += 3
			if lvl > 14 {
				lvl = 14
			}
		}
		//first hit
		h1d := 0
		s.AddAction(func(s *combat.Sim) bool {
			h1d++
			if h1d < 20 {
				return false
			}
			d := c.Snapshot(combat.Pyro)
			d.Abil = "Pyronado"
			d.AbilType = combat.ActionTypeBurst
			d.Mult = pyronado1[lvl]
			damage := s.ApplyDamage(d)
			log.Infof("[%v]: Xiangling Pyronado initial hit 1 dealt %.0f damage", combat.PrintFrames(s.Frame), damage)
			return true
		}, fmt.Sprintf("%v-Xiangling-Burst-Hit-1", s.Frame))
		h2d := 0
		s.AddAction(func(s *combat.Sim) bool {
			h2d++
			if h2d < 50 {
				return false
			}
			d := c.Snapshot(combat.Pyro)
			d.Abil = "Pyronado"
			d.AbilType = combat.ActionTypeBurst
			d.Mult = pyronado2[lvl]
			damage := s.ApplyDamage(d)
			log.Infof("[%v]: Xiangling Pyronado initial hit 2 dealt %.0f damage", combat.PrintFrames(s.Frame), damage)
			return true
		}, fmt.Sprintf("%v-Xiangling-Burst-Hit-2", s.Frame))
		h3d := 0
		s.AddAction(func(s *combat.Sim) bool {
			h3d++
			if h3d < 75 {
				return false
			}
			d := c.Snapshot(combat.Pyro)
			d.Abil = "Pyronado"
			d.AbilType = combat.ActionTypeBurst
			d.Mult = pyronado3[lvl]
			damage := s.ApplyDamage(d)
			log.Infof("[%v]: Xiangling Pyronado initial hit 3 dealt %.0f damage", combat.PrintFrames(s.Frame), damage)
			return true
		}, fmt.Sprintf("%v-Xiangling-Burst-Hit-3", s.Frame))
		//ok for now we assume it's 80 frames per cycle, that gives us roughly 10s uptime
		tick := 0
		next := 70
		//max is either 10s or 14s
		max := 10 * 60
		if c.Profile.Constellation >= 4 {
			max = 14 * 60
		}
		count := 0
		//pyronado snaps at cast time
		pd := c.Snapshot(combat.Pyro)
		pd.Abil = "Pyronado"
		pd.AbilType = combat.ActionTypeBurst
		pd.Mult = pyronadoSpin[lvl]
		s.AddAction(func(s *combat.Sim) bool {
			tick++
			if tick < next {
				return false
			}
			//exit if expired
			if tick >= max {
				return true
			}
			if tick == next {
				count++
				//make a copy of the snapshot
				next += 70
				damage := s.ApplyDamage(pd)
				log.Infof("[%v]: Xiangling (Pyronado - tick #%v) dealt %.0f damage", combat.PrintFrames(s.Frame), count, damage)
			}
			return false
		}, fmt.Sprintf("%v-Xiangling-Burst-Spin", s.Frame))
		//add an effect starting at frame 70 to end of duration to increase pyro dmg by 15% if c6
		if c.Profile.Constellation >= 6 {
			//wait 70 frames, add effect
			//count to max, remove effect
			c6tick := 0
			s.AddAction(func(s *combat.Sim) bool {
				c6tick++
				if c6tick < 70 {
					return false
				}
				if c6tick == 70 {
					for _, cs := range s.Characters {
						cs.Mods["Xiangling C6"] = make(map[combat.StatType]float64)
						cs.Mods["Xiangling C6"][combat.PyroP] = 0.15
					}
					return false
				}
				if c6tick >= max {
					for _, cs := range s.Characters {
						delete(cs.Mods, "Xiangling C6")
					}
					return true
				}
				return false
			}, fmt.Sprintf("%v-Xiangling-Burst-C6", s.Frame))

		}

		//add cooldown to sim
		c.Cooldown["burst-cd"] = 20 * 60
		//use up energy
		c.Energy = 0

		c.NormalResetTimer = 0
		//return animation cd
		return 140
	}
}

func skill(c *combat.Character, log *zap.SugaredLogger) combat.AbilFunc {
	return func(s *combat.Sim) int {
		//check if on cd first
		if _, ok := c.Cooldown["skill-cd"]; ok {
			log.Debugf("\tXiangling skill still on CD; skipping")
			return 0
		}

		d := c.Snapshot(combat.Pyro)
		d.Abil = "Guoba"
		d.AbilType = combat.ActionTypeSkill
		lvl := c.Profile.TalentLevel[combat.ActionTypeSkill] - 1
		if c.Profile.Constellation >= 5 {
			lvl += 3
			if lvl > 14 {
				lvl = 14
			}
		}
		d.Mult = guoba[lvl]
		d.ApplyAura = true
		d.AuraGauge = 1
		d.AuraUnit = "A"

		//we get orb after a delay each tick, tick 4 times
		tick := 0
		next := 120
		count := 0
		g := func(s *combat.Sim) bool {
			//cast 1630
			//first app 1750, next app after 90 frames @ 1840, total 4 casts
			//120 frames delay initially
			tick++
			if tick < next {
				return false
			}
			if tick == next {
				//make a copy of the snapshot
				c := combat.Snapshot{}
				c = d
				c.ResMod = make(map[combat.EleType]float64)

				next += 90
				count++
				damage := s.ApplyDamage(c)
				log.Infof("[%v]: Xiangling (Gouba - tick) dealt %.0f damage", combat.PrintFrames(s.Frame), damage)
				//generate orbs
				//add delayed orb for travel time
				orbDelay := 0
				s.AddAction(func(s *combat.Sim) bool {
					if orbDelay < 90+60 { //it takes 90 frames to generate orb, add another 60 frames to get it
						orbDelay++
						return false
					}
					s.GenerateOrb(1, combat.Pyro, false)
					return true
				}, fmt.Sprintf("%v-Xiangling-Skill-Orb", s.Frame))
			}
			if count == 4 {
				log.Infof("[%v]: Xiangling (Gouba) expired", combat.PrintFrames(s.Frame))
				return true
			}
			return false
		}
		s.AddAction(g, fmt.Sprintf("%v-Xiangling-Skill", s.Frame))
		//add cooldown to sim
		c.Cooldown["skill-cd"] = 12 * 60
		c.NormalResetTimer = 0
		//return animation cd
		return 40
	}
}
