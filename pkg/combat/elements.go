package combat

type Aura interface {
	React(ds Snapshot, s *Sim) Aura
	Tick(s *Sim) bool //remove if true
	Attach(e EleType, durability float64, f int)
	Refresh(dur float64, s *Sim)
	E() EleType
}

func (s *Sim) checkAura() {
	if s.TargetAura.Tick(s) {
		s.TargetAura = NewNoAura()
	}
}

const (
	WeakDurability   = 25
	MedDurability    = 50
	StrongDurability = 100
)

type Element struct {
	Type          EleType
	MaxDurability float64
	Durability    float64
	Expiry        int //when the aura is gone, use this instead of ticks
}

//react with next element, modifying the existing to = whatever the
//result should be
func (e *Element) React(ds Snapshot, s *Sim) Aura {
	return e
}

func (e *Element) E() EleType {
	return e.Type
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

func (e *Element) Reduce(durability float64, s *Sim) {
	e.Durability -= durability
	if e.Durability < 0 {
		e.Durability = 0
		e.Expiry = s.F
		return
	}
	//recalc duration
	n := auraDuration(e.MaxDurability)
	next := int(float64(n) * e.Durability / e.MaxDurability)
	e.Expiry = s.F + next
}

func (e *Element) Tick(s *Sim) bool {
	return e.Expiry < s.F
}

func (e *Element) Attach(ele EleType, durability float64, f int) {
	e.Type = ele
	e.MaxDurability = durability
	e.Durability = durability * 0.8
	e.Expiry = f + auraDuration(durability)
}

func auraDuration(d float64) int {
	//calculate duration
	return int(6*d + 420)
}

type NoAura struct {
	*Element
}

func (n *NoAura) React(ds Snapshot, s *Sim) Aura {
	if ds.Durability == 0 {
		return n
	}
	var r Aura
	switch ds.Element {
	case Pyro:
		r = NewPyro()
	case Hydro:
		r = NewHydro()
	case Cryo:
		r = NewCryo()
	case Electro:
		r = NewElectro()
	default:
		return NewNoAura()
	}
	r.Attach(ds.Element, ds.Durability, s.F)
	return r
}

type PyroAura struct {
	*Element
}

func (p *PyroAura) React(ds Snapshot, s *Sim) Aura {
	if ds.Durability == 0 {
		return p
	}
	switch ds.Element {
	case Pyro:
		p.Refresh(ds.Durability, s)
		return p
	case Hydro:
		//hydro on pyro, x2, strong so x2
		p.Reduce(2*ds.Durability, s)
		s.GlobalFlags.NextAttackMVMult = 2
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = Vaporize
		//fire gone
		if p.Durability == 0 {
			return NewNoAura()
		}
		return p
	case Cryo:
		//cyro on pyro, x1.5, weak so only .5 applied
		p.Reduce(0.5*ds.Durability, s)
		//fire gone
		if p.Durability == 0 {
			return NewNoAura()
		}
		s.GlobalFlags.NextAttackMVMult = 1.5
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = Melt
		return p
	case Electro:
		//electro on pyro, queue overload
		//reduction in durability is 1:1, no tax on the
		p.Reduce(ds.Durability, s)
		s.GlobalFlags.NextAttackOverloadTriggered = true
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = Overload
		//fire gone
		if p.Durability == 0 {
			return NewNoAura()
		}
		return p
	default:
		return NewNoAura()
	}
}

type HydroAura struct {
	*Element
}

func (h *HydroAura) React(ds Snapshot, s *Sim) Aura {
	if ds.Durability == 0 {
		return h
	}
	switch ds.Element {
	case Pyro:
		//hydro on pyro, x1.5
		h.Reduce(0.5*ds.Durability, s)
		s.GlobalFlags.NextAttackMVMult = 1.5
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = Vaporize
		//fire gone
		if h.Durability == 0 {
			return NewNoAura()
		}
		return h
	case Hydro:
		h.Refresh(ds.Durability, s)
		return h
	case Cryo:
		//freeze??
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = Freeze
		return h
	case Electro:
		//ec??
		e := NewElectro()
		e.Attach(ds.Element, ds.Durability, s.F)
		ec := NewEC()
		ec.Init(e, h, ds, s)
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = ElectroCharged
		return ec
	default:
		return NewNoAura()
	}
}

type ElectroAura struct {
	*Element
}

func (e *ElectroAura) React(ds Snapshot, s *Sim) Aura {
	if ds.Durability == 0 {
		return e
	}
	switch ds.Element {
	case Pyro:
		//electro on pyro, queue overload
		//reduction in durability is 1:1, no tax on the
		e.Reduce(ds.Durability, s)
		s.GlobalFlags.NextAttackOverloadTriggered = true
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = Overload
		//fire gone
		if e.Durability == 0 {
			return NewNoAura()
		}
		return e
	case Hydro:
		//ec??
		h := NewHydro()
		h.Attach(ds.Element, ds.Durability, s.F)
		ec := NewEC()
		ec.Init(e, h, ds, s)
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = ElectroCharged
		return ec
	case Cryo:
		//superconduct, 1:1
		e.Reduce(ds.Durability, s)
		s.GlobalFlags.NextAttackSuperconductTriggered = true
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = Superconduct
		if e.Durability == 0 {
			return NewNoAura()
		}
		return e
	case Electro:
		//ec??
		e.Refresh(ds.Durability, s)
		return e
	default:
		return NewNoAura()
	}
}

type CryoAura struct {
	*Element
}

func (c *CryoAura) React(ds Snapshot, s *Sim) Aura {
	if ds.Durability == 0 {
		return c
	}
	switch ds.Element {
	case Pyro:
		//pyro on cryo, x2
		c.Reduce(2*ds.Durability, s)
		s.GlobalFlags.NextAttackMVMult = 2
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = Melt
		//fire gone
		if c.Durability == 0 {
			return NewNoAura()
		}
		return c
	case Hydro:
		//freeze?
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = Freeze
		return c
	case Cryo:
		c.Refresh(ds.Durability, s)
		return c
	case Electro:
		//electro on cryo, queue on superconduct
		//reduction in durability is 1:1, no tax on the
		c.Reduce(ds.Durability, s)
		s.GlobalFlags.NextAttackSuperconductTriggered = true
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = Overload
		if c.Durability == 0 {
			return NewNoAura()
		}
		return c
	default:
		return NewNoAura()
	}
}

type FreezeAura struct {
	*Element
}

func (f *FreezeAura) Init() {

}

func (f *FreezeAura) React(ds Snapshot, s *Sim) Aura {
	if ds.Durability == 0 && !ds.IsHeavyAttack {
		return f
	}
	switch ds.Element {
	case Pyro:
		//just melt
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = Melt
		return f
	case Hydro:
		//extend??
		return f
	case Cryo:
		//extend??
		return f
	case Electro:
		//just superconuct
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = Superconduct
		return f
	default:
		return NewNoAura()
	}
}

type ElectroChargeAura struct {
	*Element
	Hydro    *HydroAura
	Electro  *ElectroAura
	NextTick int
	Snap     Snapshot
}

func (e *ElectroChargeAura) Init(electro *ElectroAura, hydro *HydroAura, ds Snapshot, s *Sim) {
	e.Type = EC
	e.Hydro = hydro
	e.Electro = electro
	e.Snap = ds.Clone()
	e.NextTick = s.F
}

func (e *ElectroChargeAura) Tick(s *Sim) bool {
	//if nexttick is <= s.F, we want to trigger damage
	//and reduce durability of both hydro/electro
	if e.NextTick <= s.F {
		//queue up damage here
		s.applyReactionDamage(e.Snap, ElectroCharged)
		//reduce durability
		e.Hydro.Reduce(10, s)
		e.Electro.Reduce(10, s)
	}

	//we then check for expiry; note that if it's not time
	//for next ticks yet but either of these are ready to
	//expire due to time, then this here will take care of it
	//and prevent any further reactions
	eExp := e.Electro.Expiry <= s.F
	hExp := e.Electro.Expiry <= s.F

	//if both expired, only as a result of the ticks
	//then return true so sim can change this to NoAura
	if eExp && hExp {
		return true
	}
	//if only one expired, then we have to manually override
	//the sim's element to just the other one
	if eExp {
		s.TargetAura = e.Hydro
		return false
	}
	if hExp {
		s.TargetAura = e.Electro
		return false
	}

	//if nothing expired, then queue up next tick
	e.NextTick = s.F + 60
	return false
}

func (e *ElectroChargeAura) React(ds Snapshot, s *Sim) Aura {
	if ds.Durability == 0 {
		return e
	}
	switch ds.Element {
	case Pyro:
		//overload + vaporize
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.NextAttackOverloadTriggered = true
		s.GlobalFlags.NextAttackMVMult = 1.5
		s.GlobalFlags.ReactionType = Overload //this here will trigger our super vape
		return e
	case Hydro:
		//extend??
		return e
	case Cryo:
		//just superconduct
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.NextAttackSuperconductTriggered = true
		s.GlobalFlags.ReactionType = Superconduct

		return e
	case Electro:
		//extend??
		return e
	default:
		return NewNoAura()
	}
}

func NewElement() *Element {
	return &Element{}
}

func NewNoAura() *NoAura {
	r := &NoAura{}
	r.Element = &Element{}
	return r
}

func NewPyro() *PyroAura {
	r := &PyroAura{}
	r.Element = &Element{}
	return r
}
func NewHydro() *HydroAura {
	r := &HydroAura{}
	r.Element = &Element{}
	return r
}
func NewElectro() *ElectroAura {
	r := &ElectroAura{}
	r.Element = &Element{}
	return r
}
func NewCryo() *CryoAura {
	r := &CryoAura{}
	r.Element = &Element{}
	return r
}
func NewFreeze() *FreezeAura {
	r := &FreezeAura{}
	r.Element = &Element{}
	return r
}
func NewEC() *ElectroChargeAura {
	r := &ElectroChargeAura{}
	r.Element = &Element{}
	return r
}
