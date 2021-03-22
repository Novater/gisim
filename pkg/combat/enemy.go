package combat

import "encoding/json"

//Enemy keeps track of the status of one enemy Enemy
type Enemy struct {
	Level  int64
	Resist map[EleType]float64

	//resist mods
	ResMod map[string]float64

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
	ecTrigger []byte

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
	e.auraTick(s)
}

func (e *Enemy) auraTick(s *Sim) {

	if len(e.Auras) == 0 {
		return
	}
	//tick down gauge, reduction is based on the rate of the aura
	//multipliers are A: 9.5, B:6, C:4.25
	//decay per frame (60fps) = 1 unit / (mult * 60)
	//if only 1 aura we just decay it down
	if len(e.Auras) == 1 {
		v := e.Auras[0]
		v.Gauge -= 1 / (gaugeMul(v.Rate) * 60)
		if v.Gauge > 0 {
			e.Auras[0] = v
		} else {
			e.Auras = nil
		}
		return
	}

	if len(e.Auras) > 2 {
		s.Log.Panicw("We don't know how to handle more than 2 auras!! aborting", "aura", e.Auras)
	}

	//if multiple we need to check for reactions
	//right now that's only EC
	//future may include burning
	var next []aura

	if e.Auras[0].Ele == Electro {
		expired := false
		//do our regular decay business
		for i := 0; i < len(e.Auras); i++ {
			e.Auras[i].Gauge -= 1 / (gaugeMul(e.Auras[i].Rate) * 60)
			if e.Auras[i].Gauge < 0 {
				expired = true
			}
		}
		//check if it's time for another electrocharged tick
		cd, ok := e.Status["electrocharge icd"]
		//either cd is gone (more than 60 frames passed) OR
		//one aura expired and remaining cd is less than 0.5s
		if !ok || (expired && cd < 30) {
			var ds Snapshot
			//this is kinda hacky, we forcefully put the trigger dmg profile in
			//the electro part
			err := json.Unmarshal(e.ecTrigger, &ds)
			if err != nil {
				s.Log.Panicw("tick json error", "err", err)
			}
			ds.ReactionType = ElectroCharged
			damage := s.applyReactionDamage(ds)
			//reduce both auras
			e.Auras[0].Gauge -= 0.5
			e.Auras[1].Gauge -= 0.4
			s.Log.Debugf("[%v] EC reaction tick triggered", s.Frame())
			s.Log.Debugw("\telectrocharged", "damage", damage, "auras", e.Auras)
			e.Status["electrocharge icd"] = 60
		}
		//build next
		for _, v := range e.Auras {
			if v.Gauge > 0 {
				next = append(next, v)
			} else {
				s.Log.Debugf("[%v] aura %v expired", s.Frame(), v.Ele)
			}
		}
		e.Auras = next
		return
	}

	//also freeze is weird - assuming for now that the hydro portion ticks
	//down on the same decay as the cryo portion regardless if it's true or not
	//NEEDS MORE TESTING ON THIS
	if e.Auras[0].Ele == Cryo {
		//we need to use the cryo rate
		rate := e.Auras[0].Rate
		for _, v := range e.Auras {
			//decay first, then delete if < -1
			v.Gauge -= 1 / (gaugeMul(rate) * 60)
			if v.Gauge > 0 {
				next = append(next, v)
			} else {
				s.Log.Debugf("[%v] aura %v expired", s.Frame(), v.Ele)
			}
		}
		//as a sanity check, the length here is either 2, or 0
		l := len(e.Auras)
		if l == 1 {
			s.Log.Warnw("Freeze target: cryo and hydro should have decayed together but did not!!", "aura", e.Auras)
		}
		e.Auras = next
		return
	}

	s.Log.Panicw("Unknown sequence of auras on target!! aborting", "aura", e.Auras)
}
