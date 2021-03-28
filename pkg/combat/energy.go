package combat

type Particle struct {
	Source string
	Num    int
	Ele    EleType
	Delay  int
}

func (s *Sim) AddEnergyParticles(source string, num int, ele EleType, delay int) {
	s.particles = append(s.particles, Particle{
		Source: source,
		Num:    num,
		Ele:    ele,
		Delay:  delay,
	})
}

func (s *Sim) collectEnergyParticles() {
	next := make([]Particle, 0, len(s.particles))
	for _, v := range s.particles {
		if v.Delay == 0 {
			s.Log.Debugf("[%v] collecting particles %v of %v - %v", s.Frame(), v.Num, v.Ele, v.Source)
			s.distributeParticles(v)
		} else {
			v.Delay--
			next = append(next, v)
		}
	}
	s.particles = next
}

func (s *Sim) distributeParticles(p Particle) {
	l := len(s.Chars)
	for n, c := range s.Chars {
		a := s.ActiveChar == n
		c.ReceiveParticle(p, a, l)
	}
}
