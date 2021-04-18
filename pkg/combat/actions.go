package combat

import (
	"errors"
	"log"

	"github.com/srliao/gisim/internal/rotation"
)

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
		s.Log.Debugw("\t current status", "map", s.Status)
		return s.StatusActive(a.ConditionTarget) == a.ConditionBool
	case "energy lt":
		//check if target energy < threshold
		t := s.Chars[a.index].CurrentEnergy()
		if t > a.ConditionFloat {
			return false
		}
	case "cd":
		return s.Chars[a.index].ActionReady(ActionType(a.ConditionTarget)) == a.ConditionBool
	case "tags":
		t := s.Chars[a.index].Tag(a.ConditionTarget)
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
	ready := s.Chars[a.index].ActionReady(a.Action)
	if ready {
		// s.Log.Infof("[%v]\t\t action %v is ready", s.Frame(), a)
		//check stam cost
		if a.Action == ActionTypeChargedAttack {
			cost := s.Chars[a.index].ChargeAttackStam()
			if cost > s.Stam {
				return false
			}
		}
	}
	return ready
}

func (s *Sim) executeAbilityQueue(a ActionItem) int {

	c := s.Chars[s.ActiveIndex]

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

	s.Stats.AbilUsageCountByChar[c.Name()][string(a.Action)]++

	return f
}

//ActionItem ...
type ActionItem struct {
	CharacterName   string     `yaml:"CharacterName"`
	Action          ActionType `yaml:"Action"`
	Params          int        `yaml:"Params"`
	ConditionType   string     `yaml:"ConditionType"`   //for now either a status or aura
	ConditionTarget string     `yaml:"ConditionTarget"` //which status or aura
	ConditionBool   bool       `yaml:"ConditionBool"`   //true or false
	ConditionFloat  float64    `yaml:"ConditionFloat"`
	ConditionInt    int        `yaml:"ConditionInt"`
	SwapLock        int        `yaml:"SwapLock"`      //number of frames the sim is restricted from swapping after executing this ability
	CancelAbility   ActionType `yaml:"CancelAbility"` //ability to execute to cancel this action
	index           int
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

func (s *Sim) queueNext() int {
NEXT:
	for _, v := range s.prio {
		//check active char
		if v.ActiveCond != "" {
			if v.ActiveCond != s.ActiveChar {
				continue NEXT
			}
		}
		var next rotation.ActionItem
		if v.IsSeq {
			//if strict every action in sequence has to be ready
			if v.IsStrict {
				for _, a := range v.Exec {
					log.Println("check if is ready", a)
				}
			} else {
				//otherwise just the current abil in sequence has to bready
				next = v.Exec[v.Pos]
			}
		} else {
			//otherwise just one abil
			next = v.Exec[0]
		}
		log.Println("check if is ready", next)

		//walk the tree

		//add post actions, swap, swap to, lock
	}
	return 0 //return now many items added to queue
}

func (s *Sim) execQueue() int {
	//if length of q is 0, search for next
	if len(s.actionQueue) == 0 {
		i := s.queueNext()
		if i == 0 {
			return 0
		}
	}
	var n rotation.ActionItem
	//otherwise pop first item on queue and execute
	n, s.actionQueue = s.actionQueue[len(s.actionQueue)-1], s.actionQueue[:len(s.actionQueue)-1]
	c := s.Chars[s.ActiveIndex]
	f := 0
	switch n.Typ {
	case rotation.ActionSkill:
		c.Skill(n.Param)
	case rotation.ActionBurst:
		c.Burst(n.Param)
	case rotation.ActionAttack:
		c.Attack(n.Param)
	case rotation.ActionCharge:
		c.ChargeAttack(n.Param)
	case rotation.ActionHighPlunge:
	case rotation.ActionLowPlunge:
	case rotation.ActionAim:
		c.Aimed(n.Param)
	case rotation.ActionSwap:
		f = 20
		s.SwapCD = 150
	case rotation.ActionCancellable:
	case rotation.ActionDash:
		f = 30
	case rotation.ActionJump:
		f = 30
	}

	// s.Log.Infof("[%v] %v executing %v", s.Frame(), s.ActiveChar, a.Action)

	return f
}

func (s *Sim) evalTree() bool {
	return true
}

func (s *Sim) evalCond(c rotation.Condition) bool {
	if len(c.Fields) < 3 {
		panic("unexpected short field")
	}
	switch c.Fields[0] {
	case "buff":
	case "debuff":
	case "cd":
	}
	return true
}
