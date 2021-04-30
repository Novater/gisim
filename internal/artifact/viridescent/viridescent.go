package viridescent

import (
	"github.com/srliao/gisim/pkg/combat"
)

func init() {
	combat.RegisterSetFunc("viridescent venerer", set)
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
		s.AddSnapshotHook(func(ds *combat.Snapshot) bool {
			if ds.Actor != c.Name() {
				return false
			}
			//decrease res
			switch s.GlobalFlags.ReactionType {
			case combat.SwirlCryo:
				s.Target.AddResMod("vvcryo", combat.ResistMod{
					Duration: 600, //10 seconds
					Ele:      combat.Cryo,
					Value:    -0.4,
				})
				s.Log.Debugf("\t vv 4 pc triggered - cryo")
			case combat.SwirlElectro:
				s.Target.AddResMod("vvelectro", combat.ResistMod{
					Duration: 600, //10 seconds
					Ele:      combat.Electro,
					Value:    -0.4,
				})
				s.Log.Debugf("\t vv 4 pc triggered - electro")
			case combat.SwirlPyro:
				s.Target.AddResMod("vvpyro", combat.ResistMod{
					Duration: 600, //10 seconds
					Ele:      combat.Pyro,
					Value:    -0.4,
				})
				s.Log.Debugf("\t vv 4 pc triggered - pyro")
			case combat.SwirlHydro:
				s.Target.AddResMod("vvhydro", combat.ResistMod{
					Duration: 600, //10 seconds
					Ele:      combat.Hydro,
					Value:    -0.4,
				})
				s.Log.Debugf("\t vv 4 pc triggered - hydro")
			default:
				return false
			}

			//increase damage
			ds.ReactBonus += 0.6

			return false
		}, "viridescent venerer 4pc", combat.PostReaction)
	}
	//add flat stat to char
}
