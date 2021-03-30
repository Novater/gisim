package combat

import (
	"fmt"
	"math/rand"
	"strconv"
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
	//exposed fields
	Status      map[string]int
	ActiveChar  string
	ActiveIndex int
	Stam        float64
	Chars       []Character
	SwapCD      int
	//overwritable functions
	FindNextAction func(s *Sim) (ActionItem, error)

	Rand      *rand.Rand
	particles map[int][]Particle
	tasks     map[int][]Task
	F         int
	charPos   map[string]int
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
	s.FindNextAction = FindNextAction
	s.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	u := &Enemy{}
	u.Status = make(map[string]int)
	u.Level = p.Enemy.Level
	u.res = p.Enemy.Resist
	u.DamageDetails = make(map[string]map[string]float64)
	u.mod = make(map[string]ResistMod)
	s.Target = u

	s.initMaps()
	s.Stam = 240

	err := s.initLogs(p.LogConfig)
	if err != nil {
		return nil, err
	}
	err = s.initTeam(p)
	if err != nil {
		return nil, err
	}

	for i, v := range p.Rotation {
		//make sure char exists
		pos, ok := s.charPos[v.CharacterName]
		if !ok {
			return nil, fmt.Errorf("invalid character %v in rotation list", v.CharacterName)
		}
		p.Rotation[i].index = pos
	}
	s.actions = p.Rotation

	//add other hooks
	return s, nil
}

func (s *Sim) initTeam(p Profile) error {
	dup := make(map[string]bool)
	res := make(map[EleType]int)

	for i, v := range p.Characters {
		//call new char function
		f, ok := charMap[v.Base.Name]
		if !ok {
			return fmt.Errorf("invalid character: %v", v.Base.Name)
		}
		c, err := f(s, v)
		if err != nil {
			return err
		}

		s.Chars = append(s.Chars, c)
		s.charPos[v.Base.Name] = i

		if v.Base.Name == p.InitialActive {
			s.ActiveChar = p.InitialActive
			s.ActiveIndex = i
		}

		if _, ok := dup[v.Base.Name]; ok {
			return fmt.Errorf("duplicated character %v", v.Base.Name)
		}
		dup[v.Base.Name] = true
		s.Target.DamageDetails[v.Base.Name] = make(map[string]float64)

		//initialize weapon
		wf, ok := weaponMap[v.Weapon.Name]
		if !ok {
			return fmt.Errorf("unrecognized weapon %v for character %v", v.Weapon.Name, v.Base.Name)
		}
		wf(c, s, v.Weapon.Refine)

		//check set bonus
		sb := make(map[string]int)
		for _, a := range v.ArtifactsConfig {
			sb[a.Set]++
		}

		//add set bonus
		for key, count := range sb {
			f, ok := setMap[key]
			if ok {
				f(c, s, count)
			} else {
				s.Log.Warnf("character %v has unrecognized set %v", v.Base.Name, key)
			}
		}

		//track resonance
		res[v.Base.Element]++

	}
	if s.ActiveChar == "" {
		return fmt.Errorf("invalid active initial character %v", p.InitialActive)
	}

	s.addResonance(res)
	return nil
}

func (s *Sim) initMaps() {
	s.snapshotHooks = make(map[snapshotHookType]map[string]snapshotHookFunc)
	s.eventHooks = make(map[eventHookType]map[string]eventHookFunc)
	s.tasks = make(map[int][]Task)
	s.Status = make(map[string]int)
	s.Chars = make([]Character, 0, 4)
	s.particles = make(map[int][]Particle)
	s.charPos = make(map[string]int)
}

func (s *Sim) initLogs(p LogConfig) error {
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
		return err
	}
	s.Log = logger.Sugar()
	zap.ReplaceGlobals(logger)
	return nil
}

//Run the sim; length in seconds
func (s *Sim) Run(length int) (float64, map[string]map[string]float64, []float64) {
	graph := make([]float64, length*60)
	var skip int
	rand.Seed(time.Now().UnixNano())
	//60fps, 60s/min, 2min
	for s.F = 0; s.F < 60*length; s.F++ {
		//tick target and each character
		//target doesn't do anything, just takes punishment, so it won't affect cd
		s.Target.tick(s)

		// s.decrementStatusDuration()
		s.collectEnergyParticles()
		s.executeCharacterTicks()
		s.runEffects()
		s.runTasks()

		if s.SwapCD > 0 {
			s.SwapCD--
		}

		//damage for this frame
		graph[s.F] = s.Target.Damage

		//if in cooldown, do nothing
		if skip > 0 {
			skip--
			continue
		}

		next, err := s.FindNextAction(s)
		s.Log.Infof("[%v] off skip, next action: %v swap cd %v", s.Frame(), next, s.SwapCD)
		if err != nil {
			s.Log.Infof("[%v] no action found (%v)", s.Frame(), next)
			//skip this round
			continue
		}

		if s.ActiveChar != next.CharacterName {
			//swap
			s.Log.Infof("[%v] swapping from %v to %v; cd: %v action %v", s.Frame(), s.ActiveChar, next.CharacterName, s.SwapCD, next)
			s.SwapCD = 150
			skip = 20
			s.ActiveChar = next.CharacterName
			s.ActiveIndex = next.index
			continue
		}

		//other wise excute
		skip = s.executeAbilityQueue(next)
		s.Log.Infof("[%v] action executed; skip %v swap cd %v", s.Frame(), skip, s.SwapCD)
	}

	return s.Target.Damage, s.Target.DamageDetails, graph
}

func (s *Sim) AddCharMod(c string, key string, val map[StatType]float64) {
	pos, ok := s.charPos[c]
	if ok {
		s.Chars[pos].AddMod(key, val)
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

func (s *Sim) Frame() string {
	return strconv.Itoa(int(1000*float64(s.F)/60)) + "ms|" + strconv.Itoa(s.F)
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
