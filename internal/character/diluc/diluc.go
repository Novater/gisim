package diluc

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("diluc", NewChar)
}

type diluc struct {
	*combat.CharacterTemplate
	eStarted    bool
	eStartFrame int
	eLastUse    int
	eCounter    int
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
	- infused pyro app (wrongly done)
	- Orb generation (guessed 10% chance at 2)
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
	d.Energy = 40
	d.MaxEnergy = 40
	d.Weapon.Class = combat.WeaponClassClaymore
	d.burstHook()

	return &d, nil
}

func (d *diluc) burstHook() {
	d.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor != d.Base.Name {
			return false
		}
		if ds.AbilType != combat.ActionAttack {
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

func (d *diluc) Attack(p int) int {
	x := d.Snapshot("Normal", combat.ActionAttack, combat.Physical, combat.WeakDurability)
	x.Mult = auto[d.NormalCounter][d.TalentLvlAttack()]

	reset := false
	frames := 38
	delay := 23 //frames between execution and damage
	switch d.NormalCounter {
	case 1:
		frames = 52 //90-38
		delay = 37  //75-38
	case 2:
		frames = 40 //130-90
		delay = 25  //115-90
		x.Durability = 0
	case 3:
		frames = 64 //179-115
		delay = 62  //177-115
		reset = true
	}

	//apply attack speed
	frames = int(float64(frames) / (1 + d.Stats[combat.AtkSpd]))

	d.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(x)
		s.Log.Infof("\t Diluc normal %v dealt %.0f damage [%v]", d.NormalCounter, damage, str)
	}, fmt.Sprintf("Diluc-Normal-%v", d.NormalCounter), delay)

	d.NormalCounter++

	//add a 75 frame attackcounter reset
	d.NormalResetTimer = 70

	if reset {
		d.NormalResetTimer = 0
		d.NormalCounter = 0
	}

	return frames
}

func (d *diluc) Skill(p int) int {
	var frames, delay int
	switch d.eCounter {
	case 0:
		frames = 40
		delay = 40
		//start ticking 10 seconds total cd
		d.eStarted = true
		d.eStartFrame = d.S.F
		d.eLastUse = d.S.F
	case 1:
		frames = 50 //47 - 35
		delay = 45
		d.eLastUse = d.S.F
	case 2:
		frames = 67 //69
		delay = 60
	}
	orb := 1
	if d.S.Rand.Float64() < 0.1 {
		orb = 2
	}
	d.S.AddEnergyParticles("Diluc", orb, combat.Pyro, delay+60)

	//actual skill cd starts immediately on first cast
	//times out after 4 seconds of not using
	//every hit applies pyro
	//apply attack speed
	frames = int(float64(frames) / (1 + d.Stats[combat.AtkSpd]))

	x := d.Snapshot("Searing Onslaught", combat.ActionSkill, combat.Pyro, combat.WeakDurability)
	x.Mult = skill[d.eCounter][d.TalentLvlSkill()]

	d.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(x)
		s.Log.Infof("\t Diluc searing onslaught hit %v dealt %.0f damage [%v]", d.eCounter, damage, str)
	}, fmt.Sprintf("Diluc-Skill-%v", d.eCounter), delay)

	d.eCounter++
	if d.eCounter == 3 {
		//ability can go on cd now
		cd := 600 - (d.S.F - d.eStartFrame)
		d.S.Log.Debugf("\t Diluc skill going on cd for %v", cd)
		d.CD[combat.BurstCD] = d.S.F + cd
		d.eStarted = false
		d.eStartFrame = -1
		d.eLastUse = -1
		d.eCounter = 0
	}
	//return animation cd
	//this also depends on which hit in the chain this is
	return frames
}

func (d *diluc) Burst(p int) int {
	d.S.Status["Diluc Burst"] = 12 * 60

	//add initial damage
	x := d.Snapshot("Dawn (Initial)", combat.ActionBurst, combat.Pyro, combat.StrongDurability)
	x.Mult = burstInitial[d.TalentLvlBurst()]

	d.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(x)
		s.Log.Infof("\t Diluc burst (initial) dealt %.0f damage [%v]", damage, str)
	}, "Diluc-Burst-Initial", 100) //roughly 100

	//dot does 7 hits + explosion, roughly every 13 frame? blows up at 210 frames
	//first tick did 50 dur as well?
	for i := 1; i <= 7; i++ {
		xd := x.Clone()
		xd.Durability = 0
		xd.Mult = burstDOT[d.TalentLvlBurst()]
		//hit 5 applies pyro again, no idea why 5
		if i == 5 {
			xd.Durability = combat.MedDurability
		}
		hit := i

		d.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(x)
			s.Log.Infof("\t Diluc burst (dot %v) dealt %.0f damage [%v]", hit, damage, str)
		}, fmt.Sprintf("Diluc-Burst-Dot-%v", hit), 100+i*13)
	}

	xFinal := x.Clone()
	xFinal.Mult = burstExplode[d.TalentLvlBurst()]
	d.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(xFinal)
		s.Log.Infof("\t Diluc burst (explode) dealt %.0f damage [%v]", damage, str)
	}, "Diluc-Burst-Initial", 210)

	return 170
}

func (d *diluc) Tick() {
	d.CharacterTemplate.Tick()

	if d.eStarted {
		//check if 4 second has passed since last use
		if d.S.F-d.eLastUse >= 240 {
			//if so, set ability to be on cd equal to 10s less started
			cd := 600 - (d.S.F - d.eStartFrame)
			d.S.Log.Debugf("\t Diluc skill expired, going on cd for %v, last executed %v", cd, d.eLastUse)
			d.CD[combat.BurstCD] = d.S.F + cd
			//reset
			d.eStarted = false
			d.eStartFrame = -1
			d.eLastUse = -1
			d.eCounter = 0
		}
	}
}
