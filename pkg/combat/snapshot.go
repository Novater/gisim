package combat

type Snapshot struct {
	CharLvl  int64
	CharName string //name of the character triggering the damage

	Abil        string      //name of ability triggering the damage
	AbilType    ActionType  //type of ability triggering the damage
	WeaponClass WeaponClass //b.c. Gladiators...

	HitWeakPoint  bool
	IsHeavyAttack bool

	Mult      float64 //ability multiplier. could set to 0 from initial Mona dmg
	Element   EleType //element of ability
	UseDef    bool    //default false
	FlatDmg   float64 //flat dmg; so far only zhongli
	OtherMult float64 //so far just for xingqiu C4

	Stats    map[StatType]float64 //total character stats including from artifact, bonuses, etc...
	BaseAtk  float64              //base attack used in calc
	BaseDef  float64              //base def used in calc
	DmgBonus float64              //total damage bonus, including appropriate ele%, etc..
	DefMod   float64

	//reaction stuff
	ApplyAura bool  //if aura should be applied; false if under ICD
	AuraBase  int64 //unit base
	AuraUnits int64 //number of units

	//these are calculated fields
	WillReact bool //true if this will react
	//these two fields will only work if only reaction vs one element?!
	ReactionType ReactionType
	ReactedTo    EleType //NOT IMPLEMENTED

	IsMeltVape bool    //trigger melt/vape
	ReactMult  float64 //reaction multiplier for melt/vape
	ReactBonus float64 //reaction bonus %+ such as witch; should be 0 and only affected by hooks
}

func (s *Snapshot) Clone() Snapshot {
	c := Snapshot{}
	c = *s
	return c
}
