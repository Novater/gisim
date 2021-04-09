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
	Start         int //when the aura was first added
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
	s.Log.Debugf("\t element %v duration refreshed", e.E())
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
	return e.Expiry < s.F //TODO: not sure if this should be <, or <=
}

func (e *Element) Attach(ele EleType, durability float64, f int) {
	e.Start = f
	e.Type = ele
	e.MaxDurability = durability * 0.8 //TODO not sure on this part
	e.Durability = durability * 0.8
	e.Expiry = f + auraDuration(durability)
}

//calculate how much of the durability remains after passage of time
func (e *Element) Remain(f int) float64 {
	if f <= e.Expiry {
		return float64((e.Expiry-f)/(e.Expiry-e.Start)) * e.Durability
	}
	return 0
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
	}
	return p
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
	case Hydro:
		h.Refresh(ds.Durability, s)
	case Cryo:
		//freeze??
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = Freeze
	case Electro:
		//ec??
		e := NewElectro()
		e.Attach(ds.Element, ds.Durability, s.F)
		ec := NewEC()
		ec.Init(e, h, ds, s)
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = ElectroCharged
		return ec
	}
	return h
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
	case Electro:
		//ec??
		e.Refresh(ds.Durability, s)
	}
	return e
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
	case Hydro:
		//figure out units to reduce by
		r := c.Durability * (1 - float64(s.F-c.Start)/float64(c.Expiry))
		if r > ds.Durability {
			r = ds.Durability
		}
		h := NewHydro()
		h.Attach(ds.Element, ds.Durability, s.F) //TODO: not sure if this part is accurate
		h.Reduce(r, s)
		c.Reduce(r, s)
		f := NewFreeze()
		f.Init(c, h, r, s)
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = Freeze
		return f
	case Cryo:
		c.Refresh(ds.Durability, s)
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
	}
	return c
}

type FreezeAura struct {
	*Element
	OldHydro *HydroAura
	OldCryo  *CryoAura
	NewHydro *HydroAura
	NewCryo  *CryoAura
	Expire   int
}

const ax = -0.08652942693631073258
const bx = 10.48493477815584323194

func (f *FreezeAura) Init(cryo *CryoAura, hydro *HydroAura, dur float64, s *Sim) {
	f.Expiry = s.F + int(ax*dur*dur+bx*dur)
	f.OldCryo = cryo
	f.OldHydro = hydro
	f.Type = Frozen
}

//TODO: what happens if we apply new hydro then new cryo all before freeze expires? shouldn't be possible to due cds but
//right now this code wouldn't retrigger freeze even though it probably should
func (f *FreezeAura) React(ds Snapshot, s *Sim) Aura {
	if ds.Durability == 0 && !ds.IsHeavyAttack {
		return f
	}
	//check heavy attack
	if ds.IsHeavyAttack {
		//freeze is done, trigger shatter damage
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.NextAttackShatterTriggered = true
		//check if we have new aura to return
		if f.NewCryo != nil {
			return f.NewCryo
		}
		if f.NewHydro != nil {
			return f.NewHydro
		}
		//otherwise return max duration of the old
		if f.OldCryo.Expiry > s.F && f.OldHydro.Expiry > s.F {
			if f.OldCryo.Expiry > f.OldHydro.Expiry {
				return f.OldCryo
			}
			return f.OldHydro
		}
		if f.OldCryo.Expiry > s.F {
			return f.OldCryo
		}
		if f.OldHydro.Expiry > s.F {
			return f.OldHydro
		}
		return NewNoAura()
	}
	switch ds.Element {
	case Pyro:
		//just melt
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = Melt
		//no residual aura; melt always triggers 2x so we wipe out regardless
		return NewNoAura()
	case Hydro:
		//extend freeze?
		if f.NewCryo != nil {
			r := f.NewCryo.Durability * (1 - float64(s.F-f.NewCryo.Start)/float64(f.NewCryo.Expiry))
			if r > ds.Durability {
				r = ds.Durability
			}
			h := NewHydro()
			h.Attach(ds.Element, ds.Durability, s.F) //TODO: not sure if this part is accurate
			h.Reduce(r, s)
			f.NewCryo.Reduce(r, s)
			nf := NewFreeze()
			nf.Init(f.NewCryo, h, r, s)
			s.GlobalFlags.ReactionDidOccur = true
			s.GlobalFlags.ReactionType = Freeze
			return nf
		}
		//overwrite?
		h := NewHydro()
		h.Attach(ds.Element, ds.Durability, s.F)
		f.NewHydro = h
		return f
	case Cryo:
		//extend freeze?
		if f.NewHydro != nil {
			r := f.NewHydro.Durability * (1 - float64(s.F-f.NewHydro.Start)/float64(f.NewHydro.Expiry))
			if r > ds.Durability {
				r = ds.Durability
			}
			c := NewCryo()
			c.Attach(ds.Element, ds.Durability, s.F) //TODO: not sure if this part is accurate
			c.Reduce(r, s)
			f.NewHydro.Reduce(r, s)
			nf := NewFreeze()
			nf.Init(c, f.NewHydro, r, s)
			s.GlobalFlags.ReactionDidOccur = true
			s.GlobalFlags.ReactionType = Freeze
			return nf
		}
		//overwrite?
		c := NewCryo()
		c.Attach(ds.Element, ds.Durability, s.F)
		f.NewCryo = c
		return f
	case Electro:
		//just superconuct
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.ReactionType = Superconduct
		//no residual aura
		return NewNoAura()
	default:
		return f
	}
}

func (f *FreezeAura) Tick(s *Sim) bool {
	//do nothing if freeze not defrosted yet
	if f.Expiry > s.F {
		return false
	}
	//if we have new underlying
	if f.NewCryo != nil {
		s.TargetAura = f.NewCryo
		return false
	}
	if f.NewHydro != nil {
		s.TargetAura = f.NewHydro
		return false
	}
	//if both nil then just expire (only shatter exposes underlying)
	return true
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
	case Hydro:
		//first 984 (barb), tick 1055 (barb dmg), apply 1078(barb), (no more electro), apply 1276 (electro - bd dmg), apply 1353 (water - barb dmg)
		//1418 (razor apply dmg), 1449 (barb reapply), 1481 (xq apply), 1570 xq tick, 1719 razor, 1776 razor tick, 1801 barb auto, 1856 barb tick
		//so basically whatever triggers next overrides the dmg profile and ticks down using the new profile

		//here we extend the original hydro and then trigger a tick immediately
		e.Hydro.Refresh(ds.Durability, s)
		//also reset the snapshot
		e.Snap = ds.Clone()
		e.NextTick = s.F
	case Cryo:
		//just superconduct
		s.GlobalFlags.ReactionDidOccur = true
		s.GlobalFlags.NextAttackSuperconductTriggered = true
		s.GlobalFlags.ReactionType = Superconduct
	case Electro:
		e.Electro.Refresh(ds.Durability, s)
		//also reset the snapshot
		e.Snap = ds.Clone()
		e.NextTick = s.F
	}

	return e
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
