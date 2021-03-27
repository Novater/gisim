package diluc

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Diluc", NewChar)
}

type diluc struct {
	*combat.CharacterTemplate
	eCounter    int
	eResetTimer int
}

/**

Skills:
	A2: Diluc's Charged Attack Stamina Cost is decreased by 50%, and its duration is increased by 3s.

	A4: The Pyro Enchantment provided by Dawn lasts for 4s longer. Additionally, Diluc gains 20% Pyro DMG Bonus during the duration of this effect.

	C1: Diluc deals 15% more DMG to enemies whose HP is above 50%.

	C2: When Diluc takes DMG, his Base ATK increases by 10% and his ATK SPD increases by 5%. Lasts for 10s.
		This effect can stack up to 3 times and can only occur once every 1.5s.

	C4: Casting Searing Onslaught in sequence greatly increases damage dealt.
		Within 2s of using Searing Onslaught, casting the next Searing Onslaught in the combo deals 40% additional DMG. This effect lasts for the next 2s.

	C6: After casting Searing Onslaught, the next 2 Normal Attacks within the next 6s will have their DMG and ATK SPD increased by 30%.
		Additionally, Searing Onslaught will not interrupt the Normal Attack combo. <- what does this mean??

Checklist:
	- Frame count
	- Orb generation
	- Ascension bonus
	- Constellations

**/

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	d := diluc{}
	t, err := combat.NewTemplateChar(s, p)
	if err != nil {
		return nil, err
	}
	d.CharacterTemplate = t
	d.Energy = 60
	d.MaxEnergy = 60
	d.Profile.WeaponClass = combat.WeaponClassClaymore
	d.burstHook()

	return &d, nil
}

func (d *diluc) burstHook() {
	d.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.CharName != d.Profile.Name {
			return false
		}
		if ds.AbilType != combat.ActionTypeAttack {
			return false
		}
		if _, ok := d.S.Status["Diluc Burst"]; !ok {
			return false
		}
		ds.Element = combat.Pyro
		ds.Stats[combat.PyroP] += .2
		return false
	}, "diluc-burst-infuse", combat.PostSnapshot)
}

func (d *diluc) Attack(p map[string]interface{}) int {
	reset := false
	frames := 32 //first hit = 13 at 25fps
	delay := 10  //frames between execution and damage
	switch d.NormalCounter {
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
	frames = int(float64(frames) / (1 + d.Stats[combat.AtkSpd]))

	x := d.Snapshot("Normal", combat.ActionTypeAttack, combat.Physical)
	x.Mult = auto[d.NormalCounter][d.Profile.TalentLevel[combat.ActionTypeAttack]-1]

	d.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(x)
		s.Log.Infof("\t Diluc normal %v dealt %.0f damage", d.NormalCounter, damage)
	}, fmt.Sprintf("Diluc-Normal-%v", d.NormalCounter), delay)

	d.NormalCounter++

	//add a 75 frame attackcounter reset
	d.NormalResetTimer = 70

	if reset {
		d.NormalResetTimer = 0
		d.NormalCounter = 0
	}
	//return animation cd
	//this also depends on which hit in the chain this is
	return frames
}

func (d *diluc) Skill(p map[string]interface{}) int {
	reset := false
	frames := 32 //first hit = 13 at 25fps
	delay := 10  //frames between execution and damage
	switch d.eCounter {
	case 1:
		frames = 29 //47 - 35
		delay = 10
	case 2:
		frames = 53 //69
		delay = 15
	case 3:
		frames = 24 //79
		delay = 20
		reset = true
	}

	//apply attack speed
	frames = int(float64(frames) / (1 + d.Stats[combat.AtkSpd]))
	lvl := d.Profile.TalentLevel[combat.ActionTypeSkill] - 1
	if d.Profile.Constellation >= 3 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}

	x := d.Snapshot("Searing Onslaught", combat.ActionTypeSkill, combat.Pyro)
	x.Mult = skill[d.eCounter][lvl]

	d.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(x)
		s.Log.Infof("\t Diluc searing onslaught hit %v dealt %.0f damage", d.eCounter, damage)
	}, fmt.Sprintf("Diluc-Skill-%v", d.eCounter), delay)

	d.eCounter++

	//add a 75 frame attackcounter reset
	d.eResetTimer = 70

	if reset {
		d.eResetTimer = 0
		d.eCounter = 0
	}
	//return animation cd
	//this also depends on which hit in the chain this is
	return frames
}

func (d *diluc) Burst(p map[string]interface{}) int {
	d.S.Status["Diluc Burst"] = 12 * 60

	lvl := d.Profile.TalentLevel[combat.ActionTypeBurst] - 1
	if d.Profile.Constellation >= 5 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}

	//add initial damage
	x := d.Snapshot("Dawn (Initial)", combat.ActionTypeBurst, combat.Pyro)
	x.ApplyAura = true
	x.AuraBase = combat.WeakAuraBase
	x.AuraUnits = 1
	x.Mult = burstInitial[lvl]

	d.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(x)
		s.Log.Infof("\t Diluc burst (initial) dealt %.0f damage", damage)
	}, "Diluc-Burst-Initial", 100) //guess frames

	//no idea how many dot ticks

	xFinal := x.Clone()
	xFinal.Mult = burstExplode[lvl]
	d.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(xFinal)
		s.Log.Infof("\t Diluc burst (initial) dealt %.0f damage", damage)
	}, "Diluc-Burst-Initial", 220) //guess frames

	return 120
}

func (d *diluc) Tick() {
	d.CharacterTemplate.Tick()

	if d.eCounter > 0 {
		if d.eResetTimer > 0 {
			d.eResetTimer--
		} else {
			d.eCounter = 0
		}
	}
	d.Tags["Normal"] = d.NormalCounter
	d.Tags["Skill"] = d.eCounter
}
