package combat

type Profile struct {
	Label          string             `yaml:"Label"`
	Enemy          EnemyProfile       `yaml:"Enemy"`
	InitialActive  string             `yaml:"InitialActive"`
	Characters     []CharacterProfile `yaml:"Characters"`
	RotationString string             `yaml:"Rotation"`
	Rotation       []Action
	LogConfig      LogConfig
}

type LogConfig struct {
	LogLevel      string
	LogFile       string
	LogShowCaller bool
}

//CharacterProfile ...
type CharacterProfile struct {
	Base                 CharacterBase             `yaml:"Base"`
	Weapon               WeaponProfile             `yaml:"Weapon"`
	WeaponBonusConfig    map[string]float64        `yaml:"WeaponSecondaryStat"`
	AscensionBonusConfig map[string]float64        `yaml:"AscensionBonus"`
	TalentLevelConfig    map[string]int            `yaml:"TalentLevel"`
	ArtifactsConfig      map[string]ConfigArtifact `yaml:"Artifacts"`
}

type CharacterBase struct {
	Name    string  `yaml:"Name"`
	Element EleType `yaml:"Element"`
	Level   int64   `yaml:"Level"`
	HP      float64 `yaml:"BaseHP"`
	Atk     float64 `yaml:"BaseAtk"`
	Def     float64 `yaml:"BaseDef"`
	CR      float64 `yaml:"BaseCR"`
	CD      float64 `yaml:"BaseCD"`
	Cons    int     `yaml:"Constellation"`
}

type WeaponProfile struct {
	Name   string      `yaml:"WeaponName"`
	Class  WeaponClass `yaml:"WeaponClass"`
	Refine int         `yaml:"WeaponRefinement"`
	Atk    float64     `yaml:"WeaponBaseAtk"`
}

type TalentProfile struct {
	Attack int
	Skill  int
	Burst  int
}

//EnemyProfile ...
type EnemyProfile struct {
	Level  int64               `yaml:"Level"`
	Resist map[EleType]float64 `yaml:"Resist"`
}

type ConfigArtifact struct {
	Level int                `yaml:"Level"`
	Set   string             `yaml:"Set"`
	Main  map[string]float64 `yaml:"Main"`
	Sub   map[string]float64 `yaml:"Sub"`
}

//Slot identifies the artifact slot
type Slot int

//Types of artifact slots
const (
	Flower Slot = iota
	Feather
	Sands
	Goblet
	Circlet
)

func (s Slot) String() string {
	return [...]string{
		"Flower",
		"Feather",
		"Sands",
		"Goblet",
		"circlet",
	}[s]
}

type Stat struct {
	DEFP     float64
	DEF      float64
	HP       float64
	HPP      float64
	ATK      float64
	ATKP     float64
	ER       float64
	EM       float64
	CR       float64
	CD       float64
	Heal     float64
	PyroP    float64
	HydroP   float64
	CryoP    float64
	ElectroP float64
	AnemoP   float64
	GeoP     float64
	EleP     float64
	PhyP     float64
	DendroP  float64
	AtkSpd   float64
	DmgP     float64
}

type StatType int

//stat types
const (
	DEFP StatType = iota
	DEF
	HP
	HPP
	ATK
	ATKP
	ER
	EM
	CR
	CD
	Heal
	PyroP
	HydroP
	CryoP
	ElectroP
	AnemoP
	GeoP
	EleP
	PhyP
	DendroP
	AtkSpd
	DmgP
)

func (s StatType) String() string {
	return StatTypeString[s]
}

var StatTypeString = [...]string{
	"DEF%",
	"DEF",
	"HP",
	"HP%",
	"ATK",
	"ATK%",
	"ER",
	"EM",
	"CR",
	"CD",
	"Heal",
	"Pyro%",
	"Hydro%",
	"Cryo%",
	"Electro%",
	"Anemo%",
	"Geo%",
	"Ele%",
	"Phys%",
	"Dendro%",
	"ATKSPD%",
	"Dmg%",
}
