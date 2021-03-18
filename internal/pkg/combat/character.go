package combat

import (
	"go.uber.org/zap"
)

type NewCharacterFunc func(s *Sim) *Character

func RegisterCharFunc(name string, f NewCharacterFunc) {
	mu.Lock()
	defer mu.Unlock()
	if _, dup := charMap[name]; dup {
		panic("combat: RegisterChar called twice for character " + name)
	}
	charMap[name] = f
}

//Character contains all the information required to calculate
type Character struct {
	//track cooldowns in general; can be skill on field, ICD, etc...
	Cooldown map[string]int

	//we need some sort of key/val Store to Store information
	//specific to each character.
	//use to keep track of attack counter/diluc e counter, etc...
	Store map[string]interface{}

	//Init is used to add in any initial hooks to the sim
	// init func(s *Sim)

	//tickHooks are functions to be called on each tick
	TickHooks map[string]func(c *Character) bool
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

	//somehow we have to deal with artifact effects too?
	ArtifactSetBonus func(e *Enemy)

	//key Stats
	Stats map[StatType]float64
	Mods  map[string]map[StatType]float64 //special effect mods (character only)

	//character specific information; need this for damage calc
	Profile   CharacterProfile
	WeaponAtk float64
	Talent    map[ActionType]int64 //talent levels

	//other stats
	MaxEnergy  float64
	MaxStamina float64
	Energy     float64 //how much energy the character currently have
	Stamina    float64 //how much stam the character currently have
}

//CharacterProfile ...
type CharacterProfile struct {
	Name                string               `yaml:"Name"`
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
)

func (c *Character) tick(s *Sim) {
	//this function gets called for every character every tick
	for k, v := range c.Cooldown {
		if v == 0 {
			delete(c.Cooldown, k)
		} else {
			c.Cooldown[k]--
		}
	}
}

func (c *Character) Snapshot(e eleType) Snapshot {
	var s Snapshot
	s.Stats = make(map[StatType]float64)
	for k, v := range c.Stats {
		s.Stats[k] = v
	}
	//add char specific stat effect
	for x, m := range c.Mods {
		zap.S().Debugw("adding special char stat mod to snapshot", "key", x, "mods", m)
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

	return s
}
