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
		frames := 23
		n := 1
		switch c.AttackCounter {
		case 1:
			hits = n2
			frames = 38
			n = 2
		case 2:
			hits = n3
			frames = 56 + 3
			n = 3
		case 3:
			hits = n4
			frames = 57
			n = 4
		case 4:
			hits = n5
			frames = 60
			n = 5
			reset = true
		default:
			hits = n1
		}
		c.AttackCounter++

		for i, hit := range hits {
			d.Mult = hit[c.Profile.TalentLevel[combat.ActionTypeAttack]-1]
			s.AddAction(func(s *combat.Sim) bool {
				//no delay for now? realistically the hits should have delay but not sure if it actually makes a diff
				//since it doesnt apply any elements, only trigger weapon procs
				damage := s.ApplyDamage(d)
				log.Infof("[%v]: Xiangling normal %v (hit %v) dealt %.0f damage", combat.PrintFrames(s.Frame), n, i+1, damage)
				return true
			}, fmt.Sprintf("%v-Xiangling-Normal-%v-%v", s.Frame, n, i))
		}

		if reset {
			c.AttackCounter = 0
		}
		//return animation cd
		//this also depends on which hit in the chain this is
		return frames
	}
}

func charge(c *combat.Character, log *zap.SugaredLogger) combat.AbilFunc {
	return func(s *combat.Sim) int {

		//return animation cd
		return 0
	}
}

func burst(c *combat.Character, log *zap.SugaredLogger) combat.AbilFunc {
	return func(s *combat.Sim) int {

		//return animation cd
		return 0
	}
}

func skill(c *combat.Character, log *zap.SugaredLogger) combat.AbilFunc {
	return func(s *combat.Sim) int {

		//return animation cd
		return 0
	}
}
