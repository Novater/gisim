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

type NewWeaponFunc func(c Character, s *Sim, refine int)

func RegisterWeaponFunc(name string, f NewWeaponFunc) {
	mu.Lock()
	defer mu.Unlock()
	if _, dup := weaponMap[name]; dup {
		panic("combat: RegisterWeapon called twice for character " + name)
	}
	weaponMap[name] = f
}

type AbilFunc func(s *Sim) int
type ActionFunc func(s *Sim) bool

type effectType string

const (
	PreDamageHook   effectType = "PRE_DAMAGE"
	PostDamageHook  effectType = "POST_DAMAGE"
	PreAuraAppHook  effectType = "PRE_AURA_APP"
	PostAuraAppHook effectType = "POST_AURA_APP"
	// triggered when there will be a reaction
	PreReaction effectType = "PRE_REACTION"
	// triggered pre damage calculated; this snapshot contains info about the
	// particular reaction. important because one damage can trigger multiple
	// PreReactionDamage, one for each reaction i.e. pryo applied to electro+water
	// trigering vape and overload at the same time
	PreReactionDamage effectType = "PRE_REACTION_DAMAGE"

	PreOverload       effectType = "PRE_OVERLOAD"
	PreSuperconduct   effectType = "PRE_SUPERCONDUCT"
	PreFreeze         effectType = "PRE_FREEZE"
	PreVaporize       effectType = "PRE_VAPORIZE"
	PreMelt           effectType = "PRE_MELT"
	PreSwirl          effectType = "PRE_SWIRL"
	PreCrystallize    effectType = "PRE_CRYSTALLIZE"
	PreElectroCharged effectType = "PRE_ELECTROCHARGED"
)

type effectFunc func(s *Snapshot) bool

//Sim keeps track of one simulation
type Sim struct {
	Target     *Enemy
	characters []Character
	Active     int
	Frame      int

	Log *zap.SugaredLogger

	//per tick hooks
	actions map[string]ActionFunc
	//effects
	effects map[effectType]map[string]effectFunc

	//action priority list
	priority []RotationItem

	field map[string]map[StatType]float64
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
	//create the characters
	for _, v := range p.Characters {

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
				f(&c, s, count)
			} else {
				s.Log.Warnf("character %v has unrecognized set %v", v.Name, key)
			}
		}

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
func (s *Sim) Run(length int, list []Action) float64 {
	var cooldown int
	var active int //index of the currently active car
	var i int
	rand.Seed(time.Now().UnixNano())
	//60fps, 60s/min, 2min
	for s.Frame = 0; s.Frame < 60*length; s.Frame++ {
		//tick target and each character
		//target doesn't do anything, just takes punishment, so it won't affect cd
		s.Target.tick(s)
		for _, c := range s.characters {
			//character may affect cooldown by i.e. adding to it
			c.Tick()
		}

		s.handleTick()

		//if in cooldown, do nothing
		if cooldown > 0 {
			cooldown--
			continue
		}

		//execute first item on priority list we can execute
		//create list of cd for the priority list, execute item with lowest cd on the priority list
		// var cd []int
		for _, v := range s.priority {
			t := 0
			//check if same character
			if v.index != s.Active {
				t += 10 //add frame lag to switch char
			}
			//check abil cd
			switch v.Action {
			case ActionTypeBurst:
				t += s.characters[v.index].ActionCooldown(v.Action)
			case ActionTypeSkill:
				t += s.characters[v.index].ActionCooldown(v.Action)
			case ActionTypeChargedAttack:
				//check stam
			}

		}

		if i >= len(list) {
			//start over
			i = 0
		}
		//otherwise only either action or swaps can trigger cooldown
		//we figure out what the next action is to be
		next := list[i]

		//check if actor is active
		if next.TargetCharIndex != active {
			fmt.Printf("[%v] swapping to char #%v (current = %v)\n", s.Frame, next.TargetCharIndex, active)
			//trigger a swap
			cooldown = 150
			active = next.TargetCharIndex
			continue

		}
		//move on to next action on list
		i++

		cooldown = s.handleAction(active, next)

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
		s.Log.Debugf("\t[%v] action %v exists; overriding existing", PrintFrames(s.Frame), key)
	}
	s.actions[key] = f
	s.Log.Debugf("\t[%v] new action %v; action map: %v", PrintFrames(s.Frame), key, s.actions)
}

//GenerateOrb is called when an ability generates orb
func (s *Sim) GenerateOrb(n int, ele EleType, isOrb bool) {
	s.Log.Debugf("\t[%v]: particle/orbs picked up: %v of %v (isOrb: %v)", PrintFrames(s.Frame), n, ele, isOrb)
	count := len(s.characters)
	for i, c := range s.characters {
		active := s.Active == i
		c.ApplyOrb(n, ele, isOrb, active, count)
	}
}

//handleTick
func (s *Sim) handleTick() {
	for k, f := range s.actions {
		if f(s) {
			print(s.Frame, true, "action %v expired", k)
			delete(s.actions, k)
		}
	}
}

//handleAction executes the next action, returns the cooldown
func (s *Sim) handleAction(active int, a Action) int {
	//if active see what ability we want to use
	c := s.characters[active]

	switch a.Type {
	case ActionTypeDash:
		print(s.Frame, false, "dashing")
		return 100
	case ActionTypeJump:
		print(s.Frame, false, "dashing")
		fmt.Printf("[%v] jumping\n", s.Frame)
		return 100
	case ActionTypeAttack:
		print(s.Frame, false, "%v executing attack", c.Name())
		return c.Attack()
	case ActionTypeChargedAttack:
		print(s.Frame, false, "%v executing charged attack", c.Name())
		return c.ChargeAttack()
	case ActionTypeBurst:
		print(s.Frame, false, "%v executing burst", c.Name())
		return c.Burst()
	case ActionTypeSkill:
		print(s.Frame, false, "%v executing skill", c.Name())
		return c.Skill()
	default:
		//do nothing
		print(s.Frame, false, "no action specified: %v. Doing nothing", a.Type)
	}

	return 0
}

//Action describe one action to execute
type Action struct {
	TargetCharIndex int
	Type            ActionType
}

type Profile struct {
	Label      string             `yaml:"Label"`
	Enemy      EnemyProfile       `yaml:"Enemy"`
	Characters []CharacterProfile `yaml:"Characters"`
	Rotation   []RotationItem     `yaml:"Rotation"`
	LogLevel   string             `yaml:"LogLevel"`
}

//EnemyProfile ...
type EnemyProfile struct {
	Level  int64               `yaml:"Level"`
	Resist map[EleType]float64 `yaml:"Resist"`
}

//RotationItem ...
type RotationItem struct {
	CharacterName string `yaml:"CharacterName"`
	index         int
	Action        ActionType `yaml:"Action"`
	Condition     string     //to be implemented
}
