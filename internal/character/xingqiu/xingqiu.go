package xingqiu

import (
	"fmt"

	"github.com/srliao/gisim/internal/character/common"
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Xingqiu", NewChar)
}

type xingqiu struct {
	*common.TemplateChar
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	x := xingqiu{}
	t, err := common.New(s, p)
	if err != nil {
		return nil, err
	}
	x.TemplateChar = t
	x.Energy = 80
	x.MaxEnergy = 80
	x.Profile.WeaponClass = combat.WeaponClassSword

	a4 := make(map[combat.StatType]float64)
	a4[combat.HydroP] = 0.2
	x.AddMod("Xingqiu A4", a4)

	//c2
	if x.Profile.Constellation >= 2 {
		s.Log.Debugf("\tactivating Xingqiu C2")

		s.AddCombatHook(func(snap *combat.Snapshot) bool {
			//check if c1 debuff is on, if so, reduce resist by -0.15
			if _, ok := s.Target.Status["xingqiu-c2"]; ok {
				s.Log.Debugf("\t[%v]: applying Xingqiu C2 hydro debuff", s.Frame())
				snap.ResMod[combat.Hydro] -= 0.15
			}
			return false
		}, "xingqiu-c2", combat.PreDamageHook)
	}

	/** c6
	Activating 2 of Guhua Sword: Raincutter's sword rain attacks greatly increases the DMG of the third.
	Xingqiu regenerates 3 Energy when sword rain attacks hit opponents.
	**/

	return &x, nil
}

func (x *xingqiu) Attack(p map[string]interface{}) int {
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
		d := x.Snapshot("Normal", combat.ActionTypeAttack, combat.Physical)
		d.Mult = hit[x.Profile.TalentLevel[combat.ActionTypeAttack]-1]
		//add a 20 frame delay; should be 18 and 42 for combo 3 and 5 actual
		delay := 0
		t := i + 1
		x.S.AddAction(func(s *combat.Sim) bool {
			if delay < 20 {
				delay++
				return false
			}
			//no delay for now? realistically the hits should have delay but not sure if it actually makes a diff
			//since it doesnt apply any elements, only trigger weapon procs
			c := d.Clone()
			damage := s.ApplyDamage(c)
			s.Log.Infof("[%v]: Xingqiu normal %v (hit %v) dealt %.0f damage", s.Frame(), n, t, damage)
			return true
		}, fmt.Sprintf("%v-Xingqiu-Normal-%v-%v", x.S.Frame(), n, i))
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

func (x *xingqiu) Skill(p map[string]interface{}) int {
	//applies wet to self 30 frame after cast
	if _, ok := x.CD[common.SkillCD]; ok {
		x.S.Log.Debugf("\tXingqiu skill still on CD; skipping")
		return 0
	}
	d := x.Snapshot("Guhua Sword: Fatal Rainscreen", combat.ActionTypeSkill, combat.Hydro)
	lvl := x.Profile.TalentLevel[combat.ActionTypeSkill] - 1
	if x.Profile.Constellation >= 5 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	if x.Profile.Constellation >= 4 {
		//check if ult is up, if so increase multiplier
		if _, ok := x.S.Status["Xingqiu-Burst"]; ok {
			d.OtherMult = 1.5
		}
	}
	d.ApplyAura = true
	d.AuraBase = combat.WeakAuraBase
	d.AuraUnits = 1
	d.Mult = rainscreen[0][lvl]
	d2 := d.Clone()
	d2.Mult = rainscreen[1][lvl]

	tick := 0
	x.S.AddAction(func(s *combat.Sim) bool {
		tick++
		if tick < 19 { //first hit 19 frames
			return false
		}
		damage := s.ApplyDamage(d)
		s.Log.Infof("[%v]: Xingqiu skill hit 1 dealt %.0f damage", s.Frame(), damage)
		return true
	}, fmt.Sprintf("%v-Xingqiu-Skill-1", x.S.Frame()))

	x.S.AddAction(func(s *combat.Sim) bool {
		tick++
		if tick < 39 { //second 39
			return false
		}
		damage := s.ApplyDamage(d2)
		s.Log.Infof("[%v]: Xingqiu skill hit 2 dealt %.0f damage", s.Frame(), damage)
		return true
	}, fmt.Sprintf("%v-Xingqiu-Skill-2", x.S.Frame()))

	orbDelay := 0
	x.S.AddAction(func(s *combat.Sim) bool {
		if orbDelay < 100 { //generated on the 37th, 100 frames to get orbs if standing right up against
			orbDelay++
			return false
		}
		s.GenerateOrb(5, combat.Hydro, false)
		return true
	}, fmt.Sprintf("%v-Xingqiu-Skill-Orb", x.S.Frame()))

	//should last 15s, cd 21s
	x.CD[common.SkillCD] = 21 * 60
	return 77
}

func (x *xingqiu) Burst(p map[string]interface{}) int {
	//apply hydro every 3rd hit
	//triggered on normal attack
	//not sure what ICD is
	if _, ok := x.CD[common.BurstCD]; ok {
		x.S.Log.Debugf("\tXingqiu skill still on CD; skipping")
		return 0
	}
	//also applies hydro on cast
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
	if x.Profile.Constellation >= 2 {
		dur += 3
	}
	dur = dur * 60
	x.S.Status["Xingqiu-Burst"] = dur

	burstCounter := 0
	swords := 2
	x.S.AddCombatHook(func(ds *combat.Snapshot) bool {
		//check if buff is up
		if _, ok := x.S.Status["Xingqiu-Burst"]; !ok {
			return true //remove
		}
		//check if off ICD
		if _, ok := x.CD["Xingqiu-Burst-ICD"]; ok {
			return false
		}
		//check if normal attack
		if ds.AbilType != combat.ActionTypeAttack && ds.AbilType != combat.ActionTypeChargedAttack {
			return false
		}

		lvl := x.Profile.TalentLevel[combat.ActionTypeBurst] - 1
		if x.Profile.Constellation >= 3 {
			lvl += 3
			if lvl > 14 {
				lvl = 14
			}
		}

		//trigger swords, only first sword applies hydro
		for i := 0; i < swords; i++ {

			d := x.Snapshot("Guhua Sword: Raincutter", combat.ActionTypeBurst, combat.Hydro)
			d.Mult = burst[lvl]

			//apply aura every 3rd hit -> hit 0, 3, 6, etc...
			//only first sword summoned can apply aura
			if burstCounter%3 == 0 && i == 0 {
				d.ApplyAura = true
				d.AuraBase = combat.WeakAuraBase
				d.AuraUnits = 1
			}
			t := i + 1

			delay := 0
			x.S.AddAction(func(s *combat.Sim) bool {
				if delay < 20+(i) { //20 frames after trigger to do dmg, + i for each sword; TODO
					delay++
					return false
				}
				damage := s.ApplyDamage(d)
				s.Log.Infof("[%v]: Xingqiu burst proc hit %v dealt %.0f damage", s.Frame(), t, damage)
				//add hydro debuff for 4s
				if x.Profile.Constellation >= 2 {
					s.Target.Status["xingqiu-c2"] = 4 * 60
				}
				//on first hit, if C6, recover 3 energy
				if x.Profile.Constellation == 6 && t-1 == 0 {
					s.Log.Debugf("\tXingqiu C6 regenerating energy previous % next %v", x.Energy, x.Energy+3)
					x.Energy += 3
					if x.Energy > x.MaxEnergy {
						x.Energy = x.MaxEnergy
					}
				}
				return true
			}, fmt.Sprintf("%v-Xingqiu-Burst-Proc-Hit-%v", x.S.Frame(), i+1))
		}

		//figure out next wave # of swords
		switch swords {
		case 2:
			swords = 3
		case 3:
			if x.Profile.Constellation == 6 {
				swords = 5
			} else {
				swords = 2
			}
		case 5:
			swords = 2
		}

		//estimated 1 second ICD
		x.CD["Xingqiu-Burst-ICD"] = 60
		burstCounter++

		return false
	}, "Xingqiu-Burst", combat.PostDamageHook)

	//remove the hook if off CD
	x.S.AddAction(func(s *combat.Sim) bool {
		_, ok := s.Status["Xingqiu-Burst"]
		if !ok {
			s.RemoveCombatHook("Xingqiu-Burst", combat.PostDamageHook)
		}
		return !ok
	}, fmt.Sprintf("%v-Xingqiu-Burst", x.S.Frame()))

	x.CD[common.BurstCD] = 20 * 60
	x.Energy = 0
	return 39
}
