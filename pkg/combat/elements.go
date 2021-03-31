package combat

type Aura interface {
	React(ele EleType, durability float64, s *Sim) Aura
	Tick(s *Sim) bool //remove if true
	Attach(e EleType, durability float64, s *Sim)
	Refresh(dur float64, s *Sim)
}

type Element struct {
	Type          EleType
	MaxDurability float64
	Durability    float64
	Expiry        int //when the aura is gone, use this instead of ticks

	//we can get rid of these
	Base  int
	Units int
}

//react with next element, modifying the existing to = whatever the
//result should be
func (e *Element) React(ele EleType, durability float64, s *Sim) Aura {
	return e
}

func (e *Element) Refresh(durability float64, s *Sim) {
	e.Durability += durability
	if e.Durability > e.MaxDurability {
		e.Durability = e.MaxDurability
	}
	//new expiry is duration * dur/max dur
	n := auraDuration(e.MaxDurability)
	next := int(float64(n) * e.Durability / e.MaxDurability)
	e.Expiry = s.F + next
}

func (e *Element) Tick(s *Sim) bool {
	return e.Expiry < s.F
}

func (e *Element) Attach(ele EleType, durability float64, s *Sim) {
	e.Type = ele
	e.MaxDurability = durability
	e.Durability = durability * 0.8
	e.Expiry = s.F + auraDuration(durability)
}

type NoAura struct {
	*Element
}

func (n *NoAura) React(ele EleType, durability float64, s *Sim) Aura {
	var r Aura
	switch ele {
	case Pyro:
		r = &PyroAura{}
	case Hydro:
		r = &HydroAura{}
	case Cryo:
		r = &CryoAura{}
	case Electro:
		r = &ElectroAura{}
	default:
		return &NoAura{}
	}
	r.Attach(ele, durability, s)
	return r
}

type PyroAura struct {
	*Element
}

func (n *PyroAura) React(ele EleType, durability float64, s *Sim) Aura {
	var r Aura
	switch ele {
	case Pyro:
		n.Refresh(durability, s)
		return n
	case Hydro:
		//hydro on pyro, x2
		r = &HydroAura{}
	case Cryo:
		//cyro on pyro, x1.5
		r = &CryoAura{}
	case Electro:
		//electro on pyro, queue overload
		r = &ElectroAura{}
	default:
		return &NoAura{}
	}
	r.Attach(ele, durability, s)
	return r
}

type HydroAura struct {
	*Element
}

type ElectroAura struct {
	*Element
}

type CryoAura struct {
	*Element
}

type FreezeAura struct {
	*Element
}

type ElectroChargeAura struct {
	*Element
}
