package xiangling

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("xiangling", NewChar)
}

type char struct {
	*combat.CharacterTemplate
	delayedFunc map[int]func()
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	c := char{}
	t, err := combat.NewTemplateChar(s, p)
	if err != nil {
		return nil, err
	}
	c.CharacterTemplate = t
	c.Energy = 80
	c.MaxEnergy = 80
	c.Weapon.Class = combat.WeaponClassSpear
	c.delayedFunc = make(map[int]func())

	if c.Base.Cons >= 6 {
		c.c6()
	}

	return &c, nil
}

func (c *char) ActionFrames(a combat.ActionType, p int) int {
	switch a {
	case combat.ActionAttack:
		switch c.NormalCounter {
		case 1:
			return 26
		case 2:
			return 41
		case 3:
			return 66
		case 4:
			return 49
		case 5:
			return 17
		}
	case combat.ActionSkill:
		return 26
	case combat.ActionBurst:
		return 140
	}
	c.S.Log.Warnf("%v: unknown action, frames invalid", a)
	return 0
}

func (x *char) c1() {
	x.S.Target.AddResMod("xiangling-c1", combat.ResistMod{
		Ele:      combat.Pyro,
		Value:    -0.15,
		Duration: 6 * 60,
	})
}

func (c *char) c6() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if c.S.StatusActive("Xiangling C6") {
			ds.Stats[combat.PyroP] += 0.15
		}
		return false
	}, "Xiangling C6", combat.PostSnapshot)
}

func (c *char) Attack(p int) int {
	//register action depending on number in chain
	//3 and 4 need to be registered as multi action
	d := c.Snapshot("Normal", combat.ActionAttack, combat.Physical, combat.WeakDurability)
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
		d.Mult = hit[c.TalentLvlAttack()]
		t := i + 1
		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Xiangling normal %v (hit %v) dealt %.0f damage [%v]", n, t, damage, str)
		}, fmt.Sprintf("Xiangling-Normal-%v-%v", n, t), 5)
	}
	//if n = 5, add explosion for c2
	if c.Base.Cons >= 2 && n == 5 {
		d1 := d.Clone()
		d1.Element = combat.Pyro
		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d1)
			s.Log.Infof("\t Xiangling C2 explosion dealt %.0f damage [%v]", damage, str)
		}, "Xiangling-C2-Explosion", 120)
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

func (c *char) ChargeAttack(p int) int {
	d := c.Snapshot("Charge Attack", combat.ActionCharge, combat.Physical, combat.WeakDurability)
	d.Mult = nc[c.TalentLvlAttack()]

	//no delay for now? realistically the hits should have delay but not sure if it actually makes a diff
	//since it doesnt apply any elements, only trigger weapon procs
	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Xiangling charge attack dealt %.0f damage [%v]", damage, str)
	}, "Xiangling-Charge-Attack", 1)

	c.NormalResetTimer = 0
	//return animation cd
	return 85
}

func (c *char) Skill(p int) int {
	//check if on cd first
	if c.CD[combat.SkillCD] > c.S.F {
		c.S.Log.Debugf("\tXiangling skill still on CD; skipping")
		return 0
	}

	d := c.Snapshot("Guoba", combat.ActionSkill, combat.Pyro, combat.WeakDurability)
	d.Mult = guoba[c.TalentLvlSkill()]
	delay := 120

	for i := 0; i < 4; i++ {
		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Xiangling (Gouba - tick) dealt %.0f damage [%v]", damage, str)
			if c.Base.Cons >= 1 {
				c.c1()
			}
		}, "Xiangling Guoba", delay+i*95)
		c.S.AddEnergyParticles("Xiangling", 1, combat.Pyro, delay+i*95+90+60)

	}

	//add cooldown to sim
	c.CD[combat.SkillCD] = c.S.F + 12*60
	c.NormalResetTimer = 0
	//return animation cd
	return c.ActionFrames(combat.ActionSkill, p)
}

func (c *char) Burst(p int) int {
	//check if on cd first
	if c.CD[combat.BurstCD] > c.S.F {
		c.S.Log.Debugf("\tXiangling skill still on CD; skipping")
		return 0
	}
	//check if sufficient energy
	if c.Energy < c.MaxEnergy {
		c.S.Log.Debugf("\tXiangling burst - insufficent energy, current: %v", c.Energy)
		return 0
	}
	lvl := c.TalentLvlBurst()
	//initial 3 hits are delayed and snapshotted at execution instead of at cast... no idea why...
	c.delayedFunc[c.S.F+20] = func() {
		d := c.Snapshot("Pyronado", combat.ActionBurst, combat.Pyro, combat.WeakDurability)
		d.Mult = pyronado1[lvl]
		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Xiangling Pyronado initial hit 1 dealt %.0f damage [%v]", damage, str)
		}, "Xiangling-Burst-Hit-1", 0)
	}

	c.delayedFunc[c.S.F+50] = func() {
		d := c.Snapshot("Pyronado", combat.ActionBurst, combat.Pyro, combat.WeakDurability)
		d.Mult = pyronado2[lvl]
		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Xiangling Pyronado initial hit 2 dealt %.0f damage [%v]", damage, str)
		}, "Xiangling-Burst-Hit-2", 0)
	}

	c.delayedFunc[c.S.F+75] = func() {
		d := c.Snapshot("Pyronado", combat.ActionBurst, combat.Pyro, combat.WeakDurability)
		d.Mult = pyronado3[lvl]
		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Xiangling Pyronado initial hit 3 dealt %.0f damage [%v]", damage, str)
		}, "Xiangling-Burst-Hit-3", 0)
	}

	//spin to win; snapshot on cast
	d := c.Snapshot("Pyronado", combat.ActionBurst, combat.Pyro, combat.WeakDurability)
	d.Mult = pyronadoSpin[lvl]

	//ok for now we assume it's 80 (or 70??) frames per cycle, that gives us roughly 10s uptime
	//max is either 10s or 14s
	max := 10 * 60
	if c.Base.Cons >= 4 {
		max = 14 * 60
	}
	count := 0

	for delay := 70; delay <= max; delay += 70 {
		count++
		i := count
		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Xiangling (Pyronado - tick #%v) dealt %.0f damage [%v]", i, damage, str)
		}, "Xiangling Pyronado", delay)
	}

	//add an effect starting at frame 70 to end of duration to increase pyro dmg by 15% if c6
	if c.Base.Cons >= 6 {
		//wait 70 frames, add effect
		//count to max, remove effect
		c.delayedFunc[c.S.F+70] = func() {
			c.S.Status["Xiangling C6"] = c.S.F + max
		}

	}

	//add cooldown to sim
	c.CD[combat.BurstCD] = c.S.F + 20*60
	//use up energy
	c.Energy = 0

	c.NormalResetTimer = 0
	//return animation cd
	return c.ActionFrames(combat.ActionBurst, p)
}

func (c *char) Tick() {
	c.CharacterTemplate.Tick()
	f, ok := c.delayedFunc[c.S.F]
	if ok {
		f()
		delete(c.delayedFunc, c.S.F)
	}
}
