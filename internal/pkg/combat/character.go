package combat

import (
	"go.uber.org/zap"
)

type NewCharacterFunc func(s *Sim, p CharacterProfile) (Character, error)

func RegisterCharFunc(name string, f NewCharacterFunc) {
	mu.Lock()
	defer mu.Unlock()
	if _, dup := charMap[name]; dup {
		panic("combat: RegisterChar called twice for character " + name)
	}
	charMap[name] = f
}

type Character interface {
	//ability functions to be defined by each character on how they will
	Name() string
	//affect the unit
	Attack() int
	ChargeAttack() int
	PlungeAttack() int
	Skill() int
	Burst() int
	Tick() //function to be called every frame
	//special char mods
	AddMod(key string, val map[StatType]float64)
	RemoveMod(key string)
	HasMod(key string) bool
	//other actions
	ApplyOrb(count int, ele EleType, isOrb bool, isActive bool, partyCount int)
	ActionCooldown(a ActionType) int
}

//Char contains all the information required to calculate
type Char struct {
	//track cooldowns in general; can be skill on field, ICD, etc...
	Cooldown map[string]int

	//we need some sort of key/val Store to Store information
	//specific to each character.
	//use to keep track of attack counter/diluc e counter, etc...
	Store map[string]interface{}

	//Init is used to add in any initial hooks to the sim
	// init func(s *Sim)

	//tickHooks are functions to be called on each tick
	TickHooks map[string]func(c *Char) bool
	//this is useful for on field effect such as gouba/oz/pyronado
	//we can use store to keep track of the uptime on gouba/oz/pyronado/taunt etc..
	//for something like baron bunny, if uptime = xx, then trigger damage
	// tickHooks map[string]func(s *Sim) bool
	//what about something like bennett ult or ganyu ult that affects char in the field?? this hook can only affect current actor?

	//ability functions to be defined by each character on how they will
	//affect the unit
	Attack       func(s *Sim) int
	ChargeAttack func(s *Sim) int
	PlungeAttack func(s *Sim) int
	Skill        func(s *Sim) int
	Burst        func(s *Sim) int

	//return how many more frames until specified action comes off CD
	ActionCooldown func(a ActionType) int

	//key Stats
	Stats map[StatType]float64
	Mods  map[string]map[StatType]float64 //special effect mods (character only)

	//character specific information; need this for damage calc
	Profile   CharacterProfile
	WeaponAtk float64
	Talent    map[ActionType]int64 //talent levels

	//other stats
	MaxEnergy          float64
	Energy             float64 //how much energy the character currently have
	NormalCounter      int     //which attack in the series are we at now
	NormalResetTimer   int     //how many frames until normal reset
	normalTimerChanged bool
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
	WeaponRefinement    int                  `yaml:"WeaponRefinement"`
	WeaponBaseAtk       float64              `yaml:"WeaponBaseAtk"`
	WeaponSecondaryStat map[StatType]float64 `yaml:"WeaponSecondaryStat"`
	Artifacts           map[Slot]Artifact    `yaml:"Artifacts"`
}

type ActionType string

//ActionType constants
const (
	//motions
	ActionTypeSwap ActionType = "swap"
	ActionTypeDash ActionType = "dash"
	ActionTypeJump ActionType = "jump"
	//main actions
	ActionTypeAttack ActionType = "attack"
	ActionTypeSkill  ActionType = "skill"
	ActionTypeBurst  ActionType = "burst"
	//derivative actions
	ActionTypeChargedAttack ActionType = "charge"
	ActionTypePlungeAttack  ActionType = "plunge"
	//procs
	ActionTypeWeaponProc ActionType = "proc"
	//xiao special
	ActionTypeXiaoLowJump  ActionType = "xiao-low-jump"
	ActionTypeXiaoHighJump ActionType = "xiao-high-jump"
)

func (c *Char) CancelNormal() {
	c.NormalCounter = 0
}

func (c *Char) tick(s *Sim) {
	//this function gets called for every character every tick
	for k, v := range c.Cooldown {
		if v == 0 {
			s.Log.Debugf("\t[%v] cooldown %v finished; deleting", PrintFrames(s.Frame), k)
			delete(c.Cooldown, k)
		} else {
			c.Cooldown[k]--
		}
	}
	for k, f := range c.TickHooks {
		if f(c) {
			s.Log.Debugf("\t[%v] character hook %v expired", PrintFrames(s.Frame), k)
		}
	}
	//check normal reset
	if c.NormalResetTimer == 0 {
		if c.normalTimerChanged {
			s.Log.Debugf("\t[%v] character normal reset", PrintFrames(s.Frame))
			c.normalTimerChanged = false
		}
		c.NormalCounter = 0
	} else {
		c.normalTimerChanged = true
		c.NormalResetTimer--
	}
}

func (c *Char) AddHook(key string, f func(c *Char) bool) {
	c.TickHooks[key] = f
}

func (c *Char) applyOrb(count int, ele EleType, isOrb bool, isActive bool, partyCount int) {
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
	case ele == NonElemental:
		amt = 2
	default:
		amt = 1
	}
	if isOrb {
		amt = amt * 3
	}
	amt = amt * r //apply off field reduction
	//apply energy regen stat
	er = c.Stats[ER]
	for _, m := range c.Mods {
		er += m[ER]
	}
	amt = amt * (1 + er) * float64(count)

	zap.S().Debugw("\torb", "count", count, "ele", ele, "isOrb", isOrb, "on field", isActive, "party count", partyCount)

	c.Energy += amt
	if c.Energy > c.MaxEnergy {
		c.Energy = c.MaxEnergy
	}

	zap.S().Debugw("\torb", "energy rec'd", amt, "current energy", c.Energy, "ER", er)

}

func (c *Char) Snapshot(e EleType) Snapshot {
	var s Snapshot
	s.Stats = make(map[StatType]float64)

	for k, v := range c.Stats {
		s.Stats[k] = v
	}
	//add char specific stat effect
	for x, m := range c.Mods {
		zap.S().Debugw("\t\tchar stat mod", "key", x, "mods", m)
		for k, v := range m {
			s.Stats[k] += v
		}
	}
	//add field effects

	//other stats
	s.CharName = c.Profile.Name
	s.BaseAtk = c.Profile.BaseAtk + c.WeaponAtk
	s.CharLvl = c.Profile.Level
	s.BaseDef = c.Profile.BaseDef
	s.Element = e

	s.Stats[CR] += c.Profile.BaseCR
	s.Stats[CD] += c.Profile.BaseCD

	//maps for other mods
	s.ResMod = make(map[EleType]float64)
	s.TargetRes = make(map[EleType]float64)
	s.ExtraStatMod = make(map[StatType]float64)

	return s
}

func (s *Snapshot) Clone() Snapshot {
	c := Snapshot{}
	c = *s
	c.ResMod = make(map[EleType]float64)
	c.TargetRes = make(map[EleType]float64)
	c.ExtraStatMod = make(map[StatType]float64)
	return c
}
