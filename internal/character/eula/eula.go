package eula

import (
	"fmt"

	"github.com/srliao/gisim/internal/rotation"
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("eula", NewChar)
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
	e.Energy = 80
	e.MaxEnergy = 80
	e.Weapon.Class = combat.WeaponClassClaymore

	e.a4()

	e.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if !e.S.StatusActive("Eula Burst") {
			return false
		}
		if ds.Actor != e.Base.Name {
			return false
		}
		if ds.AbilType != rotation.ActionBurst && ds.AbilType != rotation.ActionAttack && ds.AbilType != rotation.ActionSkill {
			return false
		}
		if e.S.StatusActive("Eula Burst ICD") {
			return false
		}
		//add to counter
		e.burstCounter++
		s.Log.Debugf("\t Eula burst adding 1 stack, current count %v", e.burstCounter)
		e.S.Status["Eula Burst ICD"] = e.S.F + 6
		return false
	}, "Eula Burst", combat.PostDamageHook)

	return &e, nil
}

func (e *eula) a4() {
	e.S.AddEventHook(func(s *combat.Sim) bool {
		if s.ActiveChar != e.Base.Name {
			return false
		}
		//reset CD, add 1 stack
		v := e.Tags["grimheart"]
		if v < 2 {
			v++
		}
		e.Tags["grimheart"] = v

		s.Log.Debugf("\t Eula A4 resetting skill CD, new cd %v", s.F-1)
		e.CD[combat.SkillCD] = s.F - 1

		return false
	}, "eula-a4", combat.PostBurstHook)
}

func (e *eula) a4Old() {
	e.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor != e.Base.Name {
			return false
		}
		if ds.AbilType != rotation.ActionAttack {
			return false
		}
		//check icd
		cd := e.CD["A4-ICD"]
		if cd >= e.S.F {
			return false
		}

		e.S.Log.Debugw("\t Eula A4 triggered", "skill previous cd", e.CD[combat.SkillCD])
		e.CD["A4-ICD"] = e.S.F + 6

		e.CD[combat.SkillCD] -= 18

		return false
	}, "eula-a4", combat.OnCritDamage)
}

func (e *eula) Attack(p int) int {
	//register action depending on number in chain
	//3 and 4 need to be registered as multi action

	reset := false
	frames := 29       //first hit = 13 at 25fps
	delay := []int{11} //frames between execution and damage
	switch e.NormalCounter {
	case 1:
		frames = 25 //47 - 35
		delay = []int{25}
	case 2:
		frames = 65 //69
		delay = []int{36, 49}
	case 3:
		frames = 33 //79
		delay = []int{33}
	case 4:
		frames = 88 //118
		delay = []int{45, 63}
		reset = true
	}

	//apply attack speed
	frames = int(float64(frames) / (1 + e.Stats[combat.AtkSpd]))

	for i, mult := range auto[e.NormalCounter] {
		t := i + 1
		d := e.Snapshot("Normal", rotation.ActionAttack, combat.Physical, combat.WeakDurability)
		d.Mult = mult[e.TalentLvlAttack()]
		e.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Eula normal %v hit %v dealt %.0f damage [%v]", e.NormalCounter, t, damage, str)
		}, fmt.Sprintf("Eula-Normal-%v-%v", e.NormalCounter, t), delay[i])
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

func (e *eula) Skill(p int) int {
	cd := e.CD[combat.SkillCD]
	if cd >= e.S.F {
		e.S.Log.Debugf("\t Eula skill still on CD; skipping")
		return 0
	}
	hold := p
	switch hold {
	default:
		e.pressE()
		return 34
	case 1:
		e.holdE()
		return 80
	}

}

func (e *eula) pressE() {
	//press e (60fps vid)
	//starts 343 cancel 378
	d := e.Snapshot("Icetide (Press)", rotation.ActionSkill, combat.Cryo, combat.WeakDurability)
	d.Mult = skillPress[e.TalentLvlSkill()]

	e.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Eula skill (press) dealt %.0f damage [%v]", damage, str)
	}, "Eula-Skill-Press", 35)

	//RANDOM GUESS
	n := 2
	if e.S.Rand.Float64() < .5 {
		n = 1
	}
	e.S.AddEnergyParticles("Eula", n, combat.Cryo, 100)

	//add 1 stack to Grimheart
	v := e.Tags["grimheart"]
	if v < 2 {
		v++
	}
	e.Tags["grimheart"] = v
	e.S.Log.Debugf("\t Eula Grimheart stacks %v", v)
	e.grimheartReset = 18 * 60

	e.CD[combat.SkillCD] = e.S.F + 4*60
}

func (e *eula) holdE() {
	//hold e
	//296 to 341, but cd starts at 322
	//60 fps = 108 frames cast, cd starts 62 frames in so need to + 62 frames to cd
	lvl := e.TalentLvlSkill()
	d := e.Snapshot("Icetide (Hold)", rotation.ActionSkill, combat.Cryo, combat.WeakDurability)
	d.Mult = skillHold[lvl]

	e.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Eula skill (hold) dealt %.0f damage [%v]", damage, str)
	}, "Eula-Skill-Hold", 80)

	//multiple brand hits
	v := e.Tags["grimheart"]

	for i := 0; i < v; i++ {
		d := e.Snapshot("Icetide (Icewhirl)", rotation.ActionSkill, combat.Cryo, combat.WeakDurability)
		d.Mult = icewhirl[lvl]
		t := i + 1
		e.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Eula skill (ice whirl %v) dealt %.0f damage [%v]", t, damage, str)
		}, "Eula-Skill-Hold-Icewhirl", 90+i*7) //we're basically forcing it so we get 3 stacks
	}

	//A2
	if v == 2 {
		d := e.Snapshot("Icetide (Lightfall)", rotation.ActionSkill, combat.Cryo, combat.WeakDurability)
		d.Mult = burstExplodeBase[e.TalentLvlBurst()] * 0.5
		e.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Eula A2 on 2 Grimheart stacks dealt %.0f damage [%v]", damage, str)
		}, "Eula-Skill-Hold-A2-Lightfall", 108) //we're basically forcing it so we get 3 stacks
	}
	//RANDOM GUESS
	n := 3
	if e.S.Rand.Float64() < .5 {
		n = 2
	}
	e.S.AddEnergyParticles("Eula", n, combat.Cryo, 100)

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

	e.Tags["grimheart"] = 0
	e.CD[combat.SkillCD] = e.S.F + 10*60 + 62
}

//ult 365 to 415, 60fps = 120
//looks like ult charges for 8 seconds
func (e *eula) Burst(p int) int {
	e.S.Status["Eula Burst"] = e.S.F + 7*60 + 115 //add animation time
	e.burstCounter = 0
	lvl := e.TalentLvlBurst()
	//add initial damage
	d := e.Snapshot("Glacial Illumination", rotation.ActionBurst, combat.Cryo, combat.WeakDurability)
	d.Mult = burstInitial[lvl]

	e.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Eula burst (initial) dealt %.0f damage [%v]", damage, str)
	}, "Eula-Burst-Initial", 116) //guess frames

	e.S.AddTask(func(s *combat.Sim) {
		stacks := e.burstCounter
		if stacks > 30 {
			stacks = 30
		}
		//add blow up after 8 seconds
		d2 := e.Snapshot("Glacial Illumination (Lightfall)", rotation.ActionBurst, combat.Physical, combat.WeakDurability)
		d2.Mult = burstExplodeBase[lvl] + burstExplodeStack[lvl]*float64(stacks)
		damage, str := s.ApplyDamage(d2)
		s.Log.Infof("\t Eula burst (lightfall) dealt %.0f damage, %v stacks [%v]", damage, stacks, str)
		e.S.Status["Eula Burst"] = e.S.F
		e.burstCounter = 0
	}, "Eula-Burst-Lightfall", 7*60+116) //after 8 seconds

	e.CD[combat.BurstCD] = e.S.F + 20*60
	e.Energy = 0
	return 116
}

func (e *eula) Tick() {
	e.CharacterTemplate.Tick()
	e.grimheartReset--
	if e.grimheartReset == 0 {
		e.Tags["grimheart"] = 0
	}
}

func (e *eula) Snapshot(name string, t rotation.ActionType, x combat.EleType, d float64) combat.Snapshot {
	s := e.CharacterTemplate.Snapshot(name, t, x, d)
	s.IsHeavyAttack = true
	return s
}
