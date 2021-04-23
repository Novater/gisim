package noelle

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("noelle", NewChar)
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
	c.Energy = 60
	c.MaxEnergy = 60
	c.Weapon.Class = combat.WeaponClassClaymore

	c.burstHook()
	c.a4()

	return &c, nil
}

/**

a2: shielding if fall below hp threshold, not implemented

a4: every 4 hit decrease breastplate cd by 1; implement as hook

c1: 100% healing, not implemented

c2: decrease stam consumption, to be implemented

c4: explodes for 400% when expired or destroyed; how to implement expired?

c6: sweeping time increase additional 50%; add 1s up to 10s everytime opponent killed (NOT IMPLEMENTED, NOTHING DIES)

**/

func (c *char) a4() {
	counter := 0
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor != c.Base.Name {
			return false
		}
		if ds.AbilType != combat.ActionAttack && ds.AbilType != combat.ActionCharge {
			return false
		}
		//we're only single target so no need to worry about multiple target hit by same attack
		counter++
		if counter == 4 {
			counter = 0
			//reduce cd if currently on cd
			if c.CD[combat.SkillICD] > c.S.F {
				c.CD[combat.SkillICD] -= 60
			}
		}
		return false
	}, "noelle a4", combat.PostDamageHook)
}

func (c *char) Attack(p int) int {

	frames := []int{1, 2, 3, 4} //TODO: frames
	delay := []int{0, 0, 0, 0}  //TODO: delay from start to when dmg is dealt

	//apply attack speed TODO: not accurate implementation
	f := int(float64(frames[c.NormalCounter]) / (1 + c.Stats[combat.AtkSpd]))

	d := c.Snapshot("Normal", combat.ActionAttack, combat.Physical, combat.WeakDurability) //TODO: noelle auto gauge
	d.Mult = attack[c.NormalCounter][c.TalentLvlAttack()]

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t %v normal %v dealt %.0f damage [%v]", c.Base.Name, c.NormalCounter+1, damage, str)
	}, fmt.Sprintf("%v-Normal-%v", c.Base.Name, c.NormalCounter+1), delay[c.NormalCounter])

	c.NormalCounter++

	//add a 70 frame attackcounter reset TODO: this needs to be checked if same for every char
	c.NormalResetTimer = 70

	if c.NormalCounter >= len(frames)-1 {
		c.NormalResetTimer = 0
		c.NormalCounter = 0
	}
	//return animation cd
	//this also depends on which hit in the chain this is
	return f //TODO: frames
}

func (c *char) Skill(p int) int {
	if c.CD[combat.SkillCD] > c.S.F {
		c.S.Log.Debugf("\t Noelle skill still on CD; skipping")
		return 0
	}
	//TODO: figure out how to implement shields??
	//deal dmg on cast
	d := c.Snapshot("Breastplate", combat.ActionSkill, combat.Geo, combat.WeakDurability) //TODO: aura strength
	d.Mult = shieldcast[c.TalentLvlSkill()]
	d.UseDef = true //TODO: test if this is working ok

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Noelle skill on cast dealt %.0f damage [%v]", damage, str)
	}, "Noelle-Breastplate-Cast", 0) //TODO: damage delay on initial cast

	//if c4, queue up dmg on expire since no dmg taken??
	if c.Base.Cons >= 4 {
		d2 := c.Snapshot("Breastplate", combat.ActionSkill, combat.Geo, combat.WeakDurability) //TODO: aura strength
		d2.Mult = 4                                                                            //400% dmg of atk

		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d2)
			s.Log.Infof("\t Noelle skill on expire (c4) dealt %.0f damage [%v]", damage, str)
		}, "Noelle-Breastplate-Expire-C4", 720) //lasts 12 seconds
	}

	c.CD[combat.SkillCD] = c.S.F + 24
	return 1 //TODO: frame count
}

func (c *char) burstHook() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor != c.Base.Name {
			return false
		}
		if ds.AbilType != combat.ActionAttack && ds.AbilType != combat.ActionCharge {
			return false
		}
		if _, ok := c.S.Status["Noelle Burst"]; !ok {
			return false
		}
		ds.Element = combat.Geo
		//add flat attack equal to % of her def
		def := c.Base.Def*(1+ds.Stats[combat.DEFP]) + ds.Stats[combat.DEF]
		mult := defconv[c.TalentLvlBurst()]
		if c.Base.Cons == 6 {
			mult += 0.5
		}
		fa := mult * def
		c.S.Log.Debugf("\t adding %def to Noelle's attack, total def: %.2f attack added: %.2f %def: %.2f%%", def, fa, mult)
		ds.Stats[combat.ATK] += fa
		return false
	}, "noelle-burst-buff", combat.PostSnapshot)
}

func (c *char) Burst(p int) int {
	if c.CD[combat.BurstCD] > c.S.F {
		c.S.Log.Debugf("\t Noelle burst still on CD; skipping")
		return 0
	}
	c.S.Status["Noelle Burst"] = c.S.F + 900 // 15 seconds duration

	d := c.Snapshot("Sweeping Time", combat.ActionBurst, combat.Geo, combat.WeakDurability) //TODO: aura strength
	d.Mult = burst[c.TalentLvlBurst()]

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t Noelle burst on cast dealt %.0f damage [%v]", damage, str)
	}, "Noelle-Sweeping-Time-Cast", 0) //TODO: damage delay on initial cast

	d2 := c.Snapshot("Sweeping Time", combat.ActionBurst, combat.Geo, combat.WeakDurability) //TODO: aura strength
	d2.Mult = burstskill[c.TalentLvlBurst()]

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d2)
		s.Log.Infof("\t Noelle burst on cast 2nd part dealt %.0f damage [%v]", damage, str)
	}, "Noelle-Sweeping-Time-Skill", 0) //TODO: damage delay on initial cast

	c.CD[combat.BurstCD] = c.S.F + 900
	c.Energy = 0
	return 1 //TODO: frame count
}
