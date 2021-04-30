package combat

import (
	"fmt"
)

type Character interface {
	//ability functions to be defined by each character on how they will
	Name() string
	WeaponClass() WeaponClass
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
	Dash(p int) int
	Tick() //function to be called every frame
	//special char mods
	AddMod(key string, val map[StatType]float64)
	RemoveMod(key string)
	//info methods
	HasMod(key string) bool
	ActionReady(a ActionType) bool
	ActionFrames(a ActionType, p int) int
	ActionStam(a ActionType, p int) float64

	//other actions
	UnsafeSetStats(stats []float64)

	//status stuff
	Cooldown(a ActionType) int
	Tag(key string) int

	ReceiveParticle(p Particle, isActive bool, partyCount int)
	AddEnergy(e float64)
	Snapshot(name string, t ActionType, e EleType, d float64) Snapshot
	ResetActionCooldown(a ActionType)
	ReduceActionCooldown(a ActionType, v int)
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

type CharStatMod struct {
	Mod    map[StatType]float64
	Expiry int
}

type CharacterTemplate struct {
	S *Sim
	//this should describe the frame in which the abil becomes available
	//if frame > current then it's available. no need to decrement this way
	CD        map[string]int
	Mods      map[string]map[StatType]float64
	TimedMods map[string]CharStatMod
	Tags      map[string]int
	//Profile info
	Base    CharacterBase
	Weapon  WeaponProfile
	Stats   []float64
	Talents TalentProfile

	Energy    float64
	MaxEnergy float64

	//Tasks specific to the character to be executed at set frames
	Tasks map[int][]CharTask
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
	c.Mods = make(map[string]map[StatType]float64)
	c.Tags = make(map[string]int)
	c.Base = p.Base
	c.Weapon = p.Weapon
	c.Talents = p.Talents
	if c.Talents.Attack < 1 || c.Talents.Attack > 15 {
		return nil, fmt.Errorf("invalid talent lvl: attack - %v", c.Talents.Attack)
	}
	if c.Talents.Attack < 1 || c.Talents.Attack > 12 {
		return nil, fmt.Errorf("invalid talent lvl: skill - %v", c.Talents.Skill)
	}
	if c.Talents.Attack < 1 || c.Talents.Attack > 12 {
		return nil, fmt.Errorf("invalid talent lvl: burst - %v", c.Talents.Burst)
	}
	c.Stats = make([]float64, len(StatTypeString))
	for i, v := range p.Stats {
		c.Stats[i] = v
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

func (c *CharacterTemplate) WeaponClass() WeaponClass {
	return c.Weapon.Class
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

func (c *CharacterTemplate) AddEnergy(e float64) {
	c.Energy += e
	if c.Energy > c.MaxEnergy {
		c.Energy = c.MaxEnergy
	}
	c.S.Log.Debugw("\t\t adding energy", "rec'd", e, "next energy", c.Energy)
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

	c.S.Log.Debugf("\t\t +++++++++++++++++++++++++++++++++++++++++++++++++")
	c.S.Log.Debugf("\t\t +[%v] snapshot for %v : %v (action: %v, ele: %v, dur: %v)", c.S.Frame(), c.Base.Name, name, t, e, d)

	for key, m := range c.Mods {
		c.S.Log.Debugw("\t\t +char stat mod", "key", key, "mods", m)
		for k, v := range m {
			ds.Stats[k] += v
		}
	}

	for key, m := range c.TimedMods {
		if m.Expiry > c.S.F {
			c.S.Log.Debugw("\t\t +timed char stat mod", "key", key, "mods", m.Mod, "expires in", m.Expiry-c.S.F)
			for k, v := range m.Mod {
				ds.Stats[k] += v
			}
		} else {
			delete(c.TimedMods, key)
		}
	}

	ds.Abil = name
	ds.AbilType = t
	ds.Actor = c.Base.Name
	ds.ActorEle = c.Base.Element
	ds.SourceFrame = c.S.F
	ds.BaseAtk = c.Base.Atk + c.Weapon.Atk
	ds.CharLvl = c.Base.Level
	ds.BaseDef = c.Base.Def
	ds.Element = e
	ds.Durability = d

	c.S.Log.Debugf("\t\t +Calling snaphot functions")
	for k, f := range c.S.snapshotHooks[PostSnapshot] {
		c.S.Log.Debugf("\t\t +Applying %v", k)
		if f(&ds) {
			c.S.Log.Debugf("\t\t +Deleting %v", k)
			delete(c.S.snapshotHooks[PostSnapshot], k)
		}
	}

	c.S.Log.Debugf("\t\t +Final stats %v", PrettyPrintStats(ds.Stats))

	c.S.Log.Debugf("\t\t +++++++++++++++++++++++++++++++++++++++++++++++++")

	return ds
}

func (c *CharacterTemplate) Attack(p int) int {
	return 0
}

func (c *CharacterTemplate) AttackHelperSingle(frames []int, delay []int, mult [][]float64) int {
	reset := c.NormalCounter >= len(frames)-1

	//apply attack speed
	f := int(float64(frames[c.NormalCounter]) / (1 + c.Stats[AtkSpd]))

	x := c.Snapshot("Normal", ActionAttack, Physical, WeakDurability)
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

func (c *CharacterTemplate) PlungeAttack(p int) int {
	return 0
}

func (c *CharacterTemplate) Skill(p int) int {
	return 0
}

func (c *CharacterTemplate) Burst(p int) int {
	return 0
}

func (c *CharacterTemplate) Dash(p int) int {
	return 24
}

func (c *CharacterTemplate) AddTimedMod(key string, val map[StatType]float64, expiry int) {
	c.TimedMods[key] = CharStatMod{
		Mod:    val,
		Expiry: expiry,
	}
}

func (c *CharacterTemplate) HasTimedMod(key string) bool {
	_, ok := c.TimedMods[key]
	return ok
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

func (c *CharacterTemplate) ActionStam(a ActionType, p int) float64 {
	switch a {
	case ActionDash:
		return 15
	}
	c.S.Log.Warnf("%v ActionStam not implemented; Character stam usage may be incorrect", c.Base.Name)
	return 0
}

func (c *CharacterTemplate) ActionFrames(a ActionType, p int) int {
	c.S.Log.Warnf("%v ActionFrames not implemented; Character frame count may be incorrect", c.Base.Name)
	return 0
}

func (c *CharacterTemplate) ActionReady(a ActionType) bool {
	switch a {
	case ActionBurst:
		if c.Energy != c.MaxEnergy {
			return false
		}
		return c.CD[BurstCD] <= c.S.F
	case ActionSkill:
		return c.CD[SkillCD] <= c.S.F
	}
	return true
}

func (c *CharacterTemplate) Cooldown(a ActionType) int {
	cd := 0
	switch a {
	case ActionBurst:
		cd = c.CD[BurstCD] - c.S.F
	case ActionSkill:
		cd = c.CD[SkillCD] - c.S.F
	default:
		return -1
	}
	if cd < 0 {
		cd = 0
	}
	return cd
}

func (c *CharacterTemplate) ResetActionCooldown(a ActionType) {
	switch a {
	case ActionBurst:
		delete(c.CD, BurstCD)
	case ActionSkill:
		delete(c.CD, SkillCD)
	}
}

func (c *CharacterTemplate) ReduceActionCooldown(a ActionType, v int) {
	switch a {
	case ActionBurst:
		c.CD[BurstCD] -= v
	case ActionSkill:
		c.CD[SkillCD] -= v
	}
}

func (c *CharacterTemplate) ResetNormalCounter() {
	c.NormalResetTimer = 0
	c.NormalCounter = 0
}

func (c *CharacterTemplate) Tick() {

	//check normal reset
	// if c.NormalResetTimer == 0 {
	// 	if c.NRTChanged {
	// 		c.S.Log.Infof("[%v] character normal reset", c.S.Frame())
	// 		c.NRTChanged = false
	// 	}
	// 	c.NormalCounter = 0
	// } else {
	// 	c.NRTChanged = true
	// 	c.NormalResetTimer--
	// }

	if c.NormalResetTimer > 0 {
		c.NormalResetTimer--
		if c.NormalResetTimer == 0 {
			c.NormalCounter = 0
			c.S.Log.Infof("[%v] character normal reset", c.S.Frame())
		}
	}

	//run tasks
	for _, t := range c.Tasks[c.S.F] {
		t.F()
	}

	delete(c.Tasks, c.S.F)
}

type CharTask struct {
	Name        string
	F           func()
	Delay       int
	originFrame int
}

func (c *CharacterTemplate) AddTask(f func(), name string, delay int) {
	c.Tasks[c.S.F+delay] = append(c.Tasks[c.S.F+delay], CharTask{
		Name:        name,
		F:           f,
		Delay:       delay,
		originFrame: c.S.F,
	})
}
