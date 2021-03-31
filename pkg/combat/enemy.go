package combat

import (
	"go.uber.org/zap"
)

//Enemy keeps track of the status of one enemy Enemy
type Enemy struct {
	Level int64
	res   map[EleType]float64
	mod   map[string]ResistMod

	//tracking
	Status map[string]int //countdown to how long status last

	//stats
	Damage        float64 //total damage received
	DamageDetails map[string]map[string]float64
}

type ResistMod struct {
	Ele      EleType
	Value    float64
	Duration int
}

func (e *Enemy) Resist(log *zap.SugaredLogger) map[EleType]float64 {
	// log.Debugw("\t\t res calc", "res", e.res, "mods", e.mod)
	r := make(map[EleType]float64)
	for k, v := range e.res {
		r[k] = v
	}
	for k, v := range e.mod {
		log.Debugf("\t\t Resist %v modified by %v; from %v", v.Ele, v.Value, k)
		r[v.Ele] += v.Value
	}
	return r
}

func (e *Enemy) AddResMod(key string, val ResistMod) {
	e.mod[key] = val
}

func (e *Enemy) RemoveResMod(key string) {
	delete(e.mod, key)
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
	//tick down resist mods
	for k, v := range e.mod {
		if v.Duration == 0 {
			s.Log.Debugf("[%v] resist mod %v expired", s.Frame(), k)
			delete(e.mod, k)
		} else {
			v.Duration--
			e.mod[k] = v
		}
	}
}
