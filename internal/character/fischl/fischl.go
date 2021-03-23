package fischl

import (
	"fmt"
	"math/rand"

	"github.com/srliao/gisim/internal/character/common"
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Fischl", NewChar)
}

type fischl struct {
	*common.TemplateChar
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	f := fischl{}
	t, err := common.New(s, p)
	if err != nil {
		return nil, err
	}
	f.TemplateChar = t
	f.Energy = 60
	f.MaxEnergy = 60

	//register A4
	s.AddHook(func(ds *combat.Snapshot) bool {
		//don't trigger A4 if Fischl dealt dmg thereby triggering reaction
		if ds.CharName == "Fischl" {
			return false
		}
		//check reaction type, only care for overload, electro charge, superconduct
		ok := false
		switch ds.ReactionType {
		case combat.Overload:
			fallthrough
		case combat.ElectroCharged:
			fallthrough
		case combat.Superconduct:
			fallthrough
		case combat.Swirl:
			ok = true
		}
		if !ok {
			return false
		}
		//TODO: swirl with electro need to trigger this as well but how the hell do i check this???
		if ds.ReactionType == combat.Swirl && ds.ReactedTo != combat.Electro {
			return false
		}
		//check if Oz is on the field
		skillOz := s.HasEffect("Fischl-Oz-Skill")
		burstOz := s.HasEffect("Fischl-Oz-Burst")
		if !skillOz && !burstOz {
			return false
		}

		d := f.Snapshot("Fischl A4", combat.ActionTypeSpecialProc, combat.Electro)
		d.Mult = 0.8
		//apparently a4 doesnt apply electro
		s.AddAction(func(s *combat.Sim) bool {
			damage := s.ApplyDamage(d)
			s.Log.Infof("[%v]: Fischl (Oz - A4) dealt %.0f damage", s.Frame(), damage)
			return true
		}, fmt.Sprintf("%v-Fischl-C1", f.S.Frame()))

		return false
	}, "fischl a4", combat.PostReaction)

	if p.Constellation >= 1 {
		//if oz is not on field, trigger effect
		s.AddHook(func(ds *combat.Snapshot) bool {
			if ds.CharName != "Fischl" {
				return false
			}
			if ds.AbilType != combat.ActionTypeAttack {
				return false
			}
			d := f.Snapshot("Fischl C1", combat.ActionTypeSpecialProc, combat.Physical)
			d.Mult = 0.22
			s.AddAction(func(s *combat.Sim) bool {
				damage := s.ApplyDamage(d)
				s.Log.Infof("[%v]: Fischl (Oz - C1) dealt %.0f damage", s.Frame(), damage)
				return true
			}, fmt.Sprintf("%v-Fischl-C1", f.S.Frame()))

			return false
		}, "fischl c1", combat.PostDamageHook)
	}

	return &f, nil
}

//42
func (f *fischl) Skill() int {
	if _, ok := f.CD["skill-cd"]; ok {
		f.S.Log.Debugf("\tFischl skill still on CD; skipping")
		return 0
	}

	d := f.Snapshot("Oz", combat.ActionTypeSkill, combat.Electro)
	lvl := f.Profile.TalentLevel[combat.ActionTypeSkill] - 1
	if f.Profile.Constellation >= 3 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	d.Mult = birdSum[lvl]
	if f.Profile.Constellation >= 2 {
		d.Mult += 2
	}
	d.AuraBase = combat.WeakAuraBase
	d.AuraUnits = 1
	//clone b without info re aura
	b := d.Clone()
	b.Mult = birdAtk[lvl]

	//apply initial damage
	f.S.AddAction(func(s *combat.Sim) bool {
		if v, cd := f.CD[common.SkillICD]; cd {
			d.ApplyAura = false
			s.Log.Infof("[%v]: Fischl (Oz - summon) - aura app still on ICD (%v)", s.Frame(), v)
		} else {
			d.ApplyAura = true
			f.CD[common.SkillICD] = 150
		}
		damage := s.ApplyDamage(d)
		s.Log.Infof("[%v]: Fischl (Oz - summon) dealt %.0f damage", s.Frame(), damage)
		return true
	}, fmt.Sprintf("%v-Fischl-Skill-Initial", f.S.Frame()))

	//apply hit every 50 frames thereafter
	//NOT ENTIRELY ACCURATE BUT OH WELL
	tick := 0
	next := 40 + 50
	count := 0

	f.S.AddAction(func(s *combat.Sim) bool {
		tick++
		if tick < next {
			return false
		}
		if count >= 11 {
			return true
		}
		next += 50
		//share same icd
		if v, cd := f.CD[common.SkillICD]; cd {
			b.ApplyAura = false
			s.Log.Infof("[%v]: Fischl (Oz - tick) - aura app still on ICD (%v)", s.Frame(), v)
		} else {
			b.ApplyAura = true
			f.CD[common.SkillICD] = 150
		}
		damage := s.ApplyDamage(b)
		//assume fischl has 60% chance of generating orb every attack;
		if rand.Float64() < .6 {
			orbDelay := 0
			s.AddAction(func(s *combat.Sim) bool {
				if orbDelay < 60+60 { //random guess, 60 frame to generate, 60 frame to collect
					orbDelay++
					return false
				}
				s.GenerateOrb(1, combat.Electro, false)
				return true
			}, fmt.Sprintf("%v-Fischl-Skill-Orb-[%v]", s.Frame(), count))
		}
		s.Log.Infof("[%v]: Fischl (Oz - tick) dealt %.0f damage", s.Frame(), damage)
		count++
		return false
	}, fmt.Sprintf("%v-Fischl-Skill-Tick", f.S.Frame()))

	//register Oz with sim
	ozOnField := 0
	f.S.AddEffect(func(s *combat.Sim) bool {
		ozOnField++
		return ozOnField > 10*60
	}, "Fischl-Oz-Skill")

	f.CD["skill-cd"] = 25 * 60
	//return animation cd
	return 40
}

//first hit 40+50
//next + 50 at150
//last @ 620
