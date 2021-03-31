package fischl

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Fischl", NewChar)
}

type fischl struct {
	*combat.CharacterTemplate
	ozAuraICDHitCounter int //hit counter, apply every 4 hit
	ozAuraICDResetTimer int //timer in seconds, 5 seconds reset
	//track if oz is active
	ozActive      bool
	ozActiveTimer int
	//field use for calculating oz damage
	ozActiveSource string
	//will only shoot if active and CD == 0
	ozShootCD    int
	ozShootDelay int
	ozSnapshot   combat.Snapshot
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

	f.ozShootDelay = 50

	//register A4
	f.a4()

	if p.Base.Cons >= 1 {
		f.c1()
	}
	return &f, nil
}

func (f *fischl) a4() {
	f.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		//don't trigger A4 if Fischl dealt dmg thereby triggering reaction
		if ds.Actor == "Fischl" {
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
		if !f.ozActive {
			return false
		}
		//apparently a4 doesnt apply electro
		d := f.Snapshot("Fischl A4", combat.ActionTypeSpecialProc, combat.Electro, combat.WeakDurability)
		d.Mult = 0.8
		f.S.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(d)
			s.Log.Infof("\t Fischl (Oz - A4) dealt %.0f damage", damage)
		}, "Fischl A4", 1)

		return false
	}, "fischl a4", combat.PostReaction)
}

func (f *fischl) c1() {
	//if oz is not on field, trigger effect
	f.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if f.ozActive {
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

func (f *fischl) ozShoot() {
	//don't shoot if oz not around
	if !f.ozActive {
		return
	}
	//check oz active timer
	f.ozActiveTimer--
	if f.ozActiveTimer == 0 {
		f.ozActiveTimer = 0
		f.ozActive = false
		return
	}
	//don't shoot if oz shoot timer is on cd
	if f.ozShootCD > 0 {
		f.ozShootCD--
		return
	}
	d := f.ozSnapshot.Clone()
	if f.ozAuraICDHitCounter%4 == 0 {
		//apply aura, add to timer
		d.Durability = combat.WeakDurability
		f.ozAuraICDResetTimer = 5 * 60 // every 5 second force reset
		f.ozAuraICDHitCounter++
	}
	//so oz is active and ready to shoot, we add damage
	f.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(d)
		s.Log.Infof("\t Fischl (Oz - %v) dealt %.0f damage", f.ozActiveSource, damage)
	}, "Fischl Oz (Damage)", 1)
	f.ozShootCD += f.ozShootDelay
	//assume fischl has 60% chance of generating orb every attack;
	if f.S.Rand.Float64() < .6 {
		f.S.AddEnergyParticles("Fischl", 1, combat.Electro, 120)
	}
}

//42
func (f *fischl) Skill(p map[string]interface{}) int {
	cd := f.CD[combat.SkillCD]
	if cd > f.S.F {
		f.S.Log.Debugf("\tFischl skill still on CD; skipping")
		return 0
	}

	//reset oz
	f.ozActive = true
	f.ozActiveTimer = 10 * 60 //10 second summon
	f.ozActiveSource = "Skill"
	f.ozShootCD = 40    //40 frames before first shot
	f.ozShootDelay = 50 //50 frames per shot
	//reset hit counter as well not sure if this is true though
	f.ozAuraICDHitCounter = 0
	f.ozAuraICDResetTimer = 0
	//put a tag on the sim
	f.S.Status["Fischl-Oz"] = f.S.F + 10*60

	d := f.Snapshot("Oz", combat.ActionTypeSkill, combat.Electro, combat.WeakDurability)
	d.Mult = birdSum[f.TalentLvlSkill()]
	if f.Base.Cons >= 2 {
		d.Mult += 2
	}
	//set on field oz to be this one
	f.ozSnapshot = d
	//clone b without info re aura
	b := d.Clone()
	b.Mult = birdAtk[f.TalentLvlSkill()]
	b.ApplyAura = true
	//apply initial damage
	f.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(d)
		s.Log.Infof("\t Fischl (Oz - Skill Initial) dealt %.0f damage", damage)
	}, "Fischl Skill Initial", 1)

	f.CD[combat.SkillCD] = f.S.F + 25*60
	//return animation cd
	return 40
}

//first hit 40+50
//next + 50 at150
//last @ 620

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

	//reset oz
	f.ozActive = true
	f.ozActiveTimer = 10 * 60 //10 second summon
	f.ozActiveSource = "Burst"
	f.ozShootCD = 40    //40 frames before first shot
	f.ozShootDelay = 50 //50 frames per shot
	//reset hit counter as well not sure if this is true though
	f.ozAuraICDHitCounter = 0
	f.ozAuraICDResetTimer = 0
	//put a tag on the sim
	f.S.Status["Fischl-Oz"] = f.S.F + 10*60

	//initial damage
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
	b := f.Snapshot("Midnight Phantasmagoria (Oz)", combat.ActionTypeBurst, combat.Electro, 0)
	b.Mult = birdAtk[f.TalentLvlSkill()]
	f.ozSnapshot = b

	f.Energy = 0
	f.CD[combat.BurstCD] = f.S.F + 15*60
	return 21 //this is if you cancel immediately
}

func (f *fischl) Tick() {
	f.CharacterTemplate.Tick()
	f.ozAuraICDResetTimer--
	if f.ozAuraICDResetTimer < 0 {
		f.ozAuraICDResetTimer = 0
		f.ozAuraICDHitCounter = 0
	}
	f.ozShoot()
}
