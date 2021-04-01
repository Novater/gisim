package combat

type Particle struct {
	Source string
	Num    int
	Ele    EleType
	Delay  int
}

func (s *Sim) AddEnergyParticles(source string, num int, ele EleType, delay int) {
	s.particles[s.F+delay] = append(s.particles[s.F+delay], Particle{
		Source: source,
		Num:    num,
		Ele:    ele,
		Delay:  delay,
	})
}

func (s *Sim) collectEnergyParticles() {
	for _, p := range s.particles[s.F] {
		s.distributeParticles(p)
	}
	delete(s.particles, s.F)
}

func (s *Sim) distributeParticles(p Particle) {
	s.Log.Debugf("[%v] Distributing particles from %v; num %v ele %v", s.Frame(), p.Source, p.Num, p.Ele)
	l := len(s.Chars)
	for n, c := range s.Chars {
		a := s.ActiveIndex == n
		c.ReceiveParticle(p, a, l)
	}
}
