package combat

func (s *Sim) StatusActive(key string) bool {
	return s.Status[key] >= s.F
}
