package viridescent

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("Viridescent Venerer", set)
}

func set(c combat.Character, s *combat.Sim, count int) {
	//make use of closure to keep track of how many hits
	if count >= 2 {
		m := make(map[combat.StatType]float64)
		m[combat.AnemoP] = 0.15
		c.AddMod("viridescent venerer 2pc", m)
	}
	//increase swirl dmg by 60%, decrease elemental res to element by 40%
	if count >= 4 {
		s.AddSnapshotHook(func(snap *combat.Snapshot) bool {
			if snap.Actor != c.Name() {
				return false
			}

			if !snap.WillReact {
				return false
			}

			//decrease res
			switch snap.ReactionType {
			case combat.SwirlCryo:
				s.Target.AddResMod("VV4PCCryo", combat.ResistMod{
					Duration: 600, //10 seconds
					Ele:      combat.Cryo,
					Value:    -0.4,
				})
			case combat.SwirlElectro:
				s.Target.AddResMod("VV4PCElectro", combat.ResistMod{
					Duration: 600, //10 seconds
					Ele:      combat.Electro,
					Value:    -0.4,
				})
			case combat.SwirlPyro:
				s.Target.AddResMod("VV4PCPyro", combat.ResistMod{
					Duration: 600, //10 seconds
					Ele:      combat.Pyro,
					Value:    -0.4,
				})
			case combat.SwirlHydro:
				s.Target.AddResMod("VV4PCHydro", combat.ResistMod{
					Duration: 600, //10 seconds
					Ele:      combat.Hydro,
					Value:    -0.4,
				})
			default:
				return false
			}

			//increase damage
			snap.ReactBonus += 0.6

			return false
		}, "viridescent venerer 4pc", combat.PreReaction)
	}
	//add flat stat to char
}
