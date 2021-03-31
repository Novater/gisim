package combat

import "fmt"

type Character interface {
	//ability functions to be defined by each character on how they will
	Name() string
	CurrentEnergy() float64 //current energy
	TalentLvlSkill() int
	TalentLvlAttack() int
	TalentLvlBurst() int
	//affect the unit
	Attack(p map[string]interface{}) int
	Aimed(p map[string]interface{}) int
	ChargeAttack(p map[string]interface{}) int
	PlungeAttack(p map[string]interface{}) int
	Skill(p map[string]interface{}) int
	Burst(p map[string]interface{}) int
	Tick() //function to be called every frame
	//special char mods
	AddMod(key string, val map[StatType]float64)
	RemoveMod(key string)
	//info methods
	HasMod(key string) bool
	ActionReady(a ActionType) bool
	ChargeAttackStam() float64
	Tag(key string) int
	//other actions
	UnsafeSetStats(stats []float64)

	ReceiveParticle(p Particle, isActive bool, partyCount int)
	Snapshot(name string, t ActionType, e EleType, d float64) Snapshot
	ResetActionCooldown(a ActionType)
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
		s.Log.Debugw("loading artifacts", "a", a, "stats", c.Stats)
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

func (c *CharacterTemplate) Snapshot(name string, t ActionType, e EleType, d float64) Snapshot {
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

func (c *CharacterTemplate) Attack(p map[string]interface{}) int {
	return 0
}

func (c *CharacterTemplate) Aimed(p map[string]interface{}) int {
	return 0
}

func (c *CharacterTemplate) ChargeAttack(p map[string]interface{}) int {
	return 0
}

func (c *CharacterTemplate) ChargeAttackStam() float64 {
	return 0
}

func (c *CharacterTemplate) PlungeAttack(p map[string]interface{}) int {
	return 0
}

func (c *CharacterTemplate) Skill(p map[string]interface{}) int {
	return 0
}

func (c *CharacterTemplate) Burst(p map[string]interface{}) int {
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

func (c *CharacterTemplate) ActionReady(a ActionType) bool {
	switch a {
	case ActionTypeBurst:
		if c.Energy != c.MaxEnergy {
			return false
		}
		return c.CD[BurstCD] <= c.S.F

	case ActionTypeSkill:
		return c.CD[SkillCD] <= c.S.F
	}
	return true
}

func (c *CharacterTemplate) ResetActionCooldown(a ActionType) {
	switch a {
	case ActionTypeBurst:
		delete(c.CD, BurstCD)
	case ActionTypeSkill:
		delete(c.CD, SkillCD)
	}
}
