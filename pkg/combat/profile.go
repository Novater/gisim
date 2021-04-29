package combat

import (
	"fmt"
	"strings"
)

type Profile struct {
	Label         string             `yaml:"Label"`
	Enemy         EnemyProfile       `yaml:"Enemy"`
	InitialActive string             `yaml:"InitialActive"`
	Characters    []CharacterProfile `yaml:"Characters"`
	Rotation      []Action
	LogConfig     LogConfig
}

type LogConfig struct {
	LogLevel      string
	LogFile       string
	LogShowCaller bool
}

//CharacterProfile ...
type CharacterProfile struct {
	Base    CharacterBase `yaml:"Base"`
	Weapon  WeaponProfile `yaml:"Weapon"`
	Talents TalentProfile
	Stats   []float64
	Sets    map[string]int
}

type CharacterBase struct {
	Name    string  `yaml:"Name"`
	Element EleType `yaml:"Element"`
	Level   int     `yaml:"Level"`
	HP      float64 `yaml:"BaseHP"`
	Atk     float64 `yaml:"BaseAtk"`
	Def     float64 `yaml:"BaseDef"`
	Cons    int     `yaml:"Constellation"`
}

type WeaponProfile struct {
	Name   string      `yaml:"WeaponName"`
	Class  WeaponClass `yaml:"WeaponClass"`
	Refine int         `yaml:"WeaponRefinement"`
	Atk    float64     `yaml:"WeaponBaseAtk"`
	Param  int
}

type TalentProfile struct {
	Attack int
	Skill  int
	Burst  int
}

//EnemyProfile ...
type EnemyProfile struct {
	Level  int                 `yaml:"Level"`
	Resist map[EleType]float64 `yaml:"Resist"`
}

type ConfigArtifact struct {
	Level int                `yaml:"Level"`
	Set   string             `yaml:"Set"`
	Main  map[string]float64 `yaml:"Main"`
	Sub   map[string]float64 `yaml:"Sub"`
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
	"def%",
	"def",
	"hp",
	"hp%",
	"atk",
	"atk%",
	"er",
	"em",
	"cr",
	"cd",
	"heal",
	"pyro%",
	"hydro%",
	"cryo%",
	"electro%",
	"anemo%",
	"geo%",
	"ele%",
	"phys%",
	"dendro%",
	"atkspd%",
	"dmg%",
}

func StrToStatType(s string) StatType {
	for i, v := range StatTypeString {
		if v == s {
			return StatType(i)
		}
	}
	return -1
}

func PrettyPrintStats(stats []float64) string {
	if len(stats) != len(StatTypeString) {
		return fmt.Sprintf("invalid number of items: %v\n", stats)
	}
	var sb strings.Builder
	for i, v := range stats {
		if v != 0 {
			sb.WriteString(fmt.Sprintf("%v: %.3f ", StatTypeString[i], v))
		}
	}
	return sb.String()
}
