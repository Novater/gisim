package monte

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/srliao/gisim/pkg/combat"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	//characters
	_ "github.com/srliao/gisim/internal/character/bennett"
	_ "github.com/srliao/gisim/internal/character/eula"
	_ "github.com/srliao/gisim/internal/character/fischl"
	_ "github.com/srliao/gisim/internal/character/ganyu"
	_ "github.com/srliao/gisim/internal/character/xiangling"
	_ "github.com/srliao/gisim/internal/character/xingqiu"

	//weapons
	_ "github.com/srliao/gisim/internal/weapon/bow/favoniuswarbow"
	_ "github.com/srliao/gisim/internal/weapon/bow/prototypecrescent"
	_ "github.com/srliao/gisim/internal/weapon/claymore/skyward"
	_ "github.com/srliao/gisim/internal/weapon/spear/blacktassel"
	_ "github.com/srliao/gisim/internal/weapon/spear/skywardspine"
	_ "github.com/srliao/gisim/internal/weapon/sword/festeringdesire"
	_ "github.com/srliao/gisim/internal/weapon/sword/sacrificialsword"

	//sets
	_ "github.com/srliao/gisim/internal/artifact/blizzard"
	_ "github.com/srliao/gisim/internal/artifact/bloodstained"
	_ "github.com/srliao/gisim/internal/artifact/crimson"
	_ "github.com/srliao/gisim/internal/artifact/gladiator"
	_ "github.com/srliao/gisim/internal/artifact/noblesse"
	_ "github.com/srliao/gisim/internal/artifact/wanderer"
)

type Simulator struct {
	Log       *zap.SugaredLogger
	p         combat.Profile
	charIndex int
}

type Config struct {
}

func New(p combat.Profile, charIndex int) (*Simulator, error) {
	s := &Simulator{}

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

	s.p = p
	s.charIndex = charIndex

	return s, nil
}

type SimResult struct {
	Hist     []float64
	BinStart int64
	Min      float64
	Max      float64
	Mean     float64
	SD       float64
}

func (s *Simulator) SimDmgDist(n, b, w int64) SimResult {
	r := SimResult{}

	//calculate the damage distribution
	s.Log.Debugw("starting dmg sim", "n", n, "b", b, "w", w)

	var progress, sum, ss float64
	var data []float64
	r.Min = math.MaxFloat64
	r.Max = -1

	count := n

	resp := make(chan float64, n)
	req := make(chan bool)
	done := make(chan bool)
	for i := 0; i < int(w); i++ {
		go s.worker(resp, req, done)
	}

	//use a go routine to send out a job whenever a worker is done
	go func() {
		var wip int64
		for wip < n {
			//try sending a job to req chan while wip < cfg.NumSim
			req <- true
			wip++
		}
	}()

	fmt.Print("\tProgress: 0")

	for count > 0 {
		//process results received
		val := <-resp
		count--

		//add the avg, rest doesn't really make sense
		data = append(data, val)
		sum += val
		if val < r.Min {
			r.Min = val
		}
		if val > r.Max {
			r.Max = val
		}

		if (1 - float64(count)/float64(n)) > (progress + 0.01) {
			progress = (1 - float64(count)/float64(n))
			fmt.Printf(".%.0f", 100*progress)
		}
	}
	fmt.Print("...100%\n")

	close(done)

	r.Mean = sum / float64(n)
	r.BinStart = int64(r.Min/float64(b)) * b
	binMax := (int64(r.Max/float64(b)) + 1.0) * b
	numBin := ((binMax - r.BinStart) / b) + 1

	r.Hist = make([]float64, numBin)

	for _, v := range data {
		ss += (v - r.Mean) * (v - r.Mean)
		steps := int64((v - float64(r.BinStart)) / float64(b))
		r.Hist[steps]++
	}

	r.SD = math.Sqrt(ss / float64(n))

	return r
}

func (s *Simulator) worker(resp chan float64, req chan bool, done chan bool) {
	seed := time.Now().UnixNano()
	rand := rand.New(rand.NewSource(seed))
	//we need to make a copy of the profile, at least the char part

	prof := s.p
	artifacts := prof.Characters[s.charIndex].Artifacts

	for {
		select {
		case <-req:
			//generate a set of artifacts

			flower := s.RandArtifact(artifacts[combat.Flower], rand)
			feather := s.RandArtifact(artifacts[combat.Feather], rand)
			sands := s.RandArtifact(artifacts[combat.Sands], rand)
			goblet := s.RandArtifact(artifacts[combat.Goblet], rand)
			circlet := s.RandArtifact(artifacts[combat.Circlet], rand)

			//make a new map
			prof.Characters[s.charIndex].Artifacts = make(map[combat.Slot]combat.Artifact)
			c := prof.Characters[s.charIndex]

			c.Artifacts[combat.Flower] = flower
			c.Artifacts[combat.Feather] = feather
			c.Artifacts[combat.Sands] = sands
			c.Artifacts[combat.Goblet] = goblet
			c.Artifacts[combat.Circlet] = circlet

			prof.LogFile = ""
			prof.LogLevel = "error"

			sim, err := combat.New(prof)

			if err != nil {
				log.Panic(err)
			}

			//calculate dmg
			t := 5000
			r, _, _ := sim.Run(t)

			resp <- (r / float64(t))
		case <-done:
			return
		}

	}

}

var subIndex = []string{
	"HP",
	"ATK",
	"DEF",
	"HP%",
	"ATK%",
	"DEF%",
	"ER",
	"EM",
	"CR",
	"CD",
}

var weights = []float64{
	150,
	150,
	150,
	100,
	100,
	100,
	100,
	100,
	75,
	75,
}

var tiers = [][]float64{
	{209, 239, 269, 299},         //hp
	{14, 16, 18, 19},             //atk
	{16, 19, 21, 23},             //def
	{0.041, 0.047, 0.053, 0.058}, //hp%
	{0.041, 0.047, 0.053, 0.058}, //atk%
	{0.051, 0.058, 0.066, 0.073}, //def%
	{0.045, 0.052, 0.058, 0.065}, //er
	{16, 19, 21, 23},             //em
	{0.027, 0.031, 0.035, 0.039}, //cr
	{0.054, 0.062, 0.07, 0.078},  //cd
}

//RandArtifact generates one random artifact with specified main stat
func (s *Simulator) RandArtifact(a combat.Artifact, rand *rand.Rand) combat.Artifact {

	/**
		For example, most Artifacts rewards use the sub-drop mechanic. The order for rolling is:

			The rarity of Artifact <- this we assume 5*
			The number of initial Sub Stat <- roll this first
			The Artifact set <- this we assume constant
			The type of Artifact (flower, feather, etc) <- this we assume constant

	**/

	var r combat.Artifact

	r.Type = a.Type
	r.MainStat = a.MainStat
	r.Level = a.Level

	//how many substats
	p := rand.Float64()
	var lines = 3
	if p <= 306.0/1415.0 {
		lines = 4
	}

	//if artifact lvl is less than 4 AND lines =3, then we only want to roll 3 substats
	n := 4
	if r.Level < 4 && lines < 4 {
		n = 3
	}

	//make a copy of prob
	prb := make([]float64, len(subIndex))
	keys := make([]combat.StatType, len(subIndex))

	for i, v := range subIndex {
		w := weights[i]
		keys[i] = combat.StatType(v)
		if combat.StatType(v) == r.MainStat.Type {
			w = 0
		}
		prb[i] = w
	}

	for i := 0; i < n; i++ {
		var sumWeights float64
		for _, v := range prb {
			sumWeights += v
		}

		found := -1
		//pick a number between 0 and sumweights
		pick := rand.Float64() * sumWeights
		for i, v := range prb {
			if pick < v && found == -1 {
				found = i
			}
			pick -= v
		}
		if found == -1 {
			log.Println("sum weights ", sumWeights)
			log.Println("prb ", prb)
			log.Panic("unexpected - no random stat generated")
		}
		// log.Println("found at ", found)
		t := keys[found]
		//set prb for this stat to 0 for next iteration
		prb[found] = 0

		tier := rand.Intn(4)
		val := tiers[found][tier]
		r.Substat = append(r.Substat, combat.Stat{
			Type:  t,
			Value: val,
		})
	}

	//check how many upgrades to do
	up := r.Level / 4

	//if we started w 3 lines, then subtract one from # of upgrades
	if lines == 3 {
		up--
	}

	if len(r.Substat) != 4 {
		s.Log.Debugw("invalid artifact, less than 4 lines", "a", r)
		log.Panic("invalid artifact")
	}

	//need to index the substats
	index := make([]int, 4)
	for i, v := range r.Substat {
		n = -1
	inner:
		for i, k := range subIndex {
			if combat.StatType(k) == v.Type {
				n = i
				break inner
			}
		}
		if n == -1 {
			s.Log.Debugw("invalid stat, can't find in subindex", "a", r, "sub", subIndex, "substat", v)
			log.Panic("error indexing")
		}
		index[i] = n
	}

	//do more rolls, +4/+8/+12/+16/+20
	for i := 0; i < int(up); i++ {
		pick := rand.Intn(4)
		tier := rand.Intn(4)
		// r.Substat[pick].Value += s.SubTier[tier][r.Substat[pick].Type]
		r.Substat[pick].Value += tiers[index[pick]][tier]
	}

	return r

}
