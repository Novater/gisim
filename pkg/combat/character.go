package combat

type Character interface {
	//ability functions to be defined by each character on how they will
	Name() string
	CurrentEnergy() float64 //current energy
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

	ReceiveParticle(p Particle, isActive bool, partyCount int)
	Snapshot(name string, t ActionType, e EleType) Snapshot
	ResetActionCooldown(a ActionType)
}

//CharacterProfile ...
type CharacterProfile struct {
	Name                string               `yaml:"Name"`
	Element             EleType              `yaml:"Element"`
	Level               int64                `yaml:"Level"`
	BaseHP              float64              `yaml:"BaseHP"`
	BaseAtk             float64              `yaml:"BaseAtk"`
	BaseDef             float64              `yaml:"BaseDef"`
	BaseCR              float64              `yaml:"BaseCR"`
	BaseCD              float64              `yaml:"BaseCD"`
	Constellation       int                  `yaml:"Constellation"`
	AscensionBonus      map[StatType]float64 `yaml:"AscensionBonus"`
	TalentLevel         map[ActionType]int64 `yaml:"TalentLevel"`
	WeaponName          string               `yaml:"WeaponName"`
	WeaponClass         WeaponClass          `yaml:"WeaponClass"`
	WeaponRefinement    int                  `yaml:"WeaponRefinement"`
	WeaponBaseAtk       float64              `yaml:"WeaponBaseAtk"`
	WeaponSecondaryStat map[StatType]float64 `yaml:"WeaponSecondaryStat"`
	Artifacts           map[Slot]Artifact    `yaml:"Artifacts"`
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
	CD    map[string]int
	Stats map[StatType]float64
	Mods  map[string]map[StatType]float64
	Tags  map[string]int
	//Profile info
	Profile   CharacterProfile
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
	c := CharacterTemplate{}
	c.S = s
	c.CD = make(map[string]int)
	c.Stats = make(map[StatType]float64)
	c.Mods = make(map[string]map[StatType]float64)
	c.Tags = make(map[string]int)
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

func (c *CharacterTemplate) Tag(key string) int {
	return c.Tags[key]
}

func (c *CharacterTemplate) Name() string {
	return c.Profile.Name
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
	case p.Ele == c.Profile.Element:
		amt = 3
	case p.Ele == NonElemental:
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

	c.S.Log.Debugw("\t\t orb", "name", c.Profile.Name, "count", p.Num, "ele", p.Ele, "on field", isActive, "party count", partyCount, "pre energ", c.Energy)

	c.Energy += amt
	if c.Energy > c.MaxEnergy {
		c.Energy = c.MaxEnergy
	}

	c.S.Log.Debugw("\t\t orb", "energy rec'd", amt, "next energy", c.Energy, "ER", er)
}

func (c *CharacterTemplate) Snapshot(name string, t ActionType, e EleType) Snapshot {
	ds := Snapshot{}
	ds.Stats = make(map[StatType]float64)
	for k, v := range c.Stats {
		ds.Stats[k] = v
	}
	for key, m := range c.Mods {
		c.S.Log.Debugw("\t\t char stat mod", "key", key, "mods", m)
		for k, v := range m {
			ds.Stats[k] += v
		}
	}

	ds.Abil = name
	ds.AbilType = t
	ds.CharName = c.Profile.Name
	ds.BaseAtk = c.Profile.BaseAtk + c.Profile.WeaponBaseAtk
	ds.CharLvl = c.Profile.Level
	ds.BaseDef = c.Profile.BaseDef
	ds.Element = e
	ds.Stats[CR] += c.Profile.BaseCR
	ds.Stats[CD] += c.Profile.BaseCD

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
	//this function gets called for every character every tick
	// for k := range c.CD {
	// 	c.CD[k]--
	// 	if c.CD[k] <= 0 {
	// 		c.S.Log.Infof("[%v] cooldown %v finished; deleting", c.S.Frame(), k)
	// 		delete(c.CD, k)
	// 	}
	// }
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
