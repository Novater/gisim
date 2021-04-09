package combat

func (s *Sim) StatusActive(key string) bool {
	f, ok := s.Status[key]
	if !ok {
		return false
	}
	return f >= s.F
}
