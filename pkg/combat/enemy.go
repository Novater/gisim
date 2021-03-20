package combat

//Enemy keeps track of the status of one enemy Enemy
type Enemy struct {
	Level  int64
	Resist map[EleType]float64

	//resist mods
	ResMod map[string]float64

	//tracking
	Auras  map[EleType]aura
	Status map[string]int //countdown to how long status last

	//stats
	Damage float64 //total damage received
}

func (e *Enemy) tick(s *Sim) {
	//tick down buffs and debuffs
	for k, v := range e.Status {
		if v == 0 {
			delete(e.Status, k)
		} else {
			e.Status[k]--
		}
	}
	//tick down aura
	for k, v := range e.Auras {
		if v.duration == 0 {
			s.Log.Debugf("[%v] aura %v expired", s.Frame(), k)
			delete(e.Auras, k)
		} else {
			a := e.Auras[k]
			a.duration--
			e.Auras[k] = a
		}
	}
}
