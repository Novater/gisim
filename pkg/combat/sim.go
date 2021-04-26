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
	Log *zap.SugaredLogger
	//exposed fields
	Target           *Enemy
	Status           map[string]int
	ActiveChar       string
	ActiveIndex      int
	CharActiveLength int
	Stam             float64
	Chars            []Character
	SwapCD           int
	Stats            SimStats
	//reaction related
	TargetAura Aura

	Rand        *rand.Rand
	particles   map[int][]Particle
	tasks       map[int][]Task
	F           int
	charPos     map[string]int
	GlobalFlags Flags
	//per tick effects
	effects map[string]func(s *Sim) bool
	//event hooks
	snapshotHooks map[snapshotHookType]map[string]snapshotHookFunc
	eventHooks    map[eventHookType]map[string]eventHookFunc

	//action actions list
	prio        []Action
	actionQueue []ActionItem
}

type Flags struct {
	ChildeActive            bool
	ReactionDidOccur        bool
	ReactionType            ReactionType
	NextAttackMVMult        float64 // melt vape multiplier
	ReactionDamageTriggered bool
}

type SimStats struct {
	AuraUptime           map[EleType]int //uptime in frames
	DamageHist           []float64
	DamageByChar         map[string]map[string]float64
	CharActiveTime       map[string]int
	AbilUsageCountByChar map[string]map[string]int
	ReactionsTriggered   map[ReactionType]int
	SimDuration          int
}

//New creates new sim from given profile
func New(p Profile) (*Sim, error) {
	s := &Sim{}
	s.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	u := &Enemy{}
	u.Status = make(map[string]int)
	u.Level = p.Enemy.Level
	u.res = p.Enemy.Resist
	u.mod = make(map[string]ResistMod)
	s.Target = u
	s.GlobalFlags.NextAttackMVMult = 1

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

	// parser := rotation.New("sim", p.Rotation)
	// rotation, err := parser.Parse()
	// if err != nil {
	// 	return nil, err
	// }
	// p.Rotation
	//double check names
	cust := make(map[string]int)
	for i, v := range p.Rotation {
		if v.Name != "" {
			cust[v.Name] = i
		}
		// log.Println(v.Conditions)
	}
	for _, v := range p.Rotation {
		_, ok := cust[v.Target]
		_, ck := s.charPos[v.Target]
		if !ok && !ck {
			return nil, fmt.Errorf("invalid char in rotation %v", v.Target)
		}
	}

	s.prio = p.Rotation

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
		s.Stats.DamageByChar[v.Base.Name] = make(map[string]float64)
		s.Stats.AbilUsageCountByChar[v.Base.Name] = make(map[string]int)

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
	s.effects = make(map[string]func(s *Sim) bool)
	s.Status = make(map[string]int)
	s.Chars = make([]Character, 0, 4)
	s.particles = make(map[int][]Particle)
	s.charPos = make(map[string]int)
	s.TargetAura = NewNoAura()

	s.Stats.AuraUptime = make(map[EleType]int)
	s.Stats.DamageByChar = make(map[string]map[string]float64)
	s.Stats.CharActiveTime = make(map[string]int)
	s.Stats.AbilUsageCountByChar = make(map[string]map[string]int)
	s.Stats.ReactionsTriggered = make(map[ReactionType]int)

	s.actionQueue = make([]ActionItem, 0, 10)
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
func (s *Sim) RunHPMode(hp float64) (float64, SimStats) {
	var skip int
	rand.Seed(time.Now().UnixNano())
	s.Target.HPMode = true
	s.Target.MaxHP = hp
	s.Target.HP = hp
	//60fps, 60s/min, 2min
	for s.F = 0; s.Target.HP >= 0; s.F++ {
		//tick target and each character
		//target doesn't do anything, just takes punishment, so it won't affect cd
		s.Target.tick(s)

		// s.decrementStatusDuration()
		s.collectEnergyParticles()
		s.executeCharacterTicks()
		s.runEffects()
		s.checkAura()
		s.runTasks()

		//add char active time
		s.Stats.CharActiveTime[s.ActiveChar]++
		s.Stats.AuraUptime[s.TargetAura.E()]++
		s.Stats.SimDuration++
		s.CharActiveLength++

		if s.SwapCD > 0 {
			s.SwapCD--
		}

		//if in cooldown, do nothing
		if skip > 0 {
			skip--
			continue
		}

		//other wise excute
		skip = s.execQueue()
	}

	return s.Target.Damage, s.Stats
}

//Run the sim; length in seconds
func (s *Sim) Run(length int) (float64, SimStats) {
	s.Stats.DamageHist = make([]float64, length*60)
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
		s.checkAura()
		s.runTasks()

		//add char active time
		s.Stats.CharActiveTime[s.ActiveChar]++
		s.Stats.AuraUptime[s.TargetAura.E()]++
		s.CharActiveLength++

		if s.SwapCD > 0 {
			s.SwapCD--
		}

		//damage for this frame
		s.Stats.DamageHist[s.F] = s.Target.Damage

		//recover stam
		if s.Stam < 240 {
			s.Stam += 0.5 //30 per second
			if s.Stam > 240 {
				s.Stam = 240
			}
		}

		//if in cooldown, do nothing
		if skip > 0 {
			skip--
			continue
		}

		//other wise excute
		skip = s.execQueue()
	}

	return s.Target.Damage, s.Stats
}

func (s *Sim) AddCharMod(c string, key string, val map[StatType]float64) {
	pos, ok := s.charPos[c]
	if ok {
		s.Chars[pos].AddMod(key, val)
	}
}

func (s *Sim) addResonance(count map[EleType]int) {
	s.Log.Debugw("checking resonance", "count", count)
	for k, v := range count {
		if v >= 2 {
			switch k {
			case Pyro:
				s.Log.Debugf("activating pyro resonance")
				s.AddSnapshotHook(func(ds *Snapshot) bool {
					s.Log.Debugf("\tapplying pyro resonance + 25%% atk")
					ds.Stats[ATKP] += 0.25
					return false
				}, "Pyro Resonance", PostSnapshot)
			case Hydro:
				//heal not implemented yet
			case Cryo:
				s.Log.Debugf("activating cryo resonance")
				s.AddSnapshotHook(func(ds *Snapshot) bool {
					if s.TargetAura.E() == Cryo {
						s.Log.Debugf("\tapplying cryo resonance on cryo target")
						ds.Stats[CR] += .15
					}

					if s.TargetAura.E() == Frozen {
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

	return strconv.FormatFloat(float64(s.F)/60, 'f', 2, 64) + "s|" + strconv.Itoa(s.F)
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
