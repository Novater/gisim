package eula

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Eula", NewChar)
}

type eula struct {
	*combat.CharacterTemplate
	grimheartReset int
	burstCounter   int
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	e := eula{}
	t, err := combat.NewTemplateChar(s, p)

	if err != nil {
		return nil, err
	}
	e.CharacterTemplate = t
	e.Energy = 60
	e.MaxEnergy = 60
	e.Profile.WeaponClass = combat.WeaponClassSword

	e.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if _, ok := e.S.Status["Eula Burst"]; !ok {
			return false
		}
		if ds.CharName != e.Profile.Name {
			return false
		}
		if ds.AbilType != combat.ActionTypeBurst && ds.AbilType != combat.ActionTypeAttack && ds.AbilType != combat.ActionTypeSkill {
			return false
		}
		if _, ok := e.S.Status["Eula Burst ICD"]; ok {
			return false
		}
		//add to counter
		e.burstCounter++
		e.S.Status["Eula Burst ICD"] = 6
		return false
	}, "Eula Burst", combat.PostDamageHook)

	return &e, nil
}

func (e *eula) Attack(p map[string]interface{}) int {
	//register action depending on number in chain
	//3 and 4 need to be registered as multi action

	reset := false
	frames := 32 //first hit = 13 at 25fps
	delay := 10  //frames between execution and damage
	switch e.NormalCounter {
	case 1:
		frames = 29 //47 - 35
		delay = 10
	case 2:
		frames = 53 //69
		delay = 15
	case 3:
		frames = 24 //79
		delay = 20
	case 4:
		frames = 94 //118
		delay = 66
		reset = true
	}

	//apply attack speed
	frames = int(float64(frames) / (1 + e.Stats[combat.AtkSpd]))

	for i, mult := range auto[e.NormalCounter] {
		t := i + 1
		d := e.Snapshot("Normal", combat.ActionTypeAttack, combat.Physical)
		d.Mult = mult[e.Profile.TalentLevel[combat.ActionTypeAttack]-1]
		e.S.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(d)
			s.Log.Infof("\t Eula normal %v hit %v dealt %.0f damage", e.NormalCounter, t, damage)
		}, fmt.Sprintf("Eula-Normal-%v-%v", e.NormalCounter, t), delay)
	}

	e.NormalCounter++

	//add a 75 frame attackcounter reset
	e.NormalResetTimer = 70

	if reset {
		e.NormalResetTimer = 0
		e.NormalCounter = 0
	}
	//return animation cd
	//this also depends on which hit in the chain this is
	return frames
}

func (e *eula) Skill(p map[string]interface{}) int {
	if _, ok := e.CD[combat.SkillCD]; ok {
		e.S.Log.Debugf("\t Eula skill still on CD; skipping")
		return 0
	}
	hold := 0
	if v, ok := p["Hold"]; ok {
		h, n := v.(int)
		if n {
			hold = h
		}
	}
	switch hold {
	default:
		e.pressE()
		return 35
	case 1:
		e.holdE()
		return 108
	}

}

func (e *eula) pressE() {
	//press e (60fps vid)
	//starts 343 cancel 378
	lvl := e.Profile.TalentLevel[combat.ActionTypeSkill] - 1
	if e.Profile.Constellation >= 3 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	d := e.Snapshot("Icetide (Press)", combat.ActionTypeSkill, combat.Cryo)
	d.ApplyAura = true
	d.AuraBase = combat.WeakAuraBase
	d.AuraUnits = 1
	d.Mult = skillPress[lvl]

	e.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(d)
		s.Log.Infof("\t Eula skill (press) dealt %.0f damage", damage)
	}, "Eula-Skill-Press", 35)

	//RANDOM GUESS
	e.S.AddEnergyParticles("Eula", 2, combat.Cryo, 100)

	//add 1 stack to Grimheart
	v := e.Tags["Grimheart"]
	if v < 2 {
		v++
	}
	e.Tags["Grimheart"] = v
	e.S.Log.Debugf("\t Eula Grimheart stacks %v", v)
	e.grimheartReset = 18 * 60

	e.CD[combat.SkillCD] = 4 * 60
}

func (e *eula) holdE() {
	//hold e
	//296 to 341, but cd starts at 322
	//60 fps = 108 frames cast, cd starts 62 frames in so need to + 62 frames to cd
	lvl := e.Profile.TalentLevel[combat.ActionTypeSkill] - 1
	if e.Profile.Constellation >= 3 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	d := e.Snapshot("Icetide (Hold)", combat.ActionTypeSkill, combat.Cryo)
	d.ApplyAura = true
	d.AuraBase = combat.WeakAuraBase
	d.AuraUnits = 1
	d.Mult = skillHold[lvl]

	e.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(d)
		s.Log.Infof("\t Eula skill (hold) dealt %.0f damage", damage)
	}, "Eula-Skill-Hold", 108)

	//multiple brand hits
	v := e.Tags["Grimheart"]

	for i := 0; i < v; i++ {
		d := e.Snapshot("Icetide (Icewhirl)", combat.ActionTypeSkill, combat.Cryo)
		d.ApplyAura = true
		d.AuraBase = combat.WeakAuraBase
		d.AuraUnits = 1
		d.Mult = icewhirl[lvl]
		t := i + 1
		e.S.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(d)
			s.Log.Infof("\t Eula skill (ice whirl %v) dealt %.0f damage", t, damage)
		}, "Eula-Skill-Hold-Icewhirl", 108)
	}
	//RANDOM GUESS
	e.S.AddEnergyParticles("Eula", 2, combat.Cryo, 100)

	//add debuff per hit
	e.S.Target.AddResMod("Icewhirl Cryo", combat.ResistMod{
		Ele:      combat.Cryo,
		Value:    -resRed[lvl],
		Duration: 7 * v * 60,
	})
	e.S.Target.AddResMod("Icewhirl Physical", combat.ResistMod{
		Ele:      combat.Physical,
		Value:    -resRed[lvl],
		Duration: 7 * v * 60,
	})

	e.Tags["Grimheart"] = 0
	e.CD[combat.SkillCD] = 10*60 + 62
}

//ult 365 to 415, 60fps = 120
//looks like ult charges for 8 seconds
func (e *eula) Burst(p map[string]interface{}) int {
	e.S.Status["Eula Burst"] = 8 * 60
	e.burstCounter = 0

	lvl := e.Profile.TalentLevel[combat.ActionTypeBurst] - 1
	if e.Profile.Constellation >= 5 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}

	//add initial damage
	d := e.Snapshot("Glacial Illumination", combat.ActionTypeSkill, combat.Cryo)
	d.ApplyAura = true
	d.AuraBase = combat.WeakAuraBase
	d.AuraUnits = 1
	d.Mult = burstInitial[lvl]

	e.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(d)
		s.Log.Infof("\t Eula burst (initial) dealt %.0f damage", damage)
	}, "Eula-Burst-Initial", 100) //guess frames

	//add blow up after 8 seconds
	d2 := e.Snapshot("Glacial Illumination (Lightfall)", combat.ActionTypeSkill, combat.Cryo)
	d2.ApplyAura = true
	d2.AuraBase = combat.WeakAuraBase
	d2.AuraUnits = 1

	e.S.AddTask(func(s *combat.Sim) {
		stacks := e.burstCounter
		if stacks > 30 {
			stacks = 30
		}
		d2.Mult = burstExplodeBase[lvl] + burstExplodeStack[lvl]*float64(stacks)
		damage := s.ApplyDamage(d2)
		s.Log.Infof("\t Eula burst (lightfall) dealt %.0f damage, %v stacks", damage, stacks)
		e.S.Status["Eula Burst"] = 0
		e.burstCounter = 0
	}, "Eula-Burst-Lightfall", 8*60+100) //after 8 seconds

	e.CD[combat.BurstCD] = 20 * 60
	e.Energy = 0
	return 120
}

func (e *eula) Tick() {
	e.CharacterTemplate.Tick()
	e.grimheartReset--
	if e.grimheartReset == 0 {
		e.Tags["Grimheart"] = 0
	}
}
