package combat

type Particle struct {
	Num   int
	Ele   EleType
	Delay int
}

func (s *Sim) AddEnergyParticles(num int, ele EleType, delay int) {
	s.particles = append(s.particles, Particle{
		Num:   num,
		Ele:   ele,
		Delay: delay,
	})
}

func (s *Sim) collectEnergyParticles() {
	next := make([]Particle, 0, len(s.particles))
	for _, v := range s.particles {
		if v.Delay == 0 {
			s.distributeParticles(v)
		} else {
			v.Delay--
			next = append(next, v)
		}
	}
	s.particles = next
}

func (s *Sim) distributeParticles(p Particle) {
	l := len(s.chars)
	for n, c := range s.chars {
		a := s.ActiveChar == n
		c.ReceiveParticle(p, a, l)
	}
}
