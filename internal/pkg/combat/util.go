package combat

import (
	"fmt"

	"go.uber.org/zap"
)

func print(f int, debug bool, msg string, data ...interface{}) {
	// fmt.Printf("[%.2fs|%v]: %v\n", float64(f)/60, f, fmt.Sprintf(msg, data...))
	if debug {
		zap.S().Debugf("\t[%.2fs|%v]: %v", float64(f)/60, f, fmt.Sprintf(msg, data...))
		return
	}
	zap.S().Infof("[%.2fs|%v]: %v", float64(f)/60, f, fmt.Sprintf(msg, data...))
}

func PrintFrames(f int) string {
	return fmt.Sprintf("%.2fs|%v", float64(f)/60, f)
}

/**
Stats          map[StatType]float64 //total character stats including from artifact, bonuses, etc...
BaseAtk        float64              //base attack used in calc
BaseDef        float64              //base def used in calc
CharacterLevel int64

DmgBonus       float64              //total damage bonus, including appropriate ele%, etc..
DefMod         float64
ResistMod      float64
ReactionBonus  float64 //reaction bonus %+ such as witch
**/
