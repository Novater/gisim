package combat

//Artifact represents one artfact
type Artifact struct {
	Level    int64  `yaml:"Level"`
	Set      string `yaml:"Set"`
	Type     Slot   `yaml:"Type"`
	MainStat Stat   `yaml:"MainStat"`
	Substat  []Stat `yaml:"Substat"`
}

//ArtifactSet represents a set of artifacts
type ArtifactSet map[Slot]Artifact

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
)
