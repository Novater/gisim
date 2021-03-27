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

type EffectFunc func(s *Sim) bool

//Sim keeps track of one simulation
type Sim struct {
	Target *Enemy
	Log    *zap.SugaredLogger
	//track whatever status, ticked down by 1 each tick
	Status map[string]int

	particles []Particle
	tasks     map[string]Task

	ActiveChar string
	chars      map[string]Character
	f          int
	stam       float64
	swapCD     int
	//per tick effects
	effects []EffectFunc
	//event hooks
	snapshotHooks map[snapshotHookType]map[string]snapshotHookFunc
	eventHooks    map[eventHookType]map[string]eventHookFunc

	//action actions list
	actions []ActionItem
}

//New creates new sim from given profile
func New(p Profile) (*Sim, error) {
	s := &Sim{}

	u := &Enemy{}
	u.Status = make(map[string]int)
	u.Level = p.Enemy.Level
	u.res = p.Enemy.Resist
	u.DamageDetails = make(map[string]map[string]float64)
	u.mod = make(map[string]ResistMod)

	s.Target = u

	s.snapshotHooks = make(map[snapshotHookType]map[string]snapshotHookFunc)
	s.eventHooks = make(map[eventHookType]map[string]eventHookFunc)
	s.tasks = make(map[string]Task)
	s.Status = make(map[string]int)
	s.chars = make(map[string]Character)
	// s.effects = make(map[string]ActionFunc)

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
	if !p.LogShowCaller {
		config.EncoderConfig.CallerKey = ""
	}
	if p.LogFile != "" {
		config.OutputPaths = []string{p.LogFile}
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	s.Log = logger.Sugar()
	zap.ReplaceGlobals(logger)

	dup := make(map[string]bool)
	res := make(map[EleType]int)
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

		s.chars[v.Name] = c

		if v.Name == p.InitialActive {
			s.ActiveChar = p.InitialActive
		}

		if _, ok := dup[v.Name]; ok {
			return nil, fmt.Errorf("duplicated character %v", v.Name)
		}

		dup[v.Name] = true
		s.Target.DamageDetails[v.Name] = make(map[string]float64)
		res[v.Element]++

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
	if s.ActiveChar == "" {
		return nil, fmt.Errorf("invalid active initial character %v", p.InitialActive)
	}
	for _, v := range p.Rotation {
		//make sure char exists
		if _, ok := s.chars[v.CharacterName]; !ok {
			return nil, fmt.Errorf("invalid character %v in rotation list", v.CharacterName)
		}
	}

	s.actions = p.Rotation

	s.addResonance(res)

	//add other hooks
	return s, nil
}

//Run the sim; length in seconds
func (s *Sim) Run(length int) (float64, map[string]map[string]float64) {
	var skip int
	rand.Seed(time.Now().UnixNano())
	//60fps, 60s/min, 2min
	for s.f = 0; s.f < 60*length; s.f++ {
		//tick target and each character
		//target doesn't do anything, just takes punishment, so it won't affect cd
		s.Target.tick(s)
		s.tick()

		s.decrementStatusDuration()
		s.collectEnergyParticles()
		s.executeCharacterTicks()
		s.runEffects()
		s.runTasks()

		if s.swapCD > 0 {
			s.swapCD--
		}

		//if in cooldown, do nothing
		if skip > 0 {
			skip--
			continue
		}

		next, err := s.findNextAction()
		if err != nil {
			s.Log.Infof("[%v] no action found (%v)", s.Frame(), next)
			//skip this round
			continue
		}

		if s.ActiveChar != next.CharacterName {
			//swap
			s.Log.Infof("[%v] swapping from %v to %v", s.Frame(), s.ActiveChar, next.CharacterName)
			s.swapCD = 150
			skip = 20
			s.ActiveChar = next.CharacterName
			continue
		}

		//other wise excute
		skip = s.executeAbilityQueue(next)
	}

	return s.Target.Damage, s.Target.DamageDetails
}

func (s *Sim) decrementStatusDuration() {
	for k, v := range s.Status {
		if v == 0 {
			delete(s.Status, k)
		} else {
			s.Status[k]--
		}
	}
}

func (s *Sim) AddCharMod(c string, key string, val map[StatType]float64) {
	if _, ok := s.chars[c]; ok {
		s.chars[c].AddMod(key, val)
	}
}

func (s *Sim) addResonance(count map[EleType]int) {
	for k, v := range count {
		if v > 2 {
			switch k {
			case Pyro:
				s.AddSnapshotHook(func(ds *Snapshot) bool {
					s.Log.Debugf("\tapplying pyro resonance + 25%% atk")
					ds.Stats[ATKP] += 0.25
					return false
				}, "Pyro Resonance", PostSnapshot)
			case Hydro:
				//heal not implemented yet
			case Cryo:
				s.AddSnapshotHook(func(ds *Snapshot) bool {
					if len(s.Target.Auras) == 0 {
						return false
					}

					if s.Target.Auras[0].Ele == Cryo {
						s.Log.Debugf("\tapplying cryo resonance on cryo target")
						ds.Stats[CR] += .15
					}

					if s.Target.Auras[0].Ele == Frozen {
						s.Log.Debugf("\tapplying cryo resonance on cryo target")
						ds.Stats[CR] += .15
					}
					return false
				}, "Cryo Resonance", PreDamageHook)
			case Electro:
			case Geo:
			case Anemo:
			}
		}
	}
}

//tick
func (s *Sim) tick() {
	if s.swapCD > 0 {
		s.swapCD--
	}
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

type Profile struct {
	Label         string             `yaml:"Label"`
	Enemy         EnemyProfile       `yaml:"Enemy"`
	InitialActive string             `yaml:"InitialActive"`
	Characters    []CharacterProfile `yaml:"Characters"`
	Rotation      []ActionItem       `yaml:"Rotation"`
	LogLevel      string             `yaml:"LogLevel"`
	LogFile       string
	LogShowCaller bool
}

//EnemyProfile ...
type EnemyProfile struct {
	Level  int64               `yaml:"Level"`
	Resist map[EleType]float64 `yaml:"Resist"`
}
