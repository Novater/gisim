package bennett

import (
	"strconv"
	"strings"

	"github.com/srliao/gisim/internal/rotation"
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("bennett", NewChar)
}

type bennett struct {
	*combat.CharacterTemplate
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	b := bennett{}
	t, err := combat.NewTemplateChar(s, p)

	if err != nil {
		return nil, err
	}
	b.CharacterTemplate = t
	b.Energy = 60
	b.MaxEnergy = 60
	b.Weapon.Class = combat.WeaponClassSword

	//add effect for burst
	pc := burstatk[b.TalentLvlBurst()]
	if b.Base.Cons >= 1 {
		pc += 0.2
	}
	atk := pc * float64(b.Base.Atk+b.Weapon.Atk)
	b.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if !b.S.StatusActive("Bennett Burst") {
			return false
		}
		if b.S.ActiveChar != ds.Actor {
			return false
		}
		//TODO: should have an HP check here but no one ever takes damage in this sim..
		b.S.Log.Debugf("\t applying Bennet burst buff; adding %v", atk)
		ds.Stats[combat.ATK] += atk
		if b.Base.Cons == 6 {
			ok := ds.AbilType == rotation.ActionAttack || ds.AbilType == rotation.ActionCharge
			ok = ok && (ds.WeaponClass == combat.WeaponClassSpear || ds.WeaponClass == combat.WeaponClassSword || ds.WeaponClass == combat.WeaponClassClaymore)
			if ok {
				ds.Element = combat.Pyro
				ds.Stats[combat.PyroP] += 0.15
			}
		}
		return false
	}, "Bennett-Burst-Field", combat.PostSnapshot)

	return &b, nil
}

func (b *bennett) Attack(p int) int {
	//register action depending on number in chain
	//3 and 4 need to be registered as multi action
	d := b.Snapshot("Normal", rotation.ActionAttack, combat.Physical, combat.MedDurability)
	//figure out which hit it is
	var hits []float64
	reset := false
	frames := 21 //first hit
	delay := 10  //frames between execution and damage
	n := 1
	var sb strings.Builder
	sb.WriteString("Bennet-Normal-")
	switch b.NormalCounter {
	case 1:
		hits = n2
		frames = 48 - 21
		delay = 10
		n = 2
		sb.WriteRune('2')
	case 2:
		hits = n3
		frames = 74 - 48
		delay = 15
		n = 3
		sb.WriteRune('3')
	case 3:
		hits = n4
		frames = 114 - 74
		delay = 20
		n = 4
		sb.WriteRune('4')
	case 4:
		hits = n5
		frames = 180 - 114
		delay = 66
		n = 5
		reset = true
		sb.WriteRune('5')
	default:
		hits = n1
		sb.WriteRune('1')
	}
	b.NormalCounter++
	//apply attack speed
	frames = int(float64(frames) / (1 + b.Stats[combat.AtkSpd]))
	d.Mult = hits[b.TalentLvlAttack()]

	sb.Write([]byte(strconv.Itoa(n)))
	b.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Bennett normal %v dealt %.0f damage [%v]", n, damage, str)
	}, sb.String(), delay)

	//add a 75 frame attackcounter reset
	b.NormalResetTimer = 70

	if reset {
		b.NormalResetTimer = 0
	}
	//return animation cd
	//this also depends on which hit in the chain this is
	return frames
}

func (b *bennett) ChargeAttackStam() float64 {
	return 20
}

func (b *bennett) Skill(p int) int {
	cd := b.CD[combat.SkillCD]
	if cd > b.S.F {
		b.S.Log.Debugf("\tBennett skill still on CD; skipping")
		return 0
	}
	hold := p
	var sb strings.Builder
	sb.WriteString("Bennett-Skill-Hold-")
	var hits [][]float64
	delay := []int{26}
	switch hold {
	case 0:
		hits = skill
		sb.WriteRune('0')
	case 1:
		delay = []int{89, 115}
		hits = skill1
		sb.WriteRune('1')
	case 2:
		delay = []int{136, 154, 198}
		hits = skill2
		sb.WriteRune('2')
	}
	sb.WriteString("-Hit-")

	for i, s := range hits {
		d := b.Snapshot("Passion Overload", rotation.ActionSkill, combat.Pyro, combat.MedDurability)
		d.Mult = s[b.TalentLvlBurst()]
		t := i + 1
		sb.WriteString(strconv.Itoa(t))
		b.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Bennett skill dealt %.0f damage [%v]", damage, str)
		}, sb.String(), delay[i])
	}

	count := 3
	if b.S.Rand.Float64() < .5 {
		count = 2
	}
	b.S.AddEnergyParticles("Bennett", count, combat.Pyro, delay[0]+100)

	//A2
	reduction := 0.2
	//A4
	if _, ok := b.S.Status["Bennett Burst"]; ok {
		reduction += 0.5
	}

	switch hold {
	case 1:
		cd := int(7.5 * 60 * (1 - reduction))
		b.CD[combat.SkillCD] = cd
		return 153 //not right?
	case 2:
		cd := int(10 * 60 * (1 - reduction))
		b.CD[combat.SkillCD] = cd
		return 370 //too high as well
	}

	cd = int(5 * 60 * (1 - reduction))
	b.CD[combat.SkillCD] = b.S.F + cd //should be 7.5 or 10
	return 52
}

func (b *bennett) Burst(p int) int {
	cd := b.CD[combat.BurstCD]
	if cd > b.S.F {
		b.S.Log.Debugf("\tBennett burst still on CD; skipping")
		return 0
	}
	//check if sufficient energy
	if b.Energy < b.MaxEnergy {
		b.S.Log.Debugf("\t Bennett burst - insufficent energy, current: %v", b.Energy)
		return 0
	}

	//add field effect timer
	b.S.Status["Bennett Burst"] = 720
	//we should be adding repeating tasks here every 1sec to heal active char but
	//no one takes damage anyways

	//hook for buffs; active right away after cast

	d := b.Snapshot("Fantastic Voyage", rotation.ActionBurst, combat.Pyro, combat.MedDurability)
	d.Mult = burst[b.TalentLvlBurst()]

	b.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Bennett burst dealt %.0f damage [%v]", damage, str)
	}, "Bennett-Burst", 43)

	b.Energy = 0
	b.CD[combat.BurstCD] = b.S.F + 15*60
	return 51 //todo fix field cast time
}
