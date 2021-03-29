package combat

func (s *Sim) decrementStatusDuration() {
	for k, v := range s.Status {
		if v == 0 {
			delete(s.Status, k)
		} else {
			s.Status[k]--
		}
	}
}

func (s *Sim) StatusActive(key string) bool {
	return s.Status[key] >= s.F
}
