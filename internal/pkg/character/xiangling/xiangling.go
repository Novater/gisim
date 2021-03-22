package xiangling

import (
	"fmt"

	"github.com/srliao/gisim/internal/pkg/character/common"
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Xiangling", NewChar)
}

type xl struct {
	*common.TemplateChar
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	x := xl{}
	t, err := common.New(s, p)
	if err != nil {
		return nil, err
	}
	x.TemplateChar = t
	x.Energy = 60
	x.MaxEnergy = 60

	if x.Profile.Constellation >= 1 {
		s.Log.Debugf("\tactivating Xiangling C1")

		s.AddHook(func(snap *combat.Snapshot) bool {
			//check if c1 debuff is on, if so, reduce resist by -0.15
			if _, ok := s.Target.Status["xiangling-c1"]; ok {
				s.Log.Debugf("\t[%v]: applying Xiangling C1 pyro debuff", s.Frame())
				snap.ResMod[combat.Pyro] -= 0.15
			}
			return false
		}, "xiangling-c1", combat.PreDamageHook)

		s.AddHook(func(snap *combat.Snapshot) bool {
			if snap.CharName == "Xiangling" && snap.Abil == "Guoba" {
				// add c1 debuff to target
				s.Target.Status["xiangling-c1"] = 6 * 60
			}
			return false
		}, "xiangling-c1", combat.PostDamageHook)
	}

	return &x, nil
}

func (x *xl) FillerFrames() int {
	frames := 26
	//hit one starts at 1955 end 2097
	//1480 to 1677, 1853, 2045
	switch x.NormalCounter {
	case 1:
		frames = 41
	case 2:
		frames = 66
	case 3:
		frames = 49
	case 4:
		frames = 17
	}
	return frames
}

func (x *xl) Attack() int {
	//register action depending on number in chain
	//3 and 4 need to be registered as multi action
	d := x.Snapshot(combat.Physical)
	d.Abil = "Normal"
	d.AbilType = combat.ActionTypeAttack
	//figure out which hit it is
	var hits [][]float64
	reset := false
	frames := 26
	n := 1
	//hit one starts at 1955 end 2097
	//1480 to 1677, 1853, 2045
	switch x.NormalCounter {
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
	x.NormalCounter++
	//apply attack speed
	frames = int(float64(frames) / (1 + x.Stats[combat.AtkSpd]))
	for i, hit := range hits {
		d.Mult = hit[x.Profile.TalentLevel[combat.ActionTypeAttack]-1]
		t := i + 1
		x.S.AddAction(func(s *combat.Sim) bool {
			//no delay for now? realistically the hits should have delay but not sure if it actually makes a diff
			//since it doesnt apply any elements, only trigger weapon procs
			c := d.Clone()
			damage := s.ApplyDamage(c)
			s.Log.Infof("[%v]: Xiangling normal %v (hit %v) dealt %.0f damage", s.Frame(), n, t, damage)
			return true
		}, fmt.Sprintf("%v-Xiangling-Normal-%v-%v", x.S.Frame(), n, i))
	}
	//if n = 5, add explosion for c2
	if x.Profile.Constellation >= 2 && n == 5 {
		tick := 0
		x.S.AddAction(func(s *combat.Sim) bool {
			tick++
			if tick < 2*60 {
				return false
			}
			c := d.Clone()
			c.Element = combat.Pyro
			damage := s.ApplyDamage(c)
			s.Log.Infof("[%v]: Xiangling C2 explosion dealt %.0f damage", s.Frame(), damage)
			return true
		}, fmt.Sprintf("%v-Xiangling-C2-Explosion", x.S.Frame()))
	}
	//add a 75 frame attackcounter reset
	x.NormalResetTimer = 70

	if reset {
		x.NormalResetTimer = 0
	}
	//return animation cd
	//this also depends on which hit in the chain this is
	return frames
}

func (x *xl) ChargeAttack() int {
	d := x.Snapshot(combat.Physical)
	d.Abil = "Charge Attack"
	d.AbilType = combat.ActionTypeChargedAttack
	d.Mult = nc[x.Profile.TalentLevel[combat.ActionTypeAttack]-1]

	x.S.AddAction(func(s *combat.Sim) bool {
		//no delay for now? realistically the hits should have delay but not sure if it actually makes a diff
		//since it doesnt apply any elements, only trigger weapon procs
		damage := s.ApplyDamage(d)
		s.Log.Infof("[%v]: Xiangling charge attack dealt %.0f damage", s.Frame(), damage)
		return true
	}, fmt.Sprintf("%v-Xiangling-Charge-Attack", x.S.Frame()))
	x.NormalResetTimer = 0
	//return animation cd
	return 85
}

func (x *xl) ChargeAttackStam() float64 {
	return 25
}

func (x *xl) Skill() int {
	//check if on cd first
	if _, ok := x.CD["skill-cd"]; ok {
		x.S.Log.Debugf("\tXiangling skill still on CD; skipping")
		return 0
	}

	d := x.Snapshot(combat.Pyro)
	d.Abil = "Guoba"
	d.AbilType = combat.ActionTypeSkill
	lvl := x.Profile.TalentLevel[combat.ActionTypeSkill] - 1
	if x.Profile.Constellation >= 5 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	d.Mult = guoba[lvl]
	d.ApplyAura = true
	d.AuraGauge = 1
	d.AuraDecayRate = "A"

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

			next += 95
			count++
			damage := s.ApplyDamage(c)
			s.Log.Infof("[%v]: Xiangling (Gouba - tick) dealt %.0f damage", s.Frame(), damage)
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
			}, fmt.Sprintf("%v-Xiangling-Skill-Orb", s.Frame()))
		}
		if count == 4 {
			s.Log.Infof("[%v]: Xiangling (Gouba) expired", s.Frame())
			return true
		}
		return false
	}
	x.S.AddAction(g, fmt.Sprintf("%v-Xiangling-Skill", x.S.Frame()))
	//add cooldown to sim
	x.CD["skill-cd"] = 12 * 60
	x.NormalResetTimer = 0
	//return animation cd
	return 26
}

func (x *xl) Burst() int {
	//check if on cd first
	if _, ok := x.CD["burst-cd"]; ok {
		x.S.Log.Debugf("\tXiangling skill still on CD; skipping")
		return 0
	}
	//check if sufficient energy
	if x.Energy < x.MaxEnergy {
		x.S.Log.Debugf("\tXiangling burst - insufficent energy, current: %v", x.Energy)
		return 0
	}
	lvl := x.Profile.TalentLevel[combat.ActionTypeBurst] - 1
	if x.Profile.Constellation >= 5 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	//first hit
	h1d := 0
	x.S.AddAction(func(s *combat.Sim) bool {
		h1d++
		if h1d < 20 {
			return false
		}
		d := x.Snapshot(combat.Pyro)
		d.Abil = "Pyronado"
		d.AbilType = combat.ActionTypeBurst
		d.Mult = pyronado1[lvl]
		d.ApplyAura = true
		d.AuraGauge = 1
		d.AuraDecayRate = "A"
		damage := s.ApplyDamage(d)
		x.S.Log.Infof("[%v]: Xiangling Pyronado initial hit 1 dealt %.0f damage", s.Frame(), damage)
		return true
	}, fmt.Sprintf("%v-Xiangling-Burst-Hit-1", x.S.Frame()))
	h2d := 0
	x.S.AddAction(func(s *combat.Sim) bool {
		h2d++
		if h2d < 50 {
			return false
		}
		d := x.Snapshot(combat.Pyro)
		d.Abil = "Pyronado"
		d.AbilType = combat.ActionTypeBurst
		d.Mult = pyronado2[lvl]
		d.ApplyAura = true
		d.AuraGauge = 1
		d.AuraDecayRate = "A"
		damage := s.ApplyDamage(d)
		s.Log.Infof("[%v]: Xiangling Pyronado initial hit 2 dealt %.0f damage", s.Frame(), damage)
		return true
	}, fmt.Sprintf("%v-Xiangling-Burst-Hit-2", x.S.Frame()))
	h3d := 0
	x.S.AddAction(func(s *combat.Sim) bool {
		h3d++
		if h3d < 75 {
			return false
		}
		d := x.Snapshot(combat.Pyro)
		d.Abil = "Pyronado"
		d.AbilType = combat.ActionTypeBurst
		d.Mult = pyronado3[lvl]
		d.ApplyAura = true
		d.AuraGauge = 1
		d.AuraDecayRate = "A"
		damage := s.ApplyDamage(d)
		s.Log.Infof("[%v]: Xiangling Pyronado initial hit 3 dealt %.0f damage", s.Frame(), damage)
		return true
	}, fmt.Sprintf("%v-Xiangling-Burst-Hit-3", x.S.Frame()))
	//ok for now we assume it's 80 frames per cycle, that gives us roughly 10s uptime
	tick := 0
	next := 70
	//max is either 10s or 14s
	max := 10 * 60
	if x.Profile.Constellation >= 4 {
		max = 14 * 60
	}
	count := 0
	//pyronado snaps at cast time
	pd := x.Snapshot(combat.Pyro)
	pd.Abil = "Pyronado"
	pd.AbilType = combat.ActionTypeBurst
	pd.Mult = pyronadoSpin[lvl]
	pd.ApplyAura = true
	pd.AuraGauge = 1
	pd.AuraDecayRate = "A"
	x.S.AddAction(func(s *combat.Sim) bool {
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
			s.Log.Infof("[%v]: Xiangling (Pyronado - tick #%v) dealt %.0f damage", s.Frame(), count, damage)
		}
		return false
	}, fmt.Sprintf("%v-Xiangling-Burst-Spin", x.S.Frame()))
	//add an effect starting at frame 70 to end of duration to increase pyro dmg by 15% if c6
	if x.Profile.Constellation >= 6 {
		//wait 70 frames, add effect
		//count to max, remove effect
		c6tick := 0
		x.S.AddAction(func(s *combat.Sim) bool {
			c6tick++
			if c6tick < 70 {
				return false
			}
			if c6tick == 70 {
				m := make(map[combat.StatType]float64)
				m[combat.PyroP] = 0.15
				s.AddFieldEffect("Xiangling C6", m)
				return false
			}
			if c6tick >= max {
				s.RemoveFieldEffect("Xiangling C6")
				return true
			}
			return false
		}, fmt.Sprintf("%v-Xiangling-Burst-C6", x.S.Frame()))

	}

	//add cooldown to sim
	x.CD["burst-cd"] = 20 * 60
	//use up energy
	x.Energy = 0

	x.NormalResetTimer = 0
	//return animation cd
	return 140
}
