package combat

type Particle struct {
	Source string
	Num    int
	Ele    EleType
	Delay  int
}

func (s *Sim) AddEnergyParticles(source string, num int, ele EleType, delay int) {
	s.particles[Particle{
		Source: source,
		Num:    num,
		Ele:    ele,
		Delay:  delay,
	}] = delay
}

func (s *Sim) collectEnergyParticles() {
	for p, v := range s.particles {
		if v == 0 {
			s.Log.Debugf("[%v] collecting particles %v of %v - %v", s.Frame(), p.Num, p.Ele, p.Source)
			s.distributeParticles(p)
			delete(s.particles, p)
		} else {
			v--
			s.particles[p] = v
		}
	}
}

func (s *Sim) distributeParticles(p Particle) {
	l := len(s.Chars)
	for n, c := range s.Chars {
		a := s.ActiveChar == n
		c.ReceiveParticle(p, a, l)
	}
}
