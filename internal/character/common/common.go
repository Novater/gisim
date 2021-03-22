package common

import (
	"github.com/srliao/gisim/pkg/combat"
)

type TemplateChar struct {
	S     *combat.Sim
	CD    map[string]int
	Stats map[combat.StatType]float64
	Mods  map[string]map[combat.StatType]float64
	//Profile info
	Profile   combat.CharacterProfile
	Energy    float64
	MaxEnergy float64
	//counters
	NormalCounter    int
	NormalResetTimer int
	NRTChanged       bool
}

type ICDType string

const (
	NormalICD  string = "aura-icd-normal"   //melee, bow users can't infuse (yet)
	ChargedICD string = "aura-icd-charged"  //only applicable to catalyst
	AimModeICD string = "aura-icd-aim-mode" //bow users
	PlungeICD  string = "aura-icd-plunge"   //xiao
	SkillICD   string = "aura-icd-skill"
	BurstICD   string = "aura-icd-burst"
	SkillCD    string = "skill-cd"
	BurstCD    string = "burst-cd"
)

func New(s *combat.Sim, p combat.CharacterProfile) (*TemplateChar, error) {
	c := TemplateChar{}
	c.S = s
	c.CD = make(map[string]int)
	c.Stats = make(map[combat.StatType]float64)
	c.Mods = make(map[string]map[combat.StatType]float64)
	c.Profile = p

	for _, a := range p.Artifacts {
		c.Stats[a.MainStat.Type] += a.MainStat.Value
		for _, sub := range a.Substat {
			c.Stats[sub.Type] += sub.Value
		}
	}

	for k, v := range p.AscensionBonus {
		c.Stats[k] += v
	}

	for k, v := range p.WeaponSecondaryStat {
		c.Stats[k] += v
	}

	return &c, nil
}

func (c *TemplateChar) Name() string {
	return c.Profile.Name
}

func (c *TemplateChar) Snapshot(e combat.EleType) combat.Snapshot {
	s := combat.Snapshot{}
	s.Stats = make(map[combat.StatType]float64)
	for k, v := range c.Stats {
		s.Stats[k] = v
	}
	for key, m := range c.Mods {
		c.S.Log.Debugw("\t\tchar stat mod", "key", key, "mods", m)
		for k, v := range m {
			s.Stats[k] += v
		}
	}
	//grab field effects
	for k, v := range c.S.FieldEffects() {
		c.S.Log.Debugw("\t\tfield effect", "stat", k, "val", v)
		s.Stats[k] += v
	}
	s.CharName = c.Profile.Name
	s.BaseAtk = c.Profile.BaseAtk + c.Profile.WeaponBaseAtk
	s.CharLvl = c.Profile.Level
	s.BaseDef = c.Profile.BaseDef
	s.Element = e
	s.Stats[combat.CR] += c.Profile.BaseCR
	s.Stats[combat.CD] += c.Profile.BaseCD

	s.ResMod = make(map[combat.EleType]float64)
	s.TargetRes = make(map[combat.EleType]float64)
	s.ExtraStatMod = make(map[combat.StatType]float64)

	return s
}

func (c *TemplateChar) Attack() int {
	return 0
}

func (c *TemplateChar) Filler() int {
	return c.Attack()
}

func (c *TemplateChar) FillerFrames() int {
	return 0
}

func (c *TemplateChar) ChargeAttack() int {
	return 0
}

func (c *TemplateChar) ChargeAttackStam() float64 {
	return 0
}

func (c *TemplateChar) PlungeAttack() int {
	return 0
}

func (c *TemplateChar) Skill() int {
	return 0
}

func (c *TemplateChar) Burst() int {
	return 0
}

func (c *TemplateChar) Tick() {
	//this function gets called for every character every tick
	for k := range c.CD {
		c.CD[k]--
		if c.CD[k] == 0 {
			c.S.Log.Debugf("\t[%v] cooldown %v finished; deleting", c.S.Frame(), k)
			delete(c.CD, k)
		}
	}
	//check normal reset
	if c.NormalResetTimer == 0 {
		if c.NRTChanged {
			c.S.Log.Debugf("\t[%v] character normal reset", c.S.Frame())
			c.NRTChanged = false
		}
		c.NormalCounter = 0
	} else {
		c.NRTChanged = true
		c.NormalResetTimer--
	}
}

func (c *TemplateChar) AddMod(key string, val map[combat.StatType]float64) {
	c.Mods[key] = val
}

func (c *TemplateChar) RemoveMod(key string) {
	delete(c.Mods, key)

}

func (c *TemplateChar) HasMod(key string) bool {
	_, ok := c.Mods[key]
	return ok
}

func (c *TemplateChar) ApplyOrb(count int, ele combat.EleType, isOrb bool, isActive bool, partyCount int) {
	var amt, er, r float64
	r = 1.0
	if !isActive {
		r = 1.0 - 0.1*float64(partyCount)
	}
	//recharge amount - particles: same = 3, non-ele = 2, diff = 1
	//recharge amount - orbs: same = 9, non-ele = 6, diff = 3 (3x particles)
	switch {
	case ele == c.Profile.Element:
		amt = 3
	case ele == combat.NonElemental:
		amt = 2
	default:
		amt = 1
	}
	if isOrb {
		amt = amt * 3
	}
	amt = amt * r //apply off field reduction
	//apply energy regen stat
	er = c.Stats[combat.ER]
	for _, m := range c.Mods {
		er += m[combat.ER]
	}
	amt = amt * (1 + er) * float64(count)

	c.S.Log.Debugw("\torb", "name", c.Profile.Name, "count", count, "ele", ele, "isOrb", isOrb, "on field", isActive, "party count", partyCount)

	c.Energy += amt
	if c.Energy > c.MaxEnergy {
		c.Energy = c.MaxEnergy
	}

	c.S.Log.Debugw("\torb", "energy rec'd", amt, "current energy", c.Energy, "ER", er)

}

func (c *TemplateChar) ActionCooldown(a combat.ActionType) int {
	switch a {
	case combat.ActionTypeBurst:
		return c.CD["burst-cd"]
	case combat.ActionTypeSkill:
		return c.CD["skill-cd"]
	}
	return 0
}

func (c *TemplateChar) ActionReady(a combat.ActionType) bool {
	switch a {
	case combat.ActionTypeBurst:
		if c.Energy != c.MaxEnergy {
			return false
		}
		if _, ok := c.CD["burst-cd"]; ok {
			return false
		}
	case combat.ActionTypeSkill:
		if _, ok := c.CD["skill-cd"]; ok {
			return false
		}
	}
	return true
}

func (c *TemplateChar) ResetActionCooldown(a combat.ActionType) {
	switch a {
	case combat.ActionTypeBurst:
		delete(c.CD, BurstCD)
	case combat.ActionTypeSkill:
		delete(c.CD, SkillCD)
	}
}
