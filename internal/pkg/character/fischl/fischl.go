package fischl

import (
	"fmt"

	"github.com/srliao/gisim/internal/pkg/character/common"
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterCharFunc("Fischl", NewChar)
}

type fischl struct {
	*common.TemplateChar
}

func NewChar(s *combat.Sim, p combat.CharacterProfile) (combat.Character, error) {
	f := fischl{}
	t, err := common.New(s, p)
	if err != nil {
		return nil, err
	}
	f.TemplateChar = t
	f.Energy = 60
	f.MaxEnergy = 60

	return &f, nil
}

//42
func (f *fischl) Skill() int {
	if _, ok := f.CD["skill-cd"]; ok {
		f.S.Log.Debugf("\tFischl skill still on CD; skipping")
		return 0
	}

	d := f.Snapshot(combat.Electro)
	d.Abil = "Oz"
	d.AbilType = combat.ActionTypeSkill
	lvl := f.Profile.TalentLevel[combat.ActionTypeSkill] - 1
	if f.Profile.Constellation >= 5 {
		lvl += 3
		if lvl > 14 {
			lvl = 14
		}
	}
	d.Mult = birdSum[lvl]
	d.ApplyAura = true
	d.AuraGauge = 1
	d.AuraDecayRate = "A"

	//apply initial damage
	f.S.AddAction(func(s *combat.Sim) bool {
		damage := s.ApplyDamage(d)
		s.Log.Infof("[%v]: Fischl (Oz - summon) dealt %.0f damage", s.Frame(), damage)
		return true
	}, fmt.Sprintf("%v-Fischl-SKill", f.S.Frame()))

	//apply hit every 50 frames thereafter
	//NOT ENTIRELY ACCURATE BUT OH WELL
	tick := 0
	next := 40 + 50
	count := 0
	b := d.Clone()
	b.Mult = birdAtk[lvl]
	f.S.AddAction(func(s *combat.Sim) bool {
		tick++
		if tick < next {
			return false
		}
		damage := s.ApplyDamage(b)
		s.Log.Infof("[%v]: Fischl (Oz - summon) dealt %.0f damage", s.Frame(), damage)
		count++
		return count >= 11
	}, fmt.Sprintf("%v-Fischl-SKill", f.S.Frame()))

	f.CD["skill-cd"] = 25 * 60
	//return animation cd
	return 40
}

//first hit 40+50
//next + 50 at150
//last @ 620
