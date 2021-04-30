package sucrose

import (
	"fmt"

	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("sucrose", NewChar)
}

type char struct {
	*combat.CharacterTemplate
	c6ele combat.EleType
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
	c.Weapon.Class = combat.WeaponClassCatalyst

	c.a2()
	c.a4()
	c.absorb()

	if c.Base.Cons >= 1 {
		c.Tags["last"] = -1
	}

	if c.Base.Cons == 6 {
		c.c6()
	}

	return &c, nil
}

func (c *char) ActionFrames(a combat.ActionType, p int) int {
	switch a {
	case combat.ActionAttack:
		f := 0
		switch c.NormalCounter {
		//TODO: need to add atkspd mod
		case 0:
			f = 19 //frames from keqing lib
		case 1:
			f = 38
		case 2:
			f = 70
		case 3:
			f = 101
		}
		f = int(float64(f) / (1 + c.Stats[combat.AtkSpd]))
		return f
	case combat.ActionCharge:
		return 53 //frames from keqing lib
	case combat.ActionSkill:
		return 55 //ok
	case combat.ActionBurst:
		return 46 //ok
	default:
		c.S.Log.Warnf("%v: unknown action, frames invalid", a)
		return 0
	}
}

func (c *char) a2() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if c.Tags["a2-pyro"] >= c.S.F && ds.ActorEle == combat.Pyro {
			ds.Stats[combat.EM] += 50
			return false
		}
		if c.Tags["a2-cryo"] >= c.S.F && ds.ActorEle == combat.Cryo {
			ds.Stats[combat.EM] += 50
			return false
		}
		if c.Tags["a2-hydro"] >= c.S.F && ds.ActorEle == combat.Hydro {
			ds.Stats[combat.EM] += 50
			return false
		}
		if c.Tags["a2-electro"] >= c.S.F && ds.ActorEle == combat.Electro {
			ds.Stats[combat.EM] += 50
			return false
		}
		return false
	}, "sucrose-a2-buff", combat.PostSnapshot)

	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor != c.Base.Name {
			return false
		}
		switch c.S.GlobalFlags.ReactionType {
		case combat.SwirlCryo:
			c.Tags["a2-cryo"] = c.S.F + 480
		case combat.SwirlElectro:
			c.Tags["a2-electro"] = c.S.F + 480
		case combat.SwirlHydro:
			c.Tags["a2-hydro"] = c.S.F + 480
		case combat.SwirlPyro:
			c.Tags["a2-pyro"] = c.S.F + 480
		}
		return false
	}, "sucrose-a2-buff", combat.PostReaction)
}

func (c *char) a4() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if ds.Actor == c.Base.Name {
			return false
		}
		if c.Tags["a4"] >= c.S.F {
			ds.Stats[combat.EM] += c.Stats[combat.EM] * 0.2
		}
		return false
	}, "sucrose-a2-buff", combat.PostSnapshot)
}

func (c *char) Attack(p int) int {
	d := c.Snapshot("Normal", combat.ActionAttack, combat.Anemo, 0)
	d.Mult = attack[c.NormalCounter][c.TalentLvlAttack()]
	f := c.ActionFrames(combat.ActionAttack, p)
	delay := f - 5 //TODO: frames

	//apply every 3rd hit or every 2.5s
	//TODO: check if this code works
	if c.CD[combat.NormalICD] <= c.S.F {
		c.Tags[combat.NormalICD] = 1
		c.CD[combat.NormalICD] = c.S.F + 150 //2.5s
		d.Durability = combat.WeakDurability
	} else if c.Tags[combat.NormalICD] == 3 {
		c.Tags[combat.NormalICD] = 1
		d.Durability = combat.WeakDurability
	} else {
		c.Tags[combat.NormalICD]++
	}

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t %v normal %v dealt %.0f damage, dur %v [%v]", c.Base.Name, c.NormalCounter, damage, d.Durability, str)
	}, fmt.Sprintf("%v-Normal-%v", c.Base.Name, c.NormalCounter), delay)

	c.NormalResetTimer = 150
	c.NormalCounter++
	if c.NormalCounter == 4 {
		c.NormalCounter = 0
		c.NormalResetTimer = 0
	}

	if c.Base.Cons >= 4 {
		count := c.Tags["c4"]
		count++
		if count == 7 {
			if c.CD[combat.SkillCD] > c.S.F {
				n := c.S.Rand.Intn(7) + 1
				c.CD[combat.SkillCD] -= n * 60
			}
			count = 0
		}
		c.Tags["c4"] = count
	}

	return f
}

func (c *char) ChargeAttack(p int) int {
	d := c.Snapshot("Charge", combat.ActionCharge, combat.Anemo, 0)
	d.Mult = charge[c.TalentLvlAttack()]

	//apply every 3rd hit or every 2.5s
	//TODO: check if this code works
	if c.CD[combat.ChargedICD] <= c.S.F {
		c.Tags[combat.ChargedICD] = 1
		c.CD[combat.ChargedICD] = c.S.F + 150 //2.5s
		d.Durability = combat.WeakDurability
	} else if c.Tags[combat.ChargedICD] == 3 {
		c.Tags[combat.ChargedICD] = 1
		d.Durability = combat.WeakDurability
	} else {
		c.Tags[combat.ChargedICD]++
	}

	c.S.AddTask(func(s *combat.Sim) {
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t %v charge attack dealt %.0f damage [%v]", c.Base.Name, damage, str)
	}, fmt.Sprintf("%v-ChargeAttack", c.Base.Name), 50)

	if c.Base.Cons >= 4 {
		count := c.Tags["c4"]
		count++
		if count == 7 {
			if c.CD[combat.SkillCD] > c.S.F {
				n := c.S.Rand.Intn(7) + 1
				c.CD[combat.SkillCD] -= n * 60
			}
			count = 0
		}
		c.Tags["c4"] = count
	}

	return c.ActionFrames(combat.ActionCharge, p)
}

func (c *char) Skill(p int) int {
	if c.CD[combat.SkillCD] > c.S.F {
		c.S.Log.Debugf("\t Sucrose skill still on CD; skipping")
		return 0
	}
	//41 frame delay
	d := c.Snapshot("Astable Anemohypostasis Creation-6308", combat.ActionSkill, combat.Anemo, combat.WeakDurability)
	d.Mult = skill[c.TalentLvlSkill()]

	c.S.AddTask(func(s *combat.Sim) {
		c.Tags["a4"] = s.F + 480
		damage, str := s.ApplyDamage(d)
		s.Log.Infof("\t %v skill dealt %.0f damage [%v]", c.Base.Name, damage, str)
	}, "Sucrose - Skill", 41)

	c.S.AddEnergyParticles("sucrose", 4, combat.Anemo, 150)

	if c.Base.Cons >= 1 {
		last := c.Tags["last"]
		//we can only be here if the cooldown is up, meaning at least 1 charge is off cooldown
		//last should just represent when the next charge starts recharging, this should equal
		//to right when the first charge is off cooldown
		if last == -1 {
			c.Tags["last"] = c.S.F
			c.S.Log.Infof("\t Sucrose first time using skill, first charge cd up at %v", c.S.F+900)
		} else if c.S.F-last < 900 {
			//if last is less than 15s in the past, then 1 charge is up
			//then we move last up to when the first charge goes off CD\
			c.S.Log.Infof("\t Sucrose last diff %v", c.S.F-last)
			c.Tags["last"] = last + 900
			c.CD[combat.SkillCD] = last + 900
			c.S.Log.Infof("\t Sucrose skill going on CD until %v, last = %v", last+900, c.Tags["last"])
		} else {
			//so if last is more than 15s in the past, then both charges must be up
			//so then the charge restarts now
			c.Tags["last"] = c.S.F
			c.S.Log.Infof("\t Sucrose charge cd starts at %v", c.S.F)
		}

	} else {
		c.CD[combat.SkillCD] = c.S.F + 900
	}

	return c.ActionFrames(combat.ActionSkill, p)
}

func (c *char) Burst(p int) int {
	//tag a4
	//3 hits, 135, 249, 368; all 3 applied swirl; c2 i guess adds 2 second so one more hit
	//let's just assume 120, 240, 360, 480
	if c.CD[combat.BurstCD] > c.S.F {
		c.S.Log.Debugf("\t Sucrose burst still on CD; skipping")
		return 0
	}

	count := 360
	if c.Base.Cons >= 2 {
		count = 480
	}

	c.Tags["absorb"] = 0
	c.S.Status["sucrose q"] = c.S.F + count

	for i := 120; i <= count; i += 120 {
		d := c.Snapshot("Forbidden Creation-Isomer 75/Type II", combat.ActionBurst, combat.Anemo, combat.WeakDurability)
		d.Mult = burstDot[c.TalentLvlBurst()]
		t := i + 1

		c.S.AddTask(func(s *combat.Sim) {
			c.Tags["a4"] = s.F + 480
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Sucrose burst hit #%v dealt %.0f damage [%v]", t, damage, str)
		}, fmt.Sprintf("Sucrose - Burst #%v", t), i)
	}

	c.CD[combat.BurstCD] = c.S.F + 1200
	c.Energy = 0
	return 0
}

func (c *char) absorb() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if !c.S.StatusActive("sucrose q") {
			return false
		}
		if ds.Element == combat.Anemo || ds.Element == combat.Geo {
			return false //can't infuse anemo or geo
		}
		//only do it once
		if c.Tags["absorb"] > 0 {
			return false
		}
		c.Tags["absorb"] = 1
		//other wise do dmg
		d := c.Snapshot("Forbidden Creation-Isomer 75/Type II (Absorb)", combat.ActionBurst, ds.Element, combat.WeakDurability)
		d.Mult = burstAbsorb[c.TalentLvlBurst()]
		c.S.AddTask(func(s *combat.Sim) {
			damage, str := s.ApplyDamage(d)
			s.Log.Infof("\t Sucrose burst (absorb) dealt %.0f damage [%v]", damage, str)
		}, "Sucrose - Burst (Absorb)", 1)

		//If Forbidden Creation-Isomer 75/Type II triggers an Elemental Absorption, all part members
		//gain a 20% Elemental DMG Bonus for the corresponding absorbed elemental during its duration.
		if c.Base.Cons == 6 {
			c.S.Status["sucrose c6"] = c.S.Status["sucrose q"]
			c.c6ele = ds.Element
		}

		return false
	}, "sucrose-absorb", combat.PostDamageHook)
}

func (c *char) c6() {
	c.S.AddSnapshotHook(func(ds *combat.Snapshot) bool {
		if !c.S.StatusActive("sucrose c6") {
			return false
		}
		//If Forbidden Creation-Isomer 75/Type II triggers an Elemental Absorption, all part members
		//gain a 20% Elemental DMG Bonus for the corresponding absorbed elemental during its duration.
		p := combat.EleToDmgP(c.c6ele)
		ds.Stats[p] += 0.2
		return false
	}, "sucrose-c6", combat.PostSnapshot)
}

func (c *char) ResetActionCooldown(a combat.ActionType) {
	//we're overriding this b/c of the c1 charges
	switch a {
	case combat.ActionBurst:
		delete(c.CD, combat.BurstCD)
	case combat.ActionSkill:
		if c.Base.Cons == 0 {
			delete(c.CD, combat.SkillCD)
			return
		}
		//ok here's the fun part...
		//if last is more than 15s away from current frame then both charges are up, do nothing
		if c.S.F-c.Tags["last"] > 900 || c.Tags["last"] == 0 {
			return
		}
		//otherwise move CD and restart charging last now
		c.Tags["last"] = c.S.F
		c.CD[combat.SkillCD] = c.S.F

	}
}

func (c *char) ActionStam(a combat.ActionType, p int) float64 {
	switch a {
	case combat.ActionDash:
		return 15
	case combat.ActionCharge:
		return 50
	default:
		return 0
	}
}
