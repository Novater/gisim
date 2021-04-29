package combat

type NewSetFunc func(c Character, s *Sim, count int)

func RegisterSetFunc(name string, f NewSetFunc) {
	mu.Lock()
	defer mu.Unlock()
	if _, dup := setMap[name]; dup {
		panic("combat: RegisterSetBonus called twice for character " + name)
	}
	setMap[name] = f
}

// type ArtifactSim struct {
// 	Log       *zap.SugaredLogger
// 	p         Profile
// 	charIndex int
// }

// func NewArtifactSim(p Profile, charIndex int) (*ArtifactSim, error) {
// 	a := &ArtifactSim{}

// 	config := zap.NewDevelopmentConfig()
// 	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
// 	switch p.LogConfig.LogLevel {
// 	case "debug":
// 		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
// 	case "info":
// 		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
// 	case "warn":
// 		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
// 	case "error":
// 		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
// 	}
// 	config.EncoderConfig.TimeKey = ""
// 	config.EncoderConfig.StacktraceKey = ""
// 	if !p.LogConfig.LogShowCaller {
// 		config.EncoderConfig.CallerKey = ""
// 	}
// 	if p.LogConfig.LogFile != "" {
// 		config.OutputPaths = []string{p.LogConfig.LogFile}
// 	}

// 	logger, err := config.Build()
// 	if err != nil {
// 		return nil, err
// 	}
// 	a.Log = logger.Sugar()
// 	a.p = p
// 	a.charIndex = charIndex

// 	return a, nil
// }

// type ArtifactSimResult struct {
// 	Hist     []float64
// 	BinStart int
// 	Min      float64
// 	Max      float64
// 	Mean     float64
// 	SD       float64
// }

// func (a *ArtifactSim) RunDmgSim(b, n, w int) ArtifactSimResult {
// 	r := ArtifactSimResult{}

// 	//calculate the damage distribution
// 	a.Log.Debugw("starting dmg sim", "n", n, "b", b, "w", w)

// 	var progress, sum, ss float64
// 	var data []float64
// 	r.Min = math.MaxFloat64
// 	r.Max = -1

// 	count := n

// 	resp := make(chan float64, n)
// 	req := make(chan bool)
// 	done := make(chan bool)

// 	byt, _ := json.Marshal(a.p)

// 	for i := 0; i < int(w); i++ {
// 		var p Profile
// 		json.Unmarshal(byt, &p)
// 		go a.worker(p, resp, req, done)
// 	}

// 	//use a go routine to send out a job whenever a worker is done
// 	go func() {
// 		var wip int
// 		for wip < n {
// 			//try sending a job to req chan while wip < cfg.NumSim
// 			req <- true
// 			wip++
// 		}
// 	}()

// 	fmt.Print("\tProgress: 0")

// 	for count > 0 {
// 		//process results received
// 		val := <-resp
// 		count--

// 		//add the avg, rest doesn't really make sense
// 		data = append(data, val)
// 		sum += val
// 		if val < r.Min {
// 			r.Min = val
// 		}
// 		if val > r.Max {
// 			r.Max = val
// 		}

// 		if (1 - float64(count)/float64(n)) > (progress + 0.01) {
// 			progress = (1 - float64(count)/float64(n))
// 			fmt.Printf(".%.0f", 100*progress)
// 		}
// 	}
// 	fmt.Print("...done\n")

// 	close(done)

// 	r.Mean = sum / float64(n)
// 	r.BinStart = int(r.Min/float64(b)) * b
// 	binMax := (int(r.Max/float64(b)) + 1.0) * b
// 	numBin := ((binMax - r.BinStart) / b) + 1

// 	r.Hist = make([]float64, numBin)

// 	for _, v := range data {
// 		ss += (v - r.Mean) * (v - r.Mean)
// 		steps := int64((v - float64(r.BinStart)) / float64(b))
// 		r.Hist[steps]++
// 	}

// 	r.SD = math.Sqrt(ss / float64(n))

// 	return r
// }

// func (a *ArtifactSim) worker(p Profile, resp chan float64, req chan bool, done chan bool) {
// 	seed := time.Now().UnixNano()
// 	rand := rand.New(rand.NewSource(seed))
// 	//p is a clone unique to this worker
// 	p.LogConfig.LogFile = ""
// 	p.LogConfig.LogLevel = "error"

// 	ms := make([]StatType, 5)
// 	msv := make([]float64, 5)
// 	lvls := make([]int, 5)
// 	// log.Println(p.Characters[a.charIndex])

// 	if len(p.Characters[a.charIndex].ArtifactsConfig) > 5 {
// 		log.Panic("too many artifacts")
// 	}

// 	count := 0
// 	for _, art := range p.Characters[a.charIndex].ArtifactsConfig {
// 		//grab the first key
// 		key := ""
// 		var value float64
// 		for m, v := range art.Main {
// 			key = m
// 			value = v
// 			break
// 		}
// 		if key == "" {
// 			log.Panicf("no main stat: %v", art)
// 		}
// 		//match the key
// 		found := -1
// 		for i, v := range StatTypeString {
// 			if key == v {
// 				found = i
// 				break
// 			}
// 		}
// 		if found == -1 {
// 			log.Panicf("main stat not found: %v", art)
// 		}
// 		ms[count] = StatType(found)
// 		msv[count] = value
// 	}

// 	for {
// 		select {
// 		case <-req:
// 			//generate a set of artifacts

// 			sim, err := New(p)
// 			if err != nil {
// 				log.Panic(err)
// 			}

// 			stats := make([]float64, len(StatTypeString))

// 			for i := range ms {
// 				x := randArtifact(rand, ms[i], lvls[i])
// 				for i, v := range x {
// 					stats[i] += v
// 				}
// 				stats[ms[i]] += msv[i]
// 			}

// 			// log.Println(stats)

// 			sim.Chars[a.charIndex].UnsafeSetStats(stats)

// 			//calculate dmg
// 			t := 5000
// 			r, _ := sim.Run(t)

// 			resp <- (r / float64(t))
// 		case <-done:
// 			return
// 		}

// 	}

// }

// //generate set of random artifact with stats
// //return main stat type and the total stats
// func randArtifact(rand *rand.Rand, main StatType, lvl int) []float64 {
// 	stats := make([]float64, len(StatTypeString))

// 	//how many substats
// 	p := rand.Float64()
// 	var lines = 3
// 	if p <= 306.0/1415.0 {
// 		lines = 4
// 	}
// 	//if artifact lvl is less than 4 AND lines =3, then we only want to roll 3 substats
// 	n := 4
// 	if lvl < 4 && lines < 4 {
// 		n = 3
// 	}
// 	//make a copy of prob
// 	prb := make([]float64, len(subIndex))

// 	for i, v := range subIndex {
// 		w := weights[i]
// 		if v == main {
// 			w = 0
// 		}
// 		prb[i] = w
// 	}

// 	subType := []StatType{-1, -1, -1, -1}
// 	subLine := []int{-1, -1, -1, -1}

// 	for i := 0; i < n; i++ {
// 		var sumWeights float64
// 		for _, v := range prb {
// 			sumWeights += v
// 		}

// 		found := -1
// 		//pick a number between 0 and sumweights
// 		pick := rand.Float64() * sumWeights
// 		for i, v := range prb {
// 			if pick < v && found == -1 {
// 				found = i
// 			}
// 			pick -= v
// 		}
// 		if found == -1 {
// 			log.Println("sum weights ", sumWeights)
// 			log.Println("prb ", prb)
// 			log.Panic("unexpected - no random stat generated")
// 		}
// 		// log.Println("found at ", found)
// 		subType[i] = subIndex[found]
// 		subLine[i] = found
// 		//set prb for this stat to 0 for next iteration
// 		prb[found] = 0

// 		tier := rand.Intn(4)

// 		stats[subIndex[found]] += tiers[found][tier]
// 	}

// 	//check how many upgrades to do
// 	up := lvl / 4

// 	//if we started w 3 lines, then subtract one from # of upgrades
// 	if lines == 3 {
// 		up--
// 	}

// 	//do more rolls, +4/+8/+12/+16/+20
// 	for i := 0; i < int(up); i++ {
// 		pick := rand.Intn(4)
// 		tier := rand.Intn(4)

// 		stats[subType[pick]] += tiers[subLine[pick]][tier]
// 	}

// 	return stats
// }

// var subIndex = []StatType{
// 	HP,
// 	ATK,
// 	DEF,
// 	HPP,
// 	ATKP,
// 	DEFP,
// 	ER,
// 	EM,
// 	CR,
// 	CD,
// }

// var weights = []float64{
// 	150,
// 	150,
// 	150,
// 	100,
// 	100,
// 	100,
// 	100,
// 	100,
// 	75,
// 	75,
// }

// var tiers = [][]float64{
// 	{209, 239, 269, 299},         //hp
// 	{14, 16, 18, 19},             //atk
// 	{16, 19, 21, 23},             //def
// 	{0.041, 0.047, 0.053, 0.058}, //hp%
// 	{0.041, 0.047, 0.053, 0.058}, //atk%
// 	{0.051, 0.058, 0.066, 0.073}, //def%
// 	{0.045, 0.052, 0.058, 0.065}, //er
// 	{16, 19, 21, 23},             //em
// 	{0.027, 0.031, 0.035, 0.039}, //cr
// 	{0.054, 0.062, 0.07, 0.078},  //cd
// }
