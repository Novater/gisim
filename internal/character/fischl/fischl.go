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
	ozHitCounter int //hit counter, apply every 4 hit
	ozResetTimer int //timer in seconds, 5 seconds reset
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
	f.Profile.WeaponClass = combat.WeaponClassBow

	//register A4
	s.AddCombatHook(func(ds *combat.Snapshot) bool {
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
		_, skillOz := s.Status["Fischl-Oz-Skill"]
		_, burstOz := s.Status["Fischl-Oz-Burst"]
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
		s.AddCombatHook(func(ds *combat.Snapshot) bool {
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
func (f *fischl) Skill(p map[string]interface{}) int {
	if _, ok := f.CD[common.SkillCD]; ok {
		f.S.Log.Debugf("\tFischl skill still on CD; skipping")
		return 0
	}

	//cancel existing bird right away if any
	f.S.RemoveAction("Fischl-Oz-Burst")
	delete(f.S.Status, "Fischl-Oz-Burst")

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
	b.ApplyAura = true

	//apply initial damage
	f.S.AddAction(func(s *combat.Sim) bool {
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
		//Oz has two ICD, an internal hit counter and a reset timer
		//apply aura every 4th hit
		if f.ozHitCounter%4 == 0 {
			//apply aura, add to timer
			b.ApplyAura = true
			f.ozResetTimer = 5 * 60 // every 5 second force reset
			f.ozHitCounter++
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
	}, "Fischl-Oz-Skill")

	//register Oz with sim
	f.S.Status["Fischl-Oz-Skill"] = 10 * 60

	f.CD["skill-cd"] = 25 * 60
	//return animation cd
	return 40
}

//first hit 40+50
//next + 50 at150
//last @ 620

func (f *fischl) Burst(p map[string]interface{}) int {
	if _, ok := f.CD[common.BurstCD]; ok {
		f.S.Log.Debugf("\tFischl burst still on CD; skipping")
		return 0
	}
	//cancel existing bird right away if any
	f.S.RemoveAction("Fischl-Oz-Skill")
	delete(f.S.Status, "Fischl-Oz-Skill")

	d := f.Snapshot("Midnight Phantasmagoria", combat.ActionTypeBurst, combat.Electro)
	lvl := f.Profile.TalentLevel[combat.ActionTypeBurst] - 1
	if f.Profile.Constellation >= 5 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	d.Mult = burst[lvl]
	d.AuraBase = combat.WeakAuraBase
	d.AuraUnits = 1
	d.ApplyAura = true
	//apply initial damage
	f.S.AddAction(func(s *combat.Sim) bool {
		damage := s.ApplyDamage(d)
		s.Log.Infof("[%v]: Fischl (burst) dealt %.0f damage", s.Frame(), damage)
		return true
	}, fmt.Sprintf("%v-Fischl-Burst-Initial", f.S.Frame()))

	//check for C4 damage
	if f.Profile.Constellation >= 4 {
		d1 := f.Snapshot("Midnight Phantasmagoria C4", combat.ActionTypeSpecialProc, combat.Electro)
		d1.Mult = 2.22
		d1.AuraBase = combat.WeakAuraBase
		d1.AuraUnits = 1
		f.S.AddAction(func(s *combat.Sim) bool {
			damage := s.ApplyDamage(d1)
			s.Log.Infof("[%v]: Fischl (burst C4) dealt %.0f damage", s.Frame(), damage)
			return true
		}, fmt.Sprintf("%v-Fischl-Burst-C4", f.S.Frame()))
	}

	//add new bird after a delay

	//apply hit every 50 frames thereafter
	//NOT ENTIRELY ACCURATE BUT OH WELL
	tick := 0
	next := 40 + 50 //bird starts casting after 50, the initial is the oz animation time; we keep at 40 for now even though we're returning 20
	count := 0

	b := f.Snapshot("Midnight Phantasmagoria (Oz)", combat.ActionTypeBurst, combat.Electro)
	blvl := f.Profile.TalentLevel[combat.ActionTypeSkill] - 1
	if f.Profile.Constellation >= 3 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	b.Mult = birdAtk[blvl]
	b.ApplyAura = false
	b.AuraBase = combat.WeakAuraBase
	b.AuraUnits = 1

	f.S.AddAction(func(s *combat.Sim) bool {
		tick++
		if tick < next {
			return false
		}
		if count >= 11 {
			return true
		}
		next += 50
		//Oz has two ICD, an internal hit counter and a reset timer
		//apply aura every 4th hit
		if f.ozHitCounter%4 == 0 {
			//apply aura, add to timer
			b.ApplyAura = true
			f.ozResetTimer = 5 * 60 // every 5 second force reset
			f.ozHitCounter++
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
			}, fmt.Sprintf("%v-Fischl-Burst(Oz)-Orb-[%v]", s.Frame(), count))
		}
		s.Log.Infof("[%v]: Fischl (Burst - Oz tick) dealt %.0f damage", s.Frame(), damage)
		count++
		return false
	}, "Fischl-Oz-Burst")

	//register Oz with sim
	f.S.Status["Fischl-Oz-Burst"] = 10 * 60
	f.Energy = 0
	f.CD[common.BurstCD] = 15 * 60
	return 21 //this is if you cancel immediately
}

func (f *fischl) Tick() {
	f.TemplateChar.Tick()
	f.ozResetTimer--
	if f.ozResetTimer < 0 {
		f.ozResetTimer = 0
		f.ozHitCounter = 0
	}
}
