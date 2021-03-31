package combat

import (
	"go.uber.org/zap"
)

//Enemy keeps track of the status of one enemy Enemy
type Enemy struct {
	Level int64
	res   map[EleType]float64
	mod   map[string]ResistMod

	//auras should be stored in an array
	//there seems to be a priority system on what gets stored first; don't know how it works
	//for now we're only concerned about EC and Freeze; so we'll just hard code though; EC is electro,hydro; Freeze is cryo,hydro
	//in EC case, any additional reaction applied will first react with electro, then hydro
	//	not sure if this is only applicable if first reaction is transformative i.e. overload or superconduct;
	//	^ may make sense b/c additional application of hydro will prob trigger additional reaction? need to test this somehow??
	//  but the code would be the same, i.e hydro gets applied to electro (does nothing), then added to more hydro
	//in frozen's case we just have to hard code that any element will only react with cryo and hydro just dies off why cryo dies off
	//unless we shatter in which case cryo gets removed
	Auras []aura
	//tracking
	Status map[string]int //countdown to how long status last
	//ec store
	IsFrozen bool

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
