package combat

type Particle struct {
	Source string
	Num    int
	Ele    EleType
	Delay  int
}

func (s *Sim) AddEnergyParticles(source string, num int, ele EleType, delay int) {
	s.particles[s.f+delay] = append(s.particles[s.f+delay], Particle{
		Source: source,
		Num:    num,
		Ele:    ele,
		Delay:  delay,
	})
}

func (s *Sim) collectEnergyParticles() {

	for _, p := range s.particles[s.f] {
		s.distributeParticles(p)
	}

	delete(s.particles, s.f)
}

func (s *Sim) distributeParticles(p Particle) {
	l := len(s.Chars)
	for n, c := range s.Chars {
		a := s.ActiveChar == n
		c.ReceiveParticle(p, a, l)
	}
}
