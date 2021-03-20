package ganyu

import (
	"fmt"

	"github.com/srliao/gisim/internal/pkg/character/common"
	"github.com/srliao/gisim/internal/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Ganyu", NewChar)
}

type ganyu struct {
	*common.TemplateChar
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	g := ganyu{}
	t, err := common.New(s, p)
	if err != nil {
		return nil, err
	}
	g.TemplateChar = t

	if g.Profile.Constellation >= 1 {
		s.Log.Debugf("\tactivating Ganyu C1")

		s.AddEffect(func(snap *combat.Snapshot) bool {
			//check if c1 debuff is on, if so, reduce resist by -0.15
			if _, ok := s.Target.Status["ganyu-c1"]; ok {
				s.Log.Debugf("\t[%v]: applying Ganyu C1 cryo debuff", combat.PrintFrames(s.Frame))
				snap.ResMod[combat.Cryo] -= 0.15
			}
			return false
		}, "ganyu-c1", combat.PreDamageHook)

		s.AddEffect(func(snap *combat.Snapshot) bool {
			if snap.CharName == "Ganyu" && snap.Abil == "Frost Flake Arrow" {
				//if c1, increase character energy by 2, unaffected by ER; assume assuming arrow always hits here
				g.Energy += 2
				if g.Energy > g.MaxEnergy {
					g.Energy = g.MaxEnergy
				}
				s.Log.Debugf("\t[%v]: Ganyu C1 refunding 2 energy; current energy %v", combat.PrintFrames(s.Frame), g.Energy)
				//also add c1 debuff to target
				s.Target.Status["ganyu-c1"] = 5 * 60
			}
			return false
		}, "ganyu-c1", combat.PostDamageHook)
	}

	return &g, nil
}

func (g *ganyu) ChargeAttack() int {
	i := 0
	g.S.AddAction(func(s *combat.Sim) bool {
		if i < 20 { //assume 20 frame travel time
			i++
			return false
		}
		//abil
		d := g.Snapshot(combat.Cryo)
		d.Abil = "Frost Flake Arrow"
		d.AbilType = combat.ActionTypeChargedAttack
		d.HitWeakPoint = true
		d.Mult = ffa[g.Profile.TalentLevel[combat.ActionTypeAttack]-1]
		d.AuraGauge = 1
		d.AuraUnit = "A"
		d.ApplyAura = true
		//if not ICD, apply aura
		if _, ok := g.CD["ICD-charge"]; !ok {
			d.ApplyAura = true
		}
		//check if A4 talent is
		if _, ok := g.CD["A2"]; ok {
			d.Stats[combat.CR] += 0.2
		}
		g.CD["A2"] = 5 * 60
		//apply damage
		damage := s.ApplyDamage(d)
		s.Log.Infof("[%v]: Ganyu frost arrow dealt %.0f damage", combat.PrintFrames(s.Frame), damage)
		return true
	}, fmt.Sprintf("%v-Ganyu-CA-FFA", g.S.Frame))

	b := 0
	g.S.AddAction(func(s *combat.Sim) bool {
		if b < 50 {
			b++
			return false
		}
		//abil
		d := g.Snapshot(combat.Cryo)
		d.Abil = "Frost Flake Bloom"
		d.AbilType = combat.ActionTypeChargedAttack
		d.Mult = ffb[g.Profile.TalentLevel[combat.ActionTypeAttack]-1]
		d.ApplyAura = true
		d.AuraGauge = 1
		d.AuraUnit = "A"
		//if not ICD, apply aura
		if _, ok := g.CD["ICD-charge"]; !ok {
			d.ApplyAura = true
		}
		if _, ok := g.CD["A2"]; ok {
			d.Stats[combat.CR] += 0.2
		}
		//apply damage
		damage := s.ApplyDamage(d)
		s.Log.Infof("[%v]: Ganyu frost flake bloom dealt %.0f damage", combat.PrintFrames(s.Frame), damage)
		return true
	}, fmt.Sprintf("%v-Ganyu-CA-FFB", g.S.Frame))

	return 137
}

func (g *ganyu) Skill() int {
	//if c2, check if either cd is cooldown
	charge := ""
	_, c2ok := g.CD["skill-cd-2"]
	_, ok := g.CD[charge]

	if g.Profile.Constellation >= 2 {
		if !c2ok {
			charge = "skill-cd-2"
		}
	}
	if !ok {
		charge = "skill-cd"
	}

	if charge == "" {
		g.S.Log.Debugf("\tGanyu skill still on CD; skipping")
		return 0
	}

	//snap shot stats at cast time here
	d := g.Snapshot(combat.Cryo)
	d.Abil = "Ice Lotus"
	d.AbilType = combat.ActionTypeSkill
	lvl := g.Profile.TalentLevel[combat.ActionTypeSkill] - 1
	if g.Profile.Constellation >= 5 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	d.Mult = lotus[lvl]
	d.ApplyAura = true
	d.AuraGauge = 1
	d.AuraUnit = "A"

	//we get the orbs right away
	//add delayed orb for travel time
	orbDelay := 0
	g.S.AddAction(func(s *combat.Sim) bool {
		if orbDelay < 90 { //1.5 second to receive the org
			orbDelay++
			return false
		}
		s.GenerateOrb(2, combat.Cryo, false)
		return true
	}, fmt.Sprintf("%v-Ganyu-Skill-Orb", g.S.Frame))

	tick := 0
	flower := func(s *combat.Sim) bool {
		if tick < 6*60 {
			tick++
			return false
		}
		//do damage
		damage := s.ApplyDamage(d)
		s.Log.Infof("[%v]: Ganyu ice lotus (tick) dealt %.0f damage", combat.PrintFrames(s.Frame), damage)
		tick++
		return true
	}
	g.S.AddAction(flower, fmt.Sprintf("%v-Ganyu-Skill", g.S.Frame))
	//add cooldown to sim
	g.CD[charge] = 10 * 60

	return 30
}

func (g *ganyu) Burst() int {
	//check if on cd first
	if _, ok := g.CD["burst-cd"]; ok {
		g.S.Log.Debugf("\tGanyu burst still on CD; skipping")
		return 0
	}
	//check if sufficient energy
	if g.Energy < g.MaxEnergy {
		g.S.Log.Debugf("\tGanyu burst - insufficent energy, current: %v", g.Energy)
		return 0
	}
	//snap shot stats at cast time here
	d := g.Snapshot(combat.Cryo)
	d.Abil = "Celestial Shower"
	d.AbilType = combat.ActionTypeBurst
	lvl := g.Profile.TalentLevel[combat.ActionTypeBurst] - 1
	if g.Profile.Constellation >= 3 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	d.Mult = shower[lvl]
	d.ApplyAura = true
	d.AuraGauge = 1
	d.AuraUnit = "A"

	//apply weapon stats here
	//burst should be instant
	//should add a hook to the unit, triggering damage every 1 sec
	//also add a field effect
	tick := 0
	storm := func(s *combat.Sim) bool {
		if tick > 900 {
			return true
		}
		//check if multiples of 60s; also add an initial delay of 120 frames
		if tick%60 != 0 || tick < 120 {
			tick++
			return false
		}
		//do damage
		damage := s.ApplyDamage(d)
		s.Log.Infof("[%v]: Ganyu burst (tick) dealt %.0f damage", combat.PrintFrames(s.Frame), damage)
		tick++
		return false
	}
	g.S.AddAction(storm, fmt.Sprintf("%v-Ganyu-Burst", g.S.Frame))
	//add cooldown to sim
	g.CD["burst-cd"] = 15 * 60
	//use up energy
	g.Energy = 0

	return 122
}

func (g *ganyu) ActionCooldown(a combat.ActionType) int {
	switch a {
	case combat.ActionTypeBurst:
		return g.CD["burst-cd"]
	case combat.ActionTypeSkill:
		cd := g.CD["skill-cd"]
		if g.Profile.Constellation >= 2 {
			cd2 := g.CD["skill-cd2"]
			if cd2 < cd {
				return cd2
			}
		}
		return cd
	}
	return 0

}
