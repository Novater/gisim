package combat

import "errors"

func (s *Sim) findNextAction() (ActionItem, error) {
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
		t := s.chars[a.ConditionTarget].CurrentEnergy()
		if t > a.ConditionFloat {
			return false
		}
	}
	return true
}

func (s *Sim) isAbilityReady(a ActionItem) bool {
	//check if character is active
	if a.CharacterName != s.ActiveChar && s.swapCD > 0 {
		return false
	}
	//ask the char if the ability is off cooldown
	ready := s.chars[a.CharacterName].ActionReady(a.Action)
	if ready {
		//check stam cost
		if a.Action == ActionTypeChargedAttack {
			cost := s.chars[a.CharacterName].ChargeAttackStam()
			if cost > s.stam {
				return false
			}
		}
	}
	return ready
}

func (s *Sim) executeAbilityQueue(a []ActionItem) (int, []ActionItem) {
	if len(a) == 0 {
		return 0, nil
	}
	//execute first item
	c := s.chars[s.ActiveChar]
	x := a[0]
	s.Log.Infof("[%v] %v executing %v", s.Frame(), s.ActiveChar, x.Action)
	f := 0
	switch x.Action {
	case ActionTypeSwap:
		s.swapCD = 150
		f = 20
		s.ActiveChar = x.CharacterName
	case ActionTypeDash:
		f = 30
	case ActionTypeJump:
		f = 30
	case ActionTypeAttack:
		f = c.Attack(x.Params)
	case ActionTypeChargedAttack:
		f = c.ChargeAttack(x.Params)
	case ActionTypeAimedShot:
		f = c.Aimed(x.Params)
	case ActionTypeBurst:
		s.executeEventHook(PreBurstHook)
		f = c.Burst(x.Params)
		s.executeEventHook(PostBurstHook)
	case ActionTypeSkill:
		f = c.Skill(x.Params)
	}

	return f, a[1:]
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
