package combat

import (
	"fmt"

	"github.com/srliao/gisim/internal/rotation"
)

type Character interface {
	//ability functions to be defined by each character on how they will
	Name() string
	CurrentEnergy() float64 //current energy
	TalentLvlSkill() int
	TalentLvlAttack() int
	TalentLvlBurst() int
	//affect the unit
	Attack(p int) int
	Aimed(p int) int
	ChargeAttack(p int) int
	PlungeAttack(p int) int
	Skill(p int) int
	Burst(p int) int
	Tick() //function to be called every frame
	//special char mods
	AddMod(key string, val map[StatType]float64)
	RemoveMod(key string)
	//info methods
	HasMod(key string) bool
	ActionReady(a rotation.ActionType) bool
	ChargeAttackStam() float64

	//other actions
	UnsafeSetStats(stats []float64)

	//status stuff
	Cooldown(a rotation.ActionType) int
	Tag(key string) int

	ReceiveParticle(p Particle, isActive bool, partyCount int)
	Snapshot(name string, t rotation.ActionType, e EleType, d float64) Snapshot
	ResetActionCooldown(a rotation.ActionType)
	ResetNormalCounter()
}

type WeaponClass string

const (
	WeaponClassSword    WeaponClass = "sword"
	WeaponClassClaymore WeaponClass = "claymore"
	WeaponClassSpear    WeaponClass = "spear"
	WeaponClassBow      WeaponClass = "bow"
	WeaponClassCatalyst WeaponClass = "catalyst"
)

func (s *Sim) executeCharacterTicks() {
	for _, c := range s.Chars {
		c.Tick()
	}
}

type CharacterTemplate struct {
	S *Sim
	//this should describe the frame in which the abil becomes available
	//if frame > current then it's available. no need to decrement this way
	CD   map[string]int
	Mods map[string]map[StatType]float64
	Tags map[string]int
	//Profile info
	Base    CharacterBase
	Weapon  WeaponProfile
	Stats   []float64
	Talents TalentProfile

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

func NewTemplateChar(s *Sim, p CharacterProfile) (*CharacterTemplate, error) {
	//error checks
	if len(p.ArtifactsConfig) > 5 {
		return nil, fmt.Errorf("number of artifacts exceeds 5 - %v", p.Base.Name)
	}
	c := CharacterTemplate{}
	c.S = s
	c.CD = make(map[string]int)
	c.Mods = make(map[string]map[StatType]float64)
	c.Tags = make(map[string]int)
	c.Base = p.Base
	c.Weapon = p.Weapon
	c.Talents.Attack = p.TalentLevelConfig["attack"]
	if c.Talents.Attack < 1 || c.Talents.Attack > 15 {
		return nil, fmt.Errorf("invalid talent lvl: attack - %v", c.Talents.Attack)
	}
	c.Talents.Skill = p.TalentLevelConfig["skill"]
	if c.Talents.Attack < 1 || c.Talents.Attack > 12 {
		return nil, fmt.Errorf("invalid talent lvl: skill - %v", c.Talents.Skill)
	}
	c.Talents.Burst = p.TalentLevelConfig["burst"]
	if c.Talents.Attack < 1 || c.Talents.Attack > 12 {
		return nil, fmt.Errorf("invalid talent lvl: burst - %v", c.Talents.Burst)
	}
	c.Stats = make([]float64, len(StatTypeString))
	//load artifacts
	for _, a := range p.ArtifactsConfig {
		//find out which main stat it is

		//load sub stats
		for i := 0; i < len(c.Stats); i++ {
			c.Stats[i] += a.Main[StatTypeString[i]]
			c.Stats[i] += a.Sub[StatTypeString[i]]
		}
		// s.Log.Debugw("loading artifacts", "a", a, "stats", c.Stats)
	}
	//load weapon and ascension bonus
	for i := 0; i < len(c.Stats); i++ {
		c.Stats[i] += p.WeaponBonusConfig[StatTypeString[i]]
		c.Stats[i] += p.AscensionBonusConfig[StatTypeString[i]]
	}

	return &c, nil
}

func (c *CharacterTemplate) UnsafeSetStats(stats []float64) {
	copy(c.Stats, stats)
}

func (c *CharacterTemplate) Tag(key string) int {
	return c.Tags[key]
}

func (c *CharacterTemplate) Name() string {
	return c.Base.Name
}

func (c *CharacterTemplate) TalentLvlSkill() int {
	if c.Base.Cons >= 3 {
		return c.Talents.Skill + 2
	}
	return c.Talents.Skill - 1
}
func (c *CharacterTemplate) TalentLvlBurst() int {
	if c.Base.Cons >= 5 {
		return c.Talents.Burst + 2
	}
	return c.Talents.Burst - 1
}
func (c *CharacterTemplate) TalentLvlAttack() int {
	if c.S.GlobalFlags.ChildeActive {
		return c.Talents.Attack
	}
	return c.Talents.Attack - 1
}

func (c *CharacterTemplate) CurrentEnergy() float64 {
	return c.Energy
}

func (c *CharacterTemplate) ReceiveParticle(p Particle, isActive bool, partyCount int) {
	var amt, er, r float64
	r = 1.0
	if !isActive {
		r = 1.0 - 0.1*float64(partyCount)
	}
	//recharge amount - particles: same = 3, non-ele = 2, diff = 1
	//recharge amount - orbs: same = 9, non-ele = 6, diff = 3 (3x particles)
	switch {
	case p.Ele == c.Base.Element:
		amt = 3
	case p.Ele == NoElement:
		amt = 2
	default:
		amt = 1
	}
	amt = amt * r //apply off field reduction
	//apply energy regen stat
	er = c.Stats[ER]
	for _, m := range c.Mods {
		er += m[ER]
	}
	amt = amt * (1 + er) * float64(p.Num)

	c.S.Log.Debugw("\t\t orb", "name", c.Base.Name, "count", p.Num, "ele", p.Ele, "on field", isActive, "party count", partyCount, "pre energ", c.Energy)

	c.Energy += amt
	if c.Energy > c.MaxEnergy {
		c.Energy = c.MaxEnergy
	}

	c.S.Log.Debugw("\t\t orb", "energy rec'd", amt, "next energy", c.Energy, "ER", er)
}

func (c *CharacterTemplate) Snapshot(name string, t rotation.ActionType, e EleType, d float64) Snapshot {
	ds := Snapshot{}
	ds.Stats = make([]float64, len(c.Stats))
	copy(ds.Stats, c.Stats)

	for key, m := range c.Mods {
		c.S.Log.Debugw("\t\t char stat mod", "key", key, "mods", m)
		for k, v := range m {
			ds.Stats[k] += v
		}
	}

	ds.Abil = name
	ds.AbilType = t
	ds.Actor = c.Base.Name
	ds.BaseAtk = c.Base.Atk + c.Weapon.Atk
	ds.CharLvl = c.Base.Level
	ds.BaseDef = c.Base.Def
	ds.Element = e
	ds.Durability = d
	ds.Stats[CR] += c.Base.CR
	ds.Stats[CD] += c.Base.CD

	for _, f := range c.S.snapshotHooks[PostSnapshot] {
		f(&ds)
	}

	return ds
}

func (c *CharacterTemplate) Attack(p int) int {
	return 0
}

func (c *CharacterTemplate) AttackHelperSingle(frames []int, delay []int, mult [][]float64) int {
	reset := c.NormalCounter >= len(frames)-1

	//apply attack speed
	f := int(float64(frames[c.NormalCounter]) / (1 + c.Stats[AtkSpd]))

	x := c.Snapshot("Normal", rotation.ActionAttack, Physical, WeakDurability)
	x.Mult = mult[c.NormalCounter][c.TalentLvlAttack()]

	c.S.AddTask(func(s *Sim) {
		damage, str := s.ApplyDamage(x)
		s.Log.Infof("\t %v normal %v dealt %.0f damage [%v]", c.Base.Name, c.NormalCounter, damage, str)
	}, fmt.Sprintf("%v-Normal-%v", c.Base.Name, c.NormalCounter), delay[c.NormalCounter])

	c.NormalCounter++

	//add a 75 frame attackcounter reset
	c.NormalResetTimer = 70

	if reset {
		c.NormalResetTimer = 0
		c.NormalCounter = 0
	}
	//return animation cd
	//this also depends on which hit in the chain this is
	return f
}

func (c *CharacterTemplate) Aimed(p int) int {
	return 0
}

func (c *CharacterTemplate) ChargeAttack(p int) int {
	return 0
}

func (c *CharacterTemplate) ChargeAttackStam() float64 {
	return 0
}

func (c *CharacterTemplate) PlungeAttack(p int) int {
	return 0
}

func (c *CharacterTemplate) Skill(p int) int {
	return 0
}

func (c *CharacterTemplate) Burst(p int) int {
	return 0
}

func (c *CharacterTemplate) Tick() {

	//check normal reset
	if c.NormalResetTimer == 0 {
		if c.NRTChanged {
			c.S.Log.Infof("[%v] character normal reset", c.S.Frame())
			c.NRTChanged = false
		}
		c.NormalCounter = 0
	} else {
		c.NRTChanged = true
		c.NormalResetTimer--
	}
}

func (c *CharacterTemplate) AddMod(key string, val map[StatType]float64) {
	c.Mods[key] = val
}

func (c *CharacterTemplate) RemoveMod(key string) {
	delete(c.Mods, key)

}

func (c *CharacterTemplate) HasMod(key string) bool {
	_, ok := c.Mods[key]
	return ok
}

func (c *CharacterTemplate) ActionReady(a rotation.ActionType) bool {
	switch a {
	case rotation.ActionBurst:
		if c.Energy != c.MaxEnergy {
			return false
		}
		return c.CD[BurstCD] <= c.S.F

	case rotation.ActionSkill:
		return c.CD[SkillCD] <= c.S.F
	}
	return true
}

func (c *CharacterTemplate) Cooldown(a rotation.ActionType) int {
	cd := 0
	switch a {
	case rotation.ActionBurst:
		cd = c.CD[BurstCD] - c.S.F
	case rotation.ActionSkill:
		cd = c.CD[SkillCD] - c.S.F
	default:
		return -1
	}
	if cd < 0 {
		cd = 0
	}
	return cd
}

func (c *CharacterTemplate) ResetActionCooldown(a rotation.ActionType) {
	switch a {
	case rotation.ActionBurst:
		delete(c.CD, BurstCD)
	case rotation.ActionSkill:
		delete(c.CD, SkillCD)
	}
}

func (c *CharacterTemplate) ResetNormalCounter() {
	c.NormalResetTimer = 0
}
