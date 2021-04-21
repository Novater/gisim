package beidou

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("beidou", NewChar)
}

type char struct {
	*combat.CharacterTemplate
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	c := char{}
	t, err := combat.NewTemplateChar(s, p)
	if err != nil {
		return nil, err
	}
	c.CharacterTemplate = t
	c.Energy = 80
	c.MaxEnergy = 80
	c.Weapon.Class = combat.WeaponClassClaymore

	c.burstProc()
	c.a4()

	return &c, nil
}

/**
Counterattacking with Tidecaller at the precise moment when the character is hit grants the maximum DMG Bonus.

Gain the following effects for 10s after unleashing Tidecaller with its maximum DMG Bonus:
• DMG dealt by Normal and Charged Attacks is increased by 15%. ATK SPD of Normal and Charged Attacks is increased by 15%.
• Greatly reduced delay before unleashing Charged Attacks.

c1
When Stormbreaker is used:
Creates a shield that absorbs up to 16% of Beidou's Max HP for 15s.
This shield absorbs Electro DMG 250% more effectively.

c2
Stormbreaker's arc lightning can jump to 2 additional targets.

c3
Within 10s of taking DMG, Beidou's Normal Attacks gain 20% additional Electro DMG.

c6
During the duration of Stormbreaker, the Electro RES of surrounding opponents is decreased by 15%.
**/

func (c *char) a4() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor != c.Base.Name {
			return false
		}
		if ds.AbilType != combat.ActionAttack && ds.AbilType != combat.ActionCharge {
			return false
		}
		if !c.S.StatusActive("Beidou-Perfect-Counter") {
			return false
		}
		//add 15% to dmg
		ds.DmgBonus += 0.15

		//TODO: missing 15% atkspeed

		return false
	}, "Beidou-Stormbreaker-Proc", combat.PostSnapshot)
}

func (c *char) Attack(p int) int {
	frames := []int{29, 21, 40, 45, 31}
	delay := []int{40, 40, 40, 40, 40}
	return c.CharacterTemplate.AttackHelperSingle(frames, delay, auto)
}

func (c *char) Skill(p int) int {
	if c.CD[combat.SkillCD] > c.S.F {
		c.S.Log.Debugf("\t Beidou skill still on CD; skipping")
		return 0
	}
	//0 for base dmg, 1 for 1x bonus, 2 for max bonus
	if p >= 2 {
		p = 2
		c.S.Status["Beidou-Perfect-Counter"] = c.S.F + 600
	}

	d := c.Snapshot("Tidecaller (E)", combat.ActionSkill, combat.Electro, combat.MedDurability)
	d.Mult = 1.0 + .5*float64(p) //TODO: update mult

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Beidou skill hit dealt %.0f damage [%v]", damage, str)
	}, "Beidou-Skill", 1) //TODO: frames

	//TODO: energy count
	c.S.AddEnergyParticles("Beidou", 2, combat.Electro, 100)

	//TODO: add shield

	c.CD[combat.SkillCD] = c.S.F + 450
	return 1 //TODO: frames
}

func (c *char) Burst(p int) int {

	d := c.Snapshot("Stormbreaker (Q)", combat.ActionSkill, combat.Cryo, combat.StrongDurability)
	d.Mult = 1 //TODO: multiplier

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Beidou burst hit dealt %.0f damage [%v]", damage, str)
	}, "Stormbreaker-Initial", 1) //TODO: frames

	c.S.Status["Beidou-Stormbreaker"] = c.S.F + 900

	if c.Base.Cons == 6 {
		//reduce electro res by 10%
		c.S.Target.AddResMod("Beidou-C6", combat.ResistMod{
			Duration: 900, //10 seconds
			Ele:      combat.Electro,
			Value:    -0.1,
		})
	}

	c.Energy = 0
	c.CD[combat.BurstCD] = c.S.F + 1200
	return 1 //TODO: frames
}

func (c *char) burstProc() {
	icd := 0
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.AbilType != combat.ActionAttack && ds.AbilType != combat.ActionCharge {
			return false
		}
		if icd > c.S.F {
			return false
		}
		//TODO: not sure if this snapshots stats on Q cast or not
		d := c.Snapshot("Stormbreaker Proc (Q)", combat.ActionSkill, combat.Cryo, combat.StrongDurability)
		d.Mult = 1 //TODO: multiplier

		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Beidou stormbreaker proc dealt %.0f damage [%v]", damage, str)
		}, "Stormbreaker-Proc", 1) //TODO: frames

		icd = c.S.F + 60 // once per second

		return false
	}, "Beidou-Stormbreaker-Proc", combat.PostDamageHook)
}
