package combat

type NewSetFunc func(c Character, s *Sim, count int)

func RegisterSetFunc(name string, f NewSetFunc) {
	mu.Lock()
	defer mu.Unlock()
	if _, dup := setMap[name]; dup {
		panic("combat: RegisterSetBonus called twice for character " + name)
	}
	setMap[name] = f
}

type ArtifactSet struct {
	Flower  Artifact
	Feather Artifact
	Sand    Artifact
	Goblet  Artifact
	Circlet Artifact
}

//Artifact represents one artfact
type Artifact struct {
	Level    int64  `yaml:"Level"`
	Set      string `yaml:"Set"`
	Type     Slot   `yaml:"Type"`
	MainStat Stat   `yaml:"MainStat"`
	Substat  []Stat `yaml:"Substat"`
}

//Stat represents one stat
type Stat struct {
	Type  StatType `yaml:"Type"`
	Value float64  `yaml:"Value"`
}

//Slot identifies the artifact slot
type Slot string

//Types of artifact slots
const (
	Flower  Slot = "Flower"
	Feather Slot = "Feather"
	Sands   Slot = "Sands"
	Goblet  Slot = "Goblet"
	Circlet Slot = "Circlet"
)

//StatType defines what stat it is
type StatType string

//stat types
const (
	DEFP     StatType = "DEF%"
	DEF      StatType = "DEF"
	HP       StatType = "HP"
	HPP      StatType = "HP%"
	ATK      StatType = "ATK"
	ATKP     StatType = "ATK%"
	ER       StatType = "ER"
	EM       StatType = "EM"
	CR       StatType = "CR"
	CD       StatType = "CD"
	Heal     StatType = "Heal"
	PyroP    StatType = "Pyro%"
	HydroP   StatType = "Hydro%"
	CryoP    StatType = "Cryo%"
	ElectroP StatType = "Electro%"
	AnemoP   StatType = "Anemo%"
	GeoP     StatType = "Geo%"
	EleP     StatType = "Ele%"
	PhyP     StatType = "Phys%"
	DendroP  StatType = "Dendro%"
	AtkSpd   StatType = "ATKSPD%"
	DmgP     StatType = "Dmg%"
)

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
	return ""
}
