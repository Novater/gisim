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
	SwirlElectro   ReactionType = "swirl-electro"
	SwirlHydro     ReactionType = "swirl-hydro"
	SwirlPyro      ReactionType = "swirl-pyro"
	SwirlCryo      ReactionType = "swirl-cryo"
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
	case SwirlElectro:
		fallthrough
	case SwirlCryo:
		fallthrough
	case SwirlHydro:
		fallthrough
	case SwirlPyro:
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
	// lvlm = lvlMult[ds.CharLvl]

	res := s.Target.Resist(s.Log)[ds.Element]
	resmod := 1 - res/2
	if res >= 0 && res < 0.75 {
		resmod = 1 - res
	} else if res > 0.75 {
		resmod = 1 / (4*res + 1)
	}
	s.Log.Debugw("\t\treact dmg", "em", em, "lvl", cl, "lvl mod", lvlm, "type", r, "ele", t, "mult", mult, "res", res, "res mod", resmod, "bonus", ds.ReactBonus)

	damage := mult * lvlm * (1 + ((6.66 * em) / (1400 + em)) + ds.ReactBonus) * resmod
	s.Log.Infof("[%v] %v (%v) caused %v, dealt %v damage", s.Frame(), ds.Actor, ds.Abil, r, damage)
	s.Target.Damage += damage
	s.Target.HP -= damage
	s.Stats.DamageByChar[ds.Actor][string(r)] += damage

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

var lvlMult = []float64{
	17.2,
	18.5,
	19.9,
	21.3,
	22.6,
	24.6,
	26.6,
	28.9,
	31.4,
	34.1,
	37.2,
	40.7,
	44.4,
	48.6,
	53.7,
	59.1,
	64.4,
	69.7,
	75.1,
	80.6,
	86.1,
	91.7,
	97.2,
	102.8,
	108.4,
	113.2,
	118.1,
	123,
	129.7,
	136.3,
	142.7,
	149,
	155.4,
	161.8,
	169.1,
	176.5,
	184.1,
	191.7,
	199.6,
	207.4,
	215.4,
	224.2,
	233.5,
	243.4,
	256.1,
	268.5,
	281.5,
	295,
	309.1,
	323.6,
	336.8,
	350.5,
	364.5,
	378.6,
	398.6,
	416.4,
	434.4,
	452.6,
	471.4,
	490.5,
	509.5,
	532.8,
	556.4,
	580.1,
	607.9,
	630.2,
	652.9,
	675.2,
	697.8,
	720.2,
	742.5,
	765.2,
	784.4,
	803.4,
	830.9,
	854.4,
	877.8,
	900.1,
	923.8,
	946.4,
	968.6,
	991,
	1013.5,
	1036.1,
	1066.6,
	1090,
	1115,
	1141.7,
	1171.9,
	1202.8,
}
