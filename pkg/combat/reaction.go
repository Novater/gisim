package combat

func EleToDmgP(e EleType) StatType {
	switch e {
	case Anemo:
		return AnemoP
	case Cryo:
		return CryoP
	case Electro:
		return ElectroP
	case Geo:
		return GeoP
	case Hydro:
		return HydroP
	case Pyro:
		return PyroP
	case Dendro:
		return DendroP
	case Physical:
		return PhyP
	}
	return -1
}

type ReactionType string

const (
	Overload       ReactionType = "overload"
	Superconduct   ReactionType = "superconduct"
	Freeze         ReactionType = "freeze"
	Melt           ReactionType = "melt"
	Vaporize       ReactionType = "vaporize"
	Crystallize    ReactionType = "crystallize"
	Swirl          ReactionType = "swirl"
	ElectroCharged ReactionType = "electrocharged"
	Shatter        ReactionType = "shatter"
	NoReaction     ReactionType = ""
	FreezeExtend   ReactionType = "FreezeExtend"
)

//split this out as a separate function so we can call it if we need to apply EC or Burning damage
func (s *Sim) applyReactionDamage(ds Snapshot, r ReactionType) float64 {
	var mult float64
	var t EleType
	switch r {
	case Overload:
		mult = 4
		t = Pyro
	case Superconduct:
		mult = 1
		t = Cryo
	case ElectroCharged:
		mult = 2.4
		t = Electro
	case Shatter:
		mult = 3
		t = Physical
	case Swirl:
		mult = 1.2
		//what element is this??
	default:
		//either not implemented or no dmg
		return 0
	}
	em := ds.Stats[EM]
	//lvl bonus is clearly a line of best fit...
	//need some datamining here
	cl := float64(ds.CharLvl)
	lvlm := 0.0002325*cl*cl*cl + 0.05547*cl*cl - 0.2523*cl + 14.74

	res := s.Target.Resist(s.Log)[ds.Element]
	resmod := 1 - res/2
	if res >= 0 && res < 0.75 {
		resmod = 1 - res
	} else if res > 0.75 {
		resmod = 1 / (4*res + 1)
	}
	s.Log.Debugw("\t\treact dmg", "em", em, "lvl", cl, "lvl mod", lvlm, "type", ds.ReactionType, "ele", t, "mult", mult, "res", res, "res mod", resmod, "bonus", ds.ReactBonus)

	damage := mult * (1 + ((6.66 * em) / (1400 + em)) + ds.ReactBonus) * lvlm * resmod
	s.Log.Infof("[%v] %v (%v) caused %v, dealt %v damage", s.Frame(), ds.Actor, ds.Abil, ds.ReactionType, damage)
	s.Target.Damage += damage
	s.Target.DamageDetails[ds.Actor][string(r)] += damage

	return damage
}

//EleType is a string representing an element i.e. HYDRO/PYRO/etc...
type EleType string

//ElementType should be pryo, Hydro, Cryo, Electro, Geo, Anemo and maybe dendro
const (
	Pyro      EleType = "pyro"
	Hydro     EleType = "hydro"
	Cryo      EleType = "cryo"
	Electro   EleType = "electro"
	Geo       EleType = "geo"
	Anemo     EleType = "anemo"
	Dendro    EleType = "dendro"
	Physical  EleType = "physical"
	Frozen    EleType = "frozen"
	EC        EleType = "electro-charged"
	NoElement EleType = ""
)

const (
	WeakAuraBase   int64 = 570
	MedAuraBase    int64 = 360
	StrongAuraBase int64 = 225
)
