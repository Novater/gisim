package ganyu

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Ganyu", NewChar)
}

type ganyu struct {
	*combat.CharacterTemplate
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	g := ganyu{}
	t, err := combat.NewTemplateChar(s, p)
	if err != nil {
		return nil, err
	}
	g.CharacterTemplate = t
	g.Energy = 60
	g.MaxEnergy = 60
	g.Profile.WeaponClass = combat.WeaponClassBow

	if g.Profile.Constellation >= 1 {
		g.c1()
	}

	return &g, nil
}

func (g *ganyu) c1() {
	s := g.S
	s.Log.Debugf("\tactivating Ganyu C1")

	s.AddSnapshotHook(func(snap *combat.Snapshot) bool {
		if snap.CharName == "Ganyu" && snap.Abil == "Frost Flake Arrow" {
			//if c1, increase character energy by 2, unaffected by ER; assume assuming arrow always hits here
			g.Energy += 2
			if g.Energy > g.MaxEnergy {
				g.Energy = g.MaxEnergy
			}
			s.Log.Debugf("\t[%v]: Ganyu C1 refunding 2 energy; current energy %v", s.Frame(), g.Energy)
			//also add c1 debuff to target
			s.Target.AddResMod("ganyu-c1", combat.ResistMod{
				Ele:      combat.Cryo,
				Value:    -0.15,
				Duration: 5 * 60,
			})
		}
		return false
	}, "ganyu-c1", combat.PostDamageHook)
}

func (g *ganyu) Aimed(p map[string]interface{}) int {
	f := g.Snapshot("Frost Flake Arrow", combat.ActionTypeAimedShot, combat.Cryo)
	f.HitWeakPoint = true
	f.Mult = ffa[g.Profile.TalentLevel[combat.ActionTypeAttack]-1]
	f.AuraBase = combat.WeakAuraBase
	f.AuraUnits = 1
	f.ApplyAura = true

	b := g.Snapshot("Frost Flake Bloom", combat.ActionTypeAimedShot, combat.Cryo)
	b.Mult = ffb[g.Profile.TalentLevel[combat.ActionTypeAttack]-1]
	b.ApplyAura = true
	b.AuraBase = combat.WeakAuraBase
	b.AuraUnits = 1

	a2 := g.CD["A2"]
	if a2 > g.S.F {
		f.Stats[combat.CR] += 0.2
		b.Stats[combat.CR] += 0.2
	}

	g.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(f)
		s.Log.Infof("\t Ganyu frost arrow dealt %.0f damage", damage)
		//apply A2 on hit
		g.CD["A2"] = g.S.F + 5*60
	}, "Ganyu-Aimed-FFA", 20+137)

	g.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(b)
		s.Log.Infof("\t Ganyu frost flake bloom dealt %.0f damage", damage)
		//apply A2 on hit
		g.CD["A2"] = g.S.F + 5*60
	}, "Ganyu-Aimed-FFB", 20+20+137)

	return 137
}

func (g *ganyu) Skill(p map[string]interface{}) int {
	//if c2, check if either cd is cooldown
	charge := ""
	c2onCD := g.CD["skill-cd-2"] > g.S.F
	onCD := g.CD[charge] > g.S.F

	if g.Profile.Constellation >= 2 {
		if !c2onCD {
			charge = "skill-cd-2"
		}
	}
	if !onCD {
		charge = "skill-cd"
	}

	if charge == "" {
		g.S.Log.Debugf("\tGanyu skill still on CD; skipping")
		return 0
	}

	//snap shot stats at cast time here
	d := g.Snapshot("Ice Lotus", combat.ActionTypeSkill, combat.Cryo)
	lvl := g.Profile.TalentLevel[combat.ActionTypeSkill] - 1
	if g.Profile.Constellation >= 5 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	d.Mult = lotus[lvl]
	d.ApplyAura = true
	d.AuraBase = combat.WeakAuraBase
	d.AuraUnits = 1

	//we get the orbs right away
	g.S.AddEnergyParticles("Ganyu", 2, combat.Cryo, 90) //90s travel time

	//flower damage is after 6 seconds
	g.S.AddTask(func(s *combat.Sim) {
		damage := s.ApplyDamage(d)
		s.Log.Infof("\t Ganyu ice lotus dealt %.0f damage", damage)
	}, "Ganyu Flower", 6*60)

	//add cooldown to sim
	g.CD[charge] = g.S.F + 10*60

	return 30
}

func (g *ganyu) Burst(p map[string]interface{}) int {
	//check if on cd first
	if g.CD[combat.BurstCD] > g.S.F {
		g.S.Log.Debugf("\tGanyu burst still on CD; skipping")
		return 0
	}
	//check if sufficient energy
	if g.Energy < g.MaxEnergy {
		g.S.Log.Debugf("\tGanyu burst - insufficent energy, current: %v", g.Energy)
		return 0
	}
	//snap shot stats at cast time here
	d := g.Snapshot("Celestial Shower", combat.ActionTypeBurst, combat.Cryo)
	lvl := g.Profile.TalentLevel[combat.ActionTypeBurst] - 1
	if g.Profile.Constellation >= 3 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	d.Mult = shower[lvl]
	d.ApplyAura = true
	d.AuraBase = combat.WeakAuraBase
	d.AuraUnits = 1

	for delay := 120; delay <= 900; delay += 60 {
		g.S.AddTask(func(s *combat.Sim) {
			damage := s.ApplyDamage(d)
			s.Log.Infof("\t Ganyu burst (tick) dealt %.0f damage", s.Frame(), damage)
		}, "Ganyu Burst", delay)
	}

	//add cooldown to sim
	g.CD[combat.BurstCD] = g.S.F + 15*60
	//use up energy
	g.Energy = 0

	return 122
}

func (g *ganyu) ActionReady(a combat.ActionType) bool {
	switch a {
	case combat.ActionTypeBurst:
		if g.Energy != g.MaxEnergy {
			return false
		}
		return g.CD[combat.BurstCD] <= g.S.F
	case combat.ActionTypeSkill:
		skillReady := g.CD[combat.SkillCD] <= g.S.F
		//if skill ready return true regardless of c2
		if skillReady {
			return true
		}
		//other wise skill-cd is there, we check c2
		if g.Profile.Constellation >= 2 {
			return g.CD["skill-cd2"] <= g.S.F
		}
		return false
	}
	return true
}

func (g *ganyu) Tick() {
	g.CharacterTemplate.Tick()
}
