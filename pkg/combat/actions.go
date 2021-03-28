package combat

import "errors"

func FindNextAction(s *Sim) (ActionItem, error) {
	for _, a := range s.actions {
		if s.isAbilityReady(a) && s.abilityConditionsOk(a) {
			return a, nil
		}
	}
	return ActionItem{}, errors.New("no ability available")
}

func (s *Sim) abilityConditionsOk(a ActionItem) bool {
	switch a.ConditionType {
	case "status":
		_, ok := s.Status[a.ConditionTarget]
		if ok != a.ConditionBool {
			return false
		}
	case "energy lt":
		//check if target energy < threshold
		t := s.Chars[a.ConditionTarget].CurrentEnergy()
		if t > a.ConditionFloat {
			return false
		}
	case "tags":
		t := s.Chars[a.CharacterName].Tag(a.ConditionTarget)
		s.Log.Debugw("\t checking for tags", "want", a.ConditionInt, "got", t, "action", a)
		if t != a.ConditionInt {
			return false
		}
	}
	return true
}

func (s *Sim) isAbilityReady(a ActionItem) bool {
	//check if character is active
	if a.CharacterName != s.ActiveChar && s.SwapCD > 0 {
		return false
	}
	//ask the char if the ability is off cooldown
	ready := s.Chars[a.CharacterName].ActionReady(a.Action)
	if ready {
		// s.Log.Infof("[%v]\t\t action %v is ready", s.Frame(), a)
		//check stam cost
		if a.Action == ActionTypeChargedAttack {
			cost := s.Chars[a.CharacterName].ChargeAttackStam()
			if cost > s.Stam {
				return false
			}
		}
	}
	return ready
}

func (s *Sim) executeAbilityQueue(a ActionItem) int {

	c := s.Chars[s.ActiveChar]

	s.Log.Infof("[%v] %v executing %v", s.Frame(), s.ActiveChar, a.Action)
	f := 0
	switch a.Action {
	case ActionTypeDash:
		f = 30
	case ActionTypeJump:
		f = 30
	case ActionTypeAttack:
		f = c.Attack(a.Params)
	case ActionTypeChargedAttack:
		f = c.ChargeAttack(a.Params)
	case ActionTypeAimedShot:
		f = c.Aimed(a.Params)
	case ActionTypeBurst:
		s.executeEventHook(PreBurstHook)
		f = c.Burst(a.Params)
		s.executeEventHook(PostBurstHook)
	case ActionTypeSkill:
		f = c.Skill(a.Params)
	}

	if a.SwapLock > 0 {
		s.SwapCD += a.SwapLock
		s.Log.Debugw("\t locking swap", "swaplock", a.SwapLock, "new cd", s.SwapCD)
	}

	return f
}

//ActionItem ...
type ActionItem struct {
	CharacterName   string                 `yaml:"CharacterName"`
	Action          ActionType             `yaml:"Action"`
	Params          map[string]interface{} `yaml:"Params"`
	ConditionType   string                 `yaml:"ConditionType"`   //for now either a status or aura
	ConditionTarget string                 `yaml:"ConditionTarget"` //which status or aura
	ConditionBool   bool                   `yaml:"ConditionBool"`   //true or false
	ConditionFloat  float64                `yaml:"ConditionFloat"`
	ConditionInt    int                    `yaml:"ConditionInt"`
	SwapLock        int                    `yaml:"SwapLock"`      //number of frames the sim is restricted from swapping after executing this ability
	CancelAbility   ActionType             `yaml:"CancelAbility"` //ability to execute to cancel this action
}

type ActionType string

//ActionType constants
const (
	//motions
	ActionTypeSwap ActionType = "swap"
	ActionTypeDash ActionType = "dash"
	ActionTypeJump ActionType = "jump"
	//main actions
	ActionTypeAttack    ActionType = "attack"
	ActionTypeAimedShot ActionType = "aimed"
	ActionTypeSkill     ActionType = "skill"
	ActionTypeBurst     ActionType = "burst"
	//derivative actions
	ActionTypeChargedAttack ActionType = "charge"
	ActionTypePlungeAttack  ActionType = "plunge"
	//procs
	ActionTypeSpecialProc ActionType = "proc"
	//xiao special
	ActionTypeXiaoLowJump  ActionType = "xiao-low-jump"
	ActionTypeXiaoHighJump ActionType = "xiao-high-jump"
)
