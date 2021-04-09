package fischl

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Fischl", NewChar)
}

type fischl struct {
	*combat.CharacterTemplate
	ozAttackCounter  int
	ozActiveUntil    int
	ozNextShootReady int
	ozICD            int
	//field use for calculating oz damage
	ozActiveSource string
	ozSnapshot     combat.Snapshot
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	f := fischl{}
	t, err := combat.NewTemplateChar(s, p)
	if err != nil {
		return nil, err
	}
	f.CharacterTemplate = t
	f.Energy = 60
	f.MaxEnergy = 60
	f.Weapon.Class = combat.WeaponClassBow

	//register A4
	f.a4()

	if p.Base.Cons >= 1 {
		f.c1()
	}

	if p.Base.Cons == 6 {
		f.c6()
	}

	f.turbo()

	return &f, nil
}

func (f *fischl) turbo() {
	f.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		//if Oz dealt damage
		if ds.Actor != "Fischl" {
			return false
		}
		//do nothing if oz not on field
		if f.ozActiveUntil < f.S.F {
			return false
		}
		if ds.AbilType != combat.ActionTypeSkill {
			return false
		}
		if f.S.GlobalFlags.ReactionType != combat.Overload && f.S.GlobalFlags.ReactionType != combat.Superconduct {
			return false
		}
		//trigger one particle
		f.S.Log.Debugf("\t Fischl (Oz) turbo triggered")
		f.S.AddEnergyParticles("Fischl", 1, combat.Electro, 120)

		return false
	}, "fischl turbo", combat.PostReaction)
}

func (f *fischl) a4() {
	f.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		//don't trigger A4 if Fischl dealt dmg thereby triggering reaction
		if ds.Actor == "Fischl" {
			return false
		}
		//check reaction type, only care for overload, electro charge, superconduct
		ok := false
		switch f.S.GlobalFlags.ReactionType {
		case combat.Overload:
			fallthrough
		case combat.ElectroCharged:
			fallthrough
		case combat.Superconduct:
			fallthrough
		case combat.SwirlElectro:
			ok = true
		}
		if !ok {
			return false
		}
		//do nothing if oz not on field
		if f.ozActiveUntil < f.S.F {
			return false
		}

		d := f.Snapshot("Fischl A4", combat.ActionTypeSpecialProc, combat.Electro, 0)
		d.Mult = 0.8
		if f.ozAttackCounter%4 == 0 {
			//apply aura, add to timer
			d.Durability = combat.WeakDurability
			f.ozICD = f.S.F + 300 //add 300 second to skill ICD
		}
		f.S.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(d)
			s.Log.Infof("\t Fischl (Oz - A4) dealt %.0f damage", damage)
		}, "Fischl A4", 1)
		//increment hit counter
		f.ozAttackCounter++
		return false
	}, "fischl a4", combat.PostReaction)
}

func (f *fischl) c1() {
	//if oz is not on field, trigger effect
	f.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		//do nothing if oz active
		if f.ozActiveUntil >= f.S.F {
			return false
		}
		if ds.Actor != "Fischl" {
			return false
		}
		if ds.AbilType != combat.ActionTypeAttack {
			return false
		}
		d := f.Snapshot("Fischl C1", combat.ActionTypeSpecialProc, combat.Physical, combat.WeakDurability)
		d.Mult = 0.22
		f.S.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(d)
			s.Log.Infof("\t Fischl (Oz - C1) dealt %.0f damage", damage)
		}, "Fischl C1", 1)

		return false
	}, "fischl c1", combat.PostDamageHook)
}

func (f *fischl) c6() {
	f.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		//do nothing if oz not on field
		if f.ozActiveUntil < f.S.F {
			return false
		}
		if ds.AbilType != combat.ActionTypeAttack {
			return false
		}

		d := f.Snapshot("Fischl C6", combat.ActionTypeSpecialProc, combat.Electro, 0)
		d.Mult = 0.3
		if f.ozAttackCounter%4 == 0 {
			//apply aura, add to timer
			d.Durability = combat.WeakDurability
			f.ozICD = f.S.F + 300 //add 300 second to skill ICD
		}
		f.S.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(d)
			s.Log.Infof("\t Fischl (Oz - C6) dealt %.0f damage", damage)
		}, "Fischl C6", 1)
		//increment hit counter
		f.ozAttackCounter++
		return false
	}, "fischl c6", combat.PostDamageHook)
}

func (f *fischl) ozAttack() {
	d := f.ozSnapshot.Clone()
	d.Durability = 0
	if f.ozAttackCounter%4 == 0 {
		//apply aura, add to timer
		d.Durability = combat.WeakDurability
		f.ozICD = f.S.F + 300 //add 300 second to skill ICD
	}
	//so oz is active and ready to shoot, we add damage
	f.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(d)
		s.Log.Infof("\t Fischl (Oz - %v) dealt %.0f damage", f.ozActiveSource, damage)
	}, "Fischl Oz (Damage)", 1)
	//put shoot on cd
	f.ozNextShootReady = f.S.F + 50
	//increment hit counter
	f.ozAttackCounter++
	//assume fischl has 60% chance of generating orb every attack;
	if f.S.Rand.Float64() < .6 {
		f.S.AddEnergyParticles("Fischl", 1, combat.Electro, 120)
	}
}

func (f *fischl) Attack(p map[string]interface{}) int {

	frames := []int{29, 21, 40, 45, 31}
	delay := []int{40, 40, 40, 40, 40}
	return f.CharacterTemplate.AttackHelperSingle(frames, delay, auto)
}

func (f *fischl) Skill(p map[string]interface{}) int {
	cd := f.CD[combat.SkillCD]
	if cd > f.S.F {
		f.S.Log.Debugf("\tFischl skill still on CD; skipping")
		return 0
	}
	c6extend := 0
	if f.Base.Cons == 6 {
		c6extend = 120
	}
	//reset oz
	f.ozActiveUntil = f.S.F + 600 + c6extend
	f.ozActiveSource = "Skill"
	f.ozNextShootReady = f.S.F + 40 //wait 40 before first shot
	//reset hit counter as well not sure if this is true though
	f.ozAttackCounter = 0
	//put a tag on the sim
	f.S.Status["Fischl-Oz"] = f.S.F + 600 + c6extend

	//always trigger electro no ICD on initial summon
	d := f.Snapshot("Oz", combat.ActionTypeSkill, combat.Electro, combat.WeakDurability)
	d.Mult = birdSum[f.TalentLvlSkill()]
	if f.Base.Cons >= 2 {
		d.Mult += 2
	}
	//apply initial damage
	f.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(d)
		s.Log.Infof("\t Fischl (Oz - Skill Initial) dealt %.0f damage", damage)
	}, "Fischl Skill Initial", 1)

	//set on field oz to be this one
	f.ozSnapshot = f.Snapshot("Oz", combat.ActionTypeSkill, combat.Electro, 0)
	f.ozSnapshot.Mult = birdAtk[f.TalentLvlSkill()]

	f.CD[combat.SkillCD] = f.S.F + 25*60
	//return animation cd
	return 40
}

func (f *fischl) Burst(p map[string]interface{}) int {
	cd := f.CD[combat.BurstCD]
	if cd > f.S.F {
		f.S.Log.Debugf("\t Fischl burst still on CD; skipping")
		return 0
	}
	//check if sufficient energy
	if f.Energy < f.MaxEnergy {
		f.S.Log.Debugf("\t Fischl burst - insufficent energy, current: %v", f.Energy)
		return 0
	}

	c6extend := 0
	if f.Base.Cons == 6 {
		c6extend = 120
	}
	//reset oz
	f.ozActiveUntil = f.S.F + 600 + c6extend
	f.ozActiveSource = "Skill"
	f.ozNextShootReady = f.S.F + 40 //wait 40 before first shot
	//reset hit counter as well not sure if this is true though
	f.ozAttackCounter = 0
	//put a tag on the sim
	f.S.Status["Fischl-Oz"] = f.S.F + 600 + c6extend

	//initial damage; part of the burst tag
	d := f.Snapshot("Midnight Phantasmagoria", combat.ActionTypeBurst, combat.Electro, combat.WeakDurability)
	d.Mult = burst[f.TalentLvlBurst()]
	//apply initial damage
	f.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(d)
		s.Log.Infof("\t Fischl (Burst Initial) dealt %.0f damage", damage)
	}, "Fischl Burst Initial", 1)

	//check for C4 damage
	if f.Base.Cons >= 4 {
		d1 := f.Snapshot("Midnight Phantasmagoria C4", combat.ActionTypeSpecialProc, combat.Electro, 0)
		d1.Mult = 2.22
		f.S.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(d)
			s.Log.Infof("\t Fischl (Burst C4) dealt %.0f damage", damage)
		}, "Fischl Burst C4", 1)
	}

	//snapshot for Oz
	f.ozSnapshot = f.Snapshot("Midnight Phantasmagoria (Oz)", combat.ActionTypeSkill, combat.Electro, 0)
	f.ozSnapshot.Mult = birdAtk[f.TalentLvlSkill()]

	f.Energy = 0
	f.CD[combat.BurstCD] = f.S.F + 15*60
	return 21 //this is if you cancel immediately
}

func (f *fischl) Tick() {
	f.CharacterTemplate.Tick()

	//check if oz is active
	if f.ozActiveUntil > f.S.F {
		//check if icd is reset
		if f.ozICD < f.S.F {
			f.ozAttackCounter = 0
		}
		//check if we should be shooting
		if f.ozNextShootReady <= f.S.F {
			//ok now we should shoot
			f.ozAttack()
		}
	}

}
