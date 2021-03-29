package xiangling

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Xiangling", NewChar)
}

type xl struct {
	*combat.CharacterTemplate
	delayedFunc map[int]func()
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	x := xl{}
	t, err := combat.NewTemplateChar(s, p)
	if err != nil {
		return nil, err
	}
	x.CharacterTemplate = t
	x.Energy = 60
	x.MaxEnergy = 60
	x.Profile.WeaponClass = combat.WeaponClassSpear
	x.delayedFunc = make(map[int]func())

	if x.Profile.Constellation >= 6 {
		x.c6()
	}

	return &x, nil
}

func (x *xl) c1() {
	x.S.Target.AddResMod("xiangling-c1", combat.ResistMod{
		Ele:      combat.Pyro,
		Value:    -0.15,
		Duration: 6 * 60,
	})
}

func (x *xl) c6() {
	x.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if x.S.StatusActive("Xiangling C6") {
			ds.Stats[combat.PyroP] += 0.15
		}
		return false
	}, "Xiangling C6", combat.PostSnapshot)
}

func (x *xl) Attack(p map[string]interface{}) int {
	//register action depending on number in chain
	//3 and 4 need to be registered as multi action
	d := x.Snapshot("Normal", combat.ActionTypeAttack, combat.Physical)
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
		x.S.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(d)
			s.Log.Infof("\t Xiangling normal %v (hit %v) dealt %.0f damage", n, t, damage)
		}, fmt.Sprintf("Xiangling-Normal-%v-%v", n, t), 5)
	}
	//if n = 5, add explosion for c2
	if x.Profile.Constellation >= 2 && n == 5 {
		c := d.Clone()
		c.Element = combat.Pyro
		c.ApplyAura = true
		c.AuraBase = combat.WeakAuraBase
		c.AuraUnits = 1
		x.S.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(c)
			s.Log.Infof("\t Xiangling C2 explosion dealt %.0f damage", damage)
		}, "Xiangling-C2-Explosion", 120)
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

func (x *xl) ChargeAttack(p map[string]interface{}) int {
	d := x.Snapshot("Charge Attack", combat.ActionTypeChargedAttack, combat.Physical)
	d.Mult = nc[x.Profile.TalentLevel[combat.ActionTypeAttack]-1]

	//no delay for now? realistically the hits should have delay but not sure if it actually makes a diff
	//since it doesnt apply any elements, only trigger weapon procs
	x.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(d)
		s.Log.Infof("\t Xiangling charge attack dealt %.0f damage", damage)
	}, "Xiangling-Charge-Attack", 1)

	x.NormalResetTimer = 0
	//return animation cd
	return 85
}

func (x *xl) ChargeAttackStam() float64 {
	return 25
}

func (x *xl) Skill(p map[string]interface{}) int {
	//check if on cd first
	if x.CD[combat.SkillCD] > x.S.F {
		x.S.Log.Debugf("\tXiangling skill still on CD; skipping")
		return 0
	}

	d := x.Snapshot("Guoba", combat.ActionTypeSkill, combat.Pyro)
	lvl := x.Profile.TalentLevel[combat.ActionTypeSkill] - 1
	if x.Profile.Constellation >= 5 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	d.Mult = guoba[lvl]
	d.ApplyAura = true //apparently every hit applies
	d.AuraBase = combat.WeakAuraBase
	d.AuraUnits = 1
	delay := 120

	for i := 0; i < 4; i++ {
		x.S.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(d)
			s.Log.Infof("\t Xiangling (Gouba - tick) dealt %.0f damage", damage)
		}, "Xiangling Guoba", delay+i*95)
		x.S.AddEnergyParticles("Xiangling", 1, combat.Pyro, delay+i*95+90+60)
		if x.Profile.Constellation >= 1 {
			x.c1()
		}
	}

	//add cooldown to sim
	x.CD[combat.SkillCD] = x.S.F + 12*60
	x.NormalResetTimer = 0
	//return animation cd
	return 26
}

func (x *xl) Burst(p map[string]interface{}) int {
	//check if on cd first
	if x.CD[combat.BurstCD] > x.S.F {
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
	//initial 3 hits are delayed and snapshotted at execution instead of at cast... no idea why...
	x.delayedFunc[x.S.F+20] = func() {
		d := x.Snapshot("Pyronado", combat.ActionTypeBurst, combat.Pyro)
		d.Mult = pyronado1[lvl]
		d.AuraBase = combat.WeakAuraBase
		d.AuraUnits = 1
		d.ApplyAura = true
		x.S.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(d)
			s.Log.Infof("\t Xiangling Pyronado initial hit 1 dealt %.0f damagee", damage)
		}, "Xiangling-Burst-Hit-1", 0)
	}

	x.delayedFunc[x.S.F+50] = func() {
		d := x.Snapshot("Pyronado", combat.ActionTypeBurst, combat.Pyro)
		d.Mult = pyronado2[lvl]
		d.AuraBase = combat.WeakAuraBase
		d.AuraUnits = 1
		d.ApplyAura = true
		x.S.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(d)
			s.Log.Infof("\t Xiangling Pyronado initial hit 2 dealt %.0f damagee", damage)
		}, "Xiangling-Burst-Hit-2", 0)
	}

	x.delayedFunc[x.S.F+75] = func() {
		d := x.Snapshot("Pyronado", combat.ActionTypeBurst, combat.Pyro)
		d.Mult = pyronado3[lvl]
		d.AuraBase = combat.WeakAuraBase
		d.AuraUnits = 1
		d.ApplyAura = true
		x.S.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(d)
			s.Log.Infof("\t Xiangling Pyronado initial hit 3 dealt %.0f damagee", damage)
		}, "Xiangling-Burst-Hit-3", 0)
	}

	//spin to win; snapshot on cast
	d := x.Snapshot("Pyronado", combat.ActionTypeBurst, combat.Pyro)
	d.Mult = pyronadoSpin[lvl]
	d.ApplyAura = true
	d.AuraBase = combat.WeakAuraBase
	d.AuraUnits = 1

	//ok for now we assume it's 80 (or 70??) frames per cycle, that gives us roughly 10s uptime
	//max is either 10s or 14s
	max := 10 * 60
	if x.Profile.Constellation >= 4 {
		max = 14 * 60
	}
	count := 0

	for delay := 70; delay <= max; delay += 70 {
		count++
		x.S.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(d)
			s.Log.Infof("\t Xiangling (Pyronado - tick #%v) dealt %.0f damage", count, damage)
		}, "Xiangling Pyronado", delay)
	}

	//add an effect starting at frame 70 to end of duration to increase pyro dmg by 15% if c6
	if x.Profile.Constellation >= 6 {
		//wait 70 frames, add effect
		//count to max, remove effect
		x.delayedFunc[x.S.F+70] = func() {
			x.S.Status["Xiangling C6"] = x.S.F + max
		}

	}

	//add cooldown to sim
	x.CD[combat.BurstCD] = x.S.F + 20*60
	//use up energy
	x.Energy = 0

	x.NormalResetTimer = 0
	//return animation cd
	return 140
}

func (x *xl) Tick() {
	x.CharacterTemplate.Tick()
	f, ok := x.delayedFunc[x.S.F]
	if ok {
		f()
		delete(x.delayedFunc, x.S.F)
	}
}
