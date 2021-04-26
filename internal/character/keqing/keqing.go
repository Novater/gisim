package keqing

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("keqing", NewChar)
}

type char struct {
	*combat.CharacterTemplate
	eStarted    bool
	eStartFrame int
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	c := char{}
	t, err := combat.NewTemplateChar(s, p)
	if err != nil {
		return nil, err
	}
	c.CharacterTemplate = t
	c.Energy = 40
	c.MaxEnergy = 40
	c.Weapon.Class = combat.WeaponClassSword

	c.skillChargeHook()
	c.a2()

	return &c, nil
}

func (c *char) ActionFrames(a combat.ActionType, p int) int {
	switch a {
	case combat.ActionAttack:
		switch c.NormalCounter {
		//TODO: need to add atkspd mod
		case 1:
			return 11
		case 2:
			return 33
		case 3:
			return 60
		case 4:
			return 97
		case 5:
			return 133
		}
	case combat.ActionSkill:
		return 1
	case combat.ActionBurst:
		return 1
	}
	c.S.Log.Warnf("%v: unknown action, frames invalid", a)
	return 0
}

func (c *char) a2() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor != c.Base.Name {
			return false
		}
		if ds.AbilType != combat.ActionAttack && ds.AbilType != combat.ActionCharge {
			return false
		}
		if !c.S.StatusActive("Keqing-A2-Infuse") {
			return false
		}
		ds.Element = combat.Electro
		return false
	}, "keqing-a2-infuse", combat.PostSnapshot)
}

func (c *char) Attack(p int) int {
	delay := [][]int{
		{11}, {33}, {60}, {87, 97}, {133},
	}

	for i := 0; i < len(attack[c.NormalCounter]); i++ {
		d := c.Snapshot("Normal", combat.ActionAttack, combat.Physical, combat.WeakDurability)
		d.Mult = attack[c.NormalCounter][i][c.TalentLvlAttack()]
		t := i + 1
		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t %v normal %v (hit %v) dealt %.0f damage [%v]", c.Base.Name, c.NormalCounter, t, damage, str)
		}, fmt.Sprintf("%v-Normal-%v-%v", c.Base.Name, c.NormalCounter, t), delay[c.NormalCounter][i])
	}

	c.NormalCounter++
	f := c.ActionFrames(combat.ActionAttack, p)

	//add a 70 frame attackcounter reset
	c.NormalResetTimer = 70

	if c.NormalCounter >= 5 {
		c.NormalResetTimer = 0
		c.NormalCounter = 0
	}

	return f
}

//p = 0 just cast w/e, p = 1 first cast specifically, p = 2 second cast
//BUT if p = 2 and first not cast, it will default to first
//realistically this is only useful in a combo where you care about
//frames checking (i.e. stam consumption); otherwise just use p = 0
//and let the sim handle it
func (c *char) Skill(p int) int {
	if c.CD[combat.SkillCD] > c.S.F {
		c.S.Log.Debugf("\t Keqing skill still on CD; skipping")
		return 0
	}
	//since p is only used for stam checking really we can just ignore it
	if c.eStarted {
		return c.skillNext()
	}
	return c.skillFirst()
}

func (c *char) skillFirst() int {
	d := c.Snapshot("Stellar Restoration", combat.ActionSkill, combat.Electro, combat.WeakDurability)
	d.Mult = skill[c.TalentLvlSkill()]
	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Keqing skill hit dealt %.0f damage [%v]", damage, str)

	}, "Keqing-Skill-1", 1) //TODO: frames
	c.eStarted = true
	c.eStartFrame = c.S.F
	//place on cd after certain frames if started is still true
	c.AddTask(func() {
		if c.eStarted {
			c.eStarted = false
			c.CD[combat.SkillCD] = c.eStartFrame + 100 //TODO: cooldown if not triggered
		}
	}, "keqing-skill-cd", c.S.F+500) //TODO: delay

	return c.ActionFrames(combat.ActionSkill, 1)
}

func (c *char) skillNext() int {
	d := c.Snapshot("Stellar Restoration (Slashing)", combat.ActionSkill, combat.Electro, combat.WeakDurability)
	d.Mult = skillPress[c.TalentLvlSkill()]
	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Keqing skill teleport dealt %.0f damage [%v]", damage, str)

	}, "Keqing-Skill-1", 1) //TODO: frames
	//add electro infusion
	c.S.Status["Keqing-A2-Infuse"] = c.S.F + 300
	//place on cooldown
	c.eStarted = false
	c.CD[combat.SkillCD] = c.eStartFrame + 100
	return c.ActionFrames(combat.ActionSkill, 2)
}

func (c *char) skillChargeHook() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if !c.eStarted {
			return false
		}
		if ds.Actor != c.Base.Name {
			return false
		}
		if ds.AbilType != combat.ActionCharge {
			return false
		}
		//ok burst the bubble

		//2 hits
		for i := 0; i < 2; i++ {
			d := c.Snapshot("Stellar Restoration (Thunderclap Slash)", combat.ActionSkill, combat.Electro, combat.WeakDurability)
			d.Mult = skillCA[c.TalentLvlSkill()]
			c.S.AddTask(func(s *combat.Sim) {
				damage, str := s.ApplyDamage(d)
				s.Log.Infof("\t Keqing skill charge attack dealt %.0f damage [%v]", damage, str)
			}, "Keqing-Skill-1", 1) //TODO: frames
		}

		//place on cooldown
		c.eStarted = false
		c.CD[combat.SkillCD] = c.eStartFrame + 100
		return false
	}, "keqing-skill-ca", combat.PostDamageHook)
}

func (c *char) Burst(p int) int {
	//a4 increase crit + ER
	val := make(map[combat.StatType]float64)
	val[combat.CR] = 0.15
	val[combat.ER] = 0.15
	c.AddMod("a4", val)
	//add hook to remove this
	c.AddTask(func() {
		c.RemoveMod("a4")
	}, "keqing-a4", c.S.F+480)

	//initial
	initial := c.Snapshot("Starward Sword", combat.ActionBurst, combat.Electro, combat.WeakDurability)
	initial.Mult = burstInitial[c.TalentLvlBurst()]
	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(initial)
		s.Log.Infof("\t Keqing burst initial dealt %.0f damage [%v]", damage, str)
	}, "Keqing-Burst-Initial", 1) //TODO: frames

	//8 hits
	dot := c.Snapshot("Starward Sword", combat.ActionBurst, combat.Electro, combat.WeakDurability) //TODO: electro application
	dot.Mult = burstDot[c.TalentLvlBurst()]
	for i := 0; i < 8; i++ {
		t := i + 1
		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(dot)
			s.Log.Infof("\t Keqing burst dot %v dealt %.0f damage [%v]", t, damage, str)
		}, fmt.Sprintf("Keqing-Burst-Dot-%v", i), 1) //TODO: frames
	}

	//final
	final := c.Snapshot("Starward Sword", combat.ActionBurst, combat.Electro, combat.WeakDurability)
	final.Mult = burstFinal[c.TalentLvlBurst()]
	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(initial)
		s.Log.Infof("\t Keqing burst final hit dealt %.0f damage [%v]", damage, str)
	}, "Keqing-Burst-Initial", 1) //TODO: frames

	c.Energy = 0
	c.CD[combat.BurstCD] = c.S.F + 720 //12s
	return c.ActionFrames(combat.ActionBurst, p)
}
