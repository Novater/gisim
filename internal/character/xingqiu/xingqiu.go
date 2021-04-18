package xingqiu

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Xingqiu", NewChar)
}

type xingqiu struct {
	*combat.CharacterTemplate
	numSwords    int
	burstCounter int
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	x := xingqiu{}
	t, err := combat.NewTemplateChar(s, p)
	if err != nil {
		return nil, err
	}
	x.CharacterTemplate = t
	x.Energy = 80
	x.MaxEnergy = 80
	x.Weapon.Class = combat.WeaponClassSword

	a4 := make(map[combat.StatType]float64)
	a4[combat.HydroP] = 0.2
	x.AddMod("Xingqiu A4", a4)
	x.burstHook()

	/** c6
	Activating 2 of Guhua Sword: Raincutter's sword rain attacks greatly increases the DMG of the third.
	Xingqiu regenerates 3 Energy when sword rain attacks hit opponents.
	**/

	return &x, nil
}

func (x *xingqiu) c2() {
	x.S.Target.AddResMod("xingqiu-c2", combat.ResistMod{
		Ele:      combat.Hydro,
		Value:    -0.15,
		Duration: 4 * 60,
	})
}

func (x *xingqiu) Attack(p int) int {
	//register action depending on number in chain
	//3 and 4 need to be registered as multi action
	//figure out which hit it is
	var hits [][]float64
	reset := false
	frames := 23 //first hit
	n := 1
	switch x.NormalCounter {
	case 1:
		hits = n2
		frames = 49 - 23
		n = 2
	case 2:
		hits = n3
		frames = 79 - 49 //61 for first half
		n = 3
	case 3:
		hits = n4
		frames = 106 - 79
		n = 4
	case 4:
		hits = n5
		frames = 178 - 106 //135 for first half
		n = 5
		reset = true
	default:
		hits = n1
	}
	x.NormalCounter++
	//apply attack speed
	frames = int(float64(frames) / (1 + x.Stats[combat.AtkSpd]))

	for i, hit := range hits {
		d := x.Snapshot("Normal", combat.ActionTypeAttack, combat.Physical, combat.WeakDurability)
		d.Mult = hit[x.TalentLvlAttack()]
		//add a 20 frame delay; should be 18 and 42 for combo 3 and 5 actual
		t := i + 1
		x.S.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(d)
			s.Log.Infof("\t Xingqiu normal %v (hit %v) dealt %.0f damage", n, t, damage)
		}, fmt.Sprintf("Xingqiu-Normal-%v-%v", n, i), 20)
	}

	//add a 75 frame attackcounter reset
	x.NormalResetTimer = 70

	if reset {
		x.NormalResetTimer = 0
	}
	//return animation cd
	//this also depends on which hit in the chain this is
	return frames
}

func (x *xingqiu) Skill(p int) int {
	//applies wet to self 30 frame after cast
	if x.CD[combat.SkillCD] > x.S.F {
		x.S.Log.Debugf("\tXingqiu skill still on CD; skipping")
		return 0
	}
	d := x.Snapshot("Guhua Sword: Fatal Rainscreen", combat.ActionTypeSkill, combat.Hydro, combat.WeakDurability)
	if x.Base.Cons >= 4 {
		//check if ult is up, if so increase multiplier
		if x.S.StatusActive("Xingqiu-Burst") {
			d.OtherMult = 1.5
		}
	}
	d.Mult = rainscreen[0][x.TalentLvlSkill()]
	d2 := d.Clone()
	d2.Mult = rainscreen[1][x.TalentLvlSkill()]

	x.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(d)
		s.Log.Infof("\t Xingqiu skill hit 1 dealt %.0f damage", damage)
	}, "Xingqiu-Skill-1", 19) //first hit 19 frames

	x.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(d2)
		s.Log.Infof("\t Xingqiu skill hit 2 dealt %.0f damage", damage)
	}, "Xingqiu-Skill-2", 39) //second hit 39 frames

	x.S.AddEnergyParticles("Xingqiu", 5, combat.Hydro, 100)

	//should last 15s, cd 21s
	x.CD[combat.SkillCD] = x.S.F + 21*60
	return 77
}

func (x *xingqiu) burstHook() {
	x.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		//check if buff is up
		if !x.S.StatusActive("Xingqiu-Burst") {
			return false
		}
		//check if off ICD
		if x.CD["Xingqiu-Burst-ICD"] > x.S.F {
			return false
		}
		//check if normal attack
		if ds.AbilType != combat.ActionTypeAttack && ds.AbilType != combat.ActionTypeChargedAttack {
			return false
		}

		//trigger swords, only first sword applies hydro
		for i := 0; i < x.numSwords; i++ {

			d := x.Snapshot("Guhua Sword: Raincutter", combat.ActionTypeBurst, combat.Hydro, 0)
			d.Mult = burst[x.TalentLvlBurst()]

			//apply aura every 3rd hit -> hit 0, 3, 6, etc...
			//only first sword summoned can apply aura
			if x.burstCounter%3 == 0 && i == 0 {
				d.Durability = combat.WeakDurability
			}
			t := i + 1

			x.S.AddTask(func(s *combat.Sim) {
				damage := s.ApplyDamage(d)
				s.Log.Infof("\t Xingqiu burst proc hit %v dealt %.0f damage", t, damage)
				if x.Base.Cons >= 2 {
					x.c2()
				}
				if x.Base.Cons == 6 && t-1 == 0 {
					s.Log.Debugf("\tXingqiu C6 regenerating energy previous %v next %v", x.Energy, x.Energy+3)
					x.Energy += 3
					if x.Energy > x.MaxEnergy {
						x.Energy = x.MaxEnergy
					}
				}
			}, fmt.Sprintf("Xingqiu-Burst-Proc-Hit-%v", i+1), 20+i) //second hit 39 frames
		}

		//figure out next wave # of swords
		switch x.numSwords {
		case 2:
			x.numSwords = 3
		case 3:
			if x.Base.Cons == 6 {
				x.numSwords = 5
			} else {
				x.numSwords = 2
			}
		case 5:
			x.numSwords = 2
		}

		//estimated 1 second ICD
		x.CD["Xingqiu-Burst-ICD"] = x.S.F + 60
		x.burstCounter++

		return false
	}, "Xingqiu-Burst", combat.PostDamageHook)
}

func (x *xingqiu) Burst(p int) int {
	//apply hydro every 3rd hit
	//triggered on normal attack
	//not sure what ICD is
	if x.CD[combat.BurstCD] > x.S.F {
		x.S.Log.Debugf("\tXingqiu skill still on CD; skipping")
		return 0
	}
	//also applies hydro on cast
	d := x.Snapshot("Xingqiu Burst", combat.ActionTypeBurst, combat.Hydro, 25)
	d.Mult = 0
	x.S.AddTask(func(s *combat.Sim) {
		s.ApplyDamage(d)
		s.Log.Infof("\t Xingqiu initial burst applying hydro")
	}, "Xingqiu-Burst-Initial", 20) //second hit 39 frames
	//how we doing that?? trigger 0 dmg?

	/**
	The number of Hydro Swords summoned per wave follows a specific pattern, usually alternating between 2 and 3 swords.
	At C6, this is upgraded and follows a pattern of 2 → 3 → 5… which then repeats.

	There is an approximately 1 second interval between summoned Hydro Sword waves, so that means a theoretical maximum of 15 or 18 waves.

	Each wave of Hydro Swords is capable of applying one (1) source of Hydro status, and each individual sword is capable of getting a crit.
	**/

	/** c2
	Extends the duration of Guhua Sword: Raincutter by 3s.
	Decreases the Hydro RES of opponents hit by sword rain attacks by 15% for 4s.
	**/
	dur := 15
	if x.Base.Cons >= 2 {
		dur += 3
	}
	dur = dur * 60
	x.S.Status["Xingqiu-Burst"] = x.S.F + dur

	x.burstCounter = 0
	x.numSwords = 2

	x.CD[combat.BurstCD] = x.S.F + 20*60
	x.Energy = 0
	return 39
}
