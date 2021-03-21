package combat

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	mu        sync.RWMutex
	charMap   = make(map[string]NewCharacterFunc)
	setMap    = make(map[string]NewSetFunc)
	weaponMap = make(map[string]NewWeaponFunc)
)

type AbilFunc func(s *Sim) int
type ActionFunc func(s *Sim) bool

type effectType string

const (
	PreDamageHook   effectType = "PRE_DAMAGE"
	PostDamageHook  effectType = "POST_DAMAGE"
	PreAuraAppHook  effectType = "PRE_AURA_APP"
	PostAuraAppHook effectType = "POST_AURA_APP"
	// triggered when there will be a reaction
	PreReaction  effectType = "PRE_REACTION"
	PostReaction effectType = "POST_REACTION"
)

type effectFunc func(s *Snapshot) bool

//Sim keeps track of one simulation
type Sim struct {
	Target *Enemy
	Log    *zap.SugaredLogger
	//track whatever status, ticked down by 1 each tick
	Status map[string]int

	active     int
	characters []Character
	f          int
	stam       float64
	swapCD     int
	//per tick hooks
	actions map[string]ActionFunc
	//effects
	effects map[effectType]map[string]effectFunc
	field   map[string]map[StatType]float64

	//action priority list
	priority []RotationItem
}

//New creates new sim from given profile
func New(p Profile) (*Sim, error) {
	s := &Sim{}

	u := &Enemy{}

	u.Auras = make(map[EleType]aura)
	u.Status = make(map[string]int)
	u.Level = p.Enemy.Level
	u.Resist = p.Enemy.Resist

	s.Target = u

	s.actions = make(map[string]ActionFunc)
	s.effects = make(map[effectType]map[string]effectFunc)
	s.Status = make(map[string]int)
	s.field = make(map[string]map[StatType]float64)

	s.stam = 240

	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	switch p.LogLevel {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	}
	config.EncoderConfig.TimeKey = ""
	config.EncoderConfig.StacktraceKey = ""
	config.EncoderConfig.CallerKey = ""

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	s.Log = logger.Sugar()
	zap.ReplaceGlobals(logger)

	dup := make(map[string]bool)
	s.active = -1
	//create the characters
	for i, v := range p.Characters {

		f, ok := charMap[v.Name]
		if !ok {
			return nil, fmt.Errorf("invalid character: %v", v.Name)
		}
		c, err := f(s, v)
		if err != nil {
			return nil, err
		}

		//check talent levels are valid
		for key, val := range v.TalentLevel {
			if key != ActionTypeAttack && key != ActionTypeSkill && key != ActionTypeBurst {
				return nil, fmt.Errorf("invalid talent type: %v - %v", v.Name, key)
			}
			if val < 1 || val > 15 {
				return nil, fmt.Errorf("invalid talent level: %v - %v - %v", v.Name, key, val)
			}
		}
		//check talents not missing
		if _, ok := v.TalentLevel[ActionTypeAttack]; !ok {
			return nil, fmt.Errorf("char %v missing talent level for %v", v.Name, ActionTypeAttack)
		}
		if _, ok := v.TalentLevel[ActionTypeSkill]; !ok {
			return nil, fmt.Errorf("char %v missing talent level for %v", v.Name, ActionTypeSkill)
		}
		if _, ok := v.TalentLevel[ActionTypeBurst]; !ok {
			return nil, fmt.Errorf("char %v missing talent level for %v", v.Name, ActionTypeBurst)
		}

		s.characters = append(s.characters, c)

		if v.Name == p.InitialActive {
			s.active = i
		}

		if _, ok := dup[v.Name]; ok {
			return nil, fmt.Errorf("duplicated character %v", v.Name)
		}

		dup[v.Name] = true

		//initialize weapon
		wf, ok := weaponMap[v.WeaponName]
		if !ok {
			return nil, fmt.Errorf("unrecognized weapon %v for character %v", v.WeaponName, v.Name)
		}
		wf(c, s, v.WeaponRefinement)

		//check set bonus
		sb := make(map[string]int)
		for _, a := range v.Artifacts {
			sb[a.Set]++
		}

		//add set bonus
		for key, count := range sb {
			f, ok := setMap[key]
			if ok {
				f(c, s, count)
			} else {
				s.Log.Warnf("character %v has unrecognized set %v", v.Name, key)
			}
		}

	}
	if s.active == -1 {
		return nil, fmt.Errorf("invalid active initial character %v", p.InitialActive)
	}
	for _, v := range p.Rotation {
		//find index
		index := -1
		for i, c := range s.characters {
			if c.Name() == v.CharacterName {
				index = i
				break
			}
		}
		if index == -1 {
			return nil, fmt.Errorf("invalid character %v in rotation list", v.CharacterName)
		}
		next := v
		next.index = index
		s.priority = append(s.priority, next)
	}

	return s, nil
}

func (s *Sim) AddFieldEffect(key string, val map[StatType]float64) {
	s.field[key] = val
}

func (s *Sim) RemoveFieldEffect(key string) {
	delete(s.field, key)
}

func (s *Sim) FieldEffects() map[StatType]float64 {
	r := make(map[StatType]float64)
	for _, m := range s.field {
		for t, v := range m {
			r[t] += v
		}
	}
	return r
}

//Run the sim; length in seconds
func (s *Sim) Run(length int) float64 {
	var cooldown int
	rand.Seed(time.Now().UnixNano())
	//60fps, 60s/min, 2min
	for s.f = 0; s.f < 60*length; s.f++ {
		//tick target and each character
		//target doesn't do anything, just takes punishment, so it won't affect cd
		s.Target.tick(s)
		for _, c := range s.characters {
			//character may affect cooldown by i.e. adding to it
			c.Tick()
		}

		s.tick()

		//if in cooldown, do nothing
		if cooldown > 0 {
			cooldown--
			continue
		}

		//execute first item on priority list we can execute
		//create list of cd for the priority list, execute item with lowest cd on the priority list
		var next RotationItem
		found := false
	prio:
		for _, v := range s.priority {
			//check if swap is required for this action
			//if true, then this action is only executable if swapCD = 0
			if v.index != s.active {
				if s.swapCD > 0 {
					continue prio
				}
			}
			if v.Action == ActionTypeChargedAttack {
				scost := s.characters[v.index].ChargeAttackStam()
				if scost < s.stam {
					next = v
					found = true
					break prio
				}
			} else {
				ok := s.characters[v.index].ActionReady(v.Action)
				if ok {
					next = v
					found = true
					break prio
				}
			}
		}

		if !found {
			continue //skip this frame, wait for something to become available
		}

		//check if actor is active
		if next.index != s.active {
			s.Log.Debugf("[%v] swapping to char #%v (current = %v)\n", s.f, next.index, s.active)
			//trigger a swap, add 20 to the cooldown
			cooldown = 20
			s.swapCD = 150
			s.active = next.index
		}

		cooldown += s.handleAction(s.active, next)

	}

	return s.Target.Damage
}

func (s *Sim) AddEffect(f effectFunc, key string, hook effectType) {
	if _, ok := s.effects[hook]; !ok {
		s.effects[hook] = make(map[string]effectFunc)
	}
	s.effects[hook][key] = f
}

//RemoveEffect forcefully remove an effect even if the call does not return true
func (s *Sim) RemoveEffect(key string, hook effectType) {
	delete(s.effects[hook], key)
}

func (s *Sim) AddAction(f ActionFunc, key string) {
	if _, ok := s.actions[key]; ok {
		s.Log.Debugf("\t[%v] action %v exists; overriding existing", s.Frame(), key)
	}
	s.actions[key] = f
	s.Log.Debugf("\t[%v] new action %v; action map: %v", s.Frame(), key, s.actions)
}

//GenerateOrb is called when an ability generates orb
func (s *Sim) GenerateOrb(n int, ele EleType, isOrb bool) {
	s.Log.Debugf("[%v]: particle/orbs picked up: %v of %v (isOrb: %v)", s.Frame(), n, ele, isOrb)
	count := len(s.characters)
	for i, c := range s.characters {
		active := s.active == i
		c.ApplyOrb(n, ele, isOrb, active, count)
	}
}

//tick
func (s *Sim) tick() {
	for k, f := range s.actions {
		if f(s) {
			s.Log.Debugf("\t[%v] action %v expired", s.Frame(), k)
			delete(s.actions, k)
		}
	}
	for k, v := range s.Status {
		if v == 0 {
			delete(s.Status, k)
		} else {
			s.Status[k]--
		}
	}
	if s.swapCD > 0 {
		s.swapCD--
	}
}

//handleAction executes the next action, returns the cooldown
func (s *Sim) handleAction(active int, a RotationItem) int {
	//if active see what ability we want to use
	c := s.characters[active]
	s.Log.Infof("[%v] executing %v", s.Frame(), a.Action)
	switch a.Action {
	case ActionTypeDash:
		return 30
	case ActionTypeJump:
		return 30
	case ActionTypeAttack:
		return c.Attack()
	case ActionTypeChargedAttack:
		return c.ChargeAttack()
	case ActionTypeBurst:
		return c.Burst()
	case ActionTypeSkill:
		return c.Skill()
	default:
		//do nothing
	}

	return 0
}

func (s *Sim) Frame() string {
	return fmt.Sprintf("%.2fs|%v", float64(s.f)/60, s.f)
}

type NewCharacterFunc func(s *Sim, p CharacterProfile) (Character, error)

func RegisterCharFunc(name string, f NewCharacterFunc) {
	mu.Lock()
	defer mu.Unlock()
	if _, dup := charMap[name]; dup {
		panic("combat: RegisterChar called twice for character " + name)
	}
	charMap[name] = f
}

type NewWeaponFunc func(c Character, s *Sim, refine int)

func RegisterWeaponFunc(name string, f NewWeaponFunc) {
	mu.Lock()
	defer mu.Unlock()
	if _, dup := weaponMap[name]; dup {
		panic("combat: RegisterWeapon called twice for character " + name)
	}
	weaponMap[name] = f
}

//Action describe one action to execute
type Action struct {
	TargetCharIndex int
	Type            ActionType
}

type Profile struct {
	Label         string             `yaml:"Label"`
	Enemy         EnemyProfile       `yaml:"Enemy"`
	InitialActive string             `yaml:"InitialActive"`
	Characters    []CharacterProfile `yaml:"Characters"`
	Rotation      []RotationItem     `yaml:"Rotation"`
	LogLevel      string             `yaml:"LogLevel"`
}

//EnemyProfile ...
type EnemyProfile struct {
	Level  int64               `yaml:"Level"`
	Resist map[EleType]float64 `yaml:"Resist"`
}

//RotationItem ...
type RotationItem struct {
	CharacterName   string `yaml:"CharacterName"`
	index           int
	Action          ActionType `yaml:"Action"`
	ConditionType   string     `yaml:"ConditionType"`   //for now either a status or aura
	ConditionTarget string     `yaml:"ConditionTarget"` //which status or aura
	Condition       bool       `yaml:"Condition"`       //true or false
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
	Filler() int //action to be called when we don't want to switch and need something to fill time until cd comes up
	Tick()       //function to be called every frame
	//special char mods
	AddMod(key string, val map[StatType]float64)
	RemoveMod(key string)
	//info methods
	HasMod(key string) bool
	ActionCooldown(a ActionType) int
	ActionReady(a ActionType) bool
	ChargeAttackStam() float64
	FillerFrames() int
	//other actions
	ApplyOrb(count int, ele EleType, isOrb bool, isActive bool, partyCount int)
	Snapshot(e EleType) Snapshot
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
)

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
