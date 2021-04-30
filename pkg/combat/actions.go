package combat

import (
	"strings"
)

func (s *Sim) queueNext() int {
NEXT:
	for i, v := range s.prio {
		ind, ok := s.charPos[v.Target]
		if !ok {
			continue NEXT
		}
		//check active char
		if v.ActiveCond != "" {
			if v.ActiveCond != s.ActiveChar {
				continue NEXT
			}
		}
		//check if we need to swap for this, and if so is swapcd = 0
		if v.Target != s.ActiveChar {
			if s.SwapCD > 0 {
				continue NEXT
			}
		}

		char := s.Chars[ind]
		ready := false

		switch {
		case v.IsSeq && v.IsStrict:
			ready = true
			for _, a := range v.Exec {
				ready = ready && char.ActionReady(a.Typ)
			}
		case v.IsSeq:
			if v.Pos >= len(v.Exec) {
				ready = false
			} else {
				ready = char.ActionReady(v.Exec[v.Pos].Typ)
			}
		default:
			ready = char.ActionReady(v.Exec[0].Typ)
		}

		if !ready {
			continue NEXT
		}

		//walk the tree
		if v.Conditions != nil {
			if !s.evalTree(v.Conditions) {
				continue NEXT
			}
		}

		//add this point ability is ready and we can queue
		//if active char is not current, then add swap first to queue
		if s.ActiveChar != v.Target {
			s.actionQueue = append(s.actionQueue, ActionItem{
				Target: v.Target,
				Typ:    ActionSwap,
			})
		}
		//queue up the abilities
		l := 1
		switch {
		case v.IsSeq && v.IsStrict:
			s.actionQueue = append(s.actionQueue, v.Exec...)
			l = len(v.Exec)
		case v.IsSeq:
			s.actionQueue = append(s.actionQueue, v.Exec[v.Pos])
			v.Pos++
		default:
			s.actionQueue = append(s.actionQueue, v.Exec[0])
		}

		//check for swap lock
		if v.SwapLock > 0 {
			s.SwapCD += v.SwapLock
			s.Log.Debugw("\t locking swap", "swaplock", v.SwapLock, "new cd", s.SwapCD)
		}
		//check for any cancel actions
		switch v.PostAction {
		case ActionDash:
			s.actionQueue = append(s.actionQueue, ActionItem{
				Typ: ActionDash,
			})
			l++
			s.Log.Debugf("\t queueing dash cancel")
		case ActionJump:
			s.actionQueue = append(s.actionQueue, ActionItem{
				Typ: ActionJump,
			})
			l++
			s.Log.Debugf("\t queueing jump cancel")
		}
		//check for any force swaps at the end
		if v.SwapTo != "" {
			if _, ok := s.charPos[v.SwapTo]; ok {
				s.actionQueue = append(s.actionQueue, ActionItem{
					Target: v.SwapTo,
					Typ:    ActionSwap,
				})
				l++
				s.Log.Debugw("\t adding swap cancel", "target", v.SwapTo, "swapped from", v.Target)
			}
		}
		s.Log.Infof("[%v] queuing item %v; name: %v, target: %v, seq: %v, strict: %v, exec: %v", s.Frame(), i+1, v.Name, v.Target, v.IsSeq, v.IsStrict, v.Exec)
		return l
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
		s.Log.Debugw("\t next action queue", "len", i, "queue", s.actionQueue)
	}
	var n ActionItem
	//otherwise pop first item on queue and execute
	n, s.actionQueue = s.actionQueue[0], s.actionQueue[1:]
	c := s.Chars[s.ActiveIndex]
	f := 0
	switch n.Typ {
	case ActionSkill:
		s.executeEventHook(PreSkillHook)
		f = c.Skill(n.Param)
		s.executeEventHook(PostSkillHook)
		s.ResetAllNormalCounter()
	case ActionBurst:
		s.executeEventHook(PreBurstHook)
		f = c.Burst(n.Param)
		s.executeEventHook(PostBurstHook)
		s.ResetAllNormalCounter()
	case ActionAttack:
		f = c.Attack(n.Param)
	case ActionCharge:
		req := c.ActionStam(ActionCharge, n.Param)
		if s.Stam < req {
			diff := int((req-s.Stam)/0.5) + 1
			s.Log.Warnf("[%v] not enough stam (req %v, have %v) to execute charge attack, waiting %v", s.Frame(), req, s.Stam, diff)
			f += diff
		}
		s.Stam -= req
		f += c.ChargeAttack(n.Param)
		s.ResetAllNormalCounter()
	case ActionHighPlunge:
	case ActionLowPlunge:
	case ActionAim:
		f = c.Aimed(n.Param)
		s.ResetAllNormalCounter()
	case ActionSwap:
		s.executeEventHook(PreSwapHook)
		f = 20
		//if we're still in cd then forcefully wait up the cd
		if s.SwapCD > 0 {
			f += s.SwapCD
		}
		s.SwapCD = 60
		s.Log.Infof("[%v] swapped from %v to %v", s.Frame(), s.ActiveChar, n.Target)
		ind := s.charPos[n.Target]
		s.ActiveChar = n.Target
		s.ActiveIndex = ind
		s.ResetAllNormalCounter()
		s.executeEventHook(PostSwapHook)
		s.CharActiveLength = 0
	case ActionCancellable:
	case ActionDash:
		//check if enough stam
		stam := c.ActionStam(ActionDash, n.Param)
		if s.Stam <= stam {
			//we need to wait enough frames for this to be greater than 15, in increments of 0.5 stam per frame
			diff := int((stam-s.Stam)/0.5) + 1
			s.Log.Warnf("[%v] not enough stam to execute dash, waiting %v", s.Frame(), diff)
			f += diff
		}
		s.Stam -= stam
		f = 24
		s.ResetAllNormalCounter()
	case ActionJump:
		f = 35
		s.ResetAllNormalCounter()
	}

	s.Stats.AbilUsageCountByChar[c.Name()][n.Typ.String()]++

	// s.Log.Infof("[%v] %v executing %v", s.Frame(), s.ActiveChar, a.Action)
	s.Log.Infof("[%v] %v executed %v; animation duration %v; swap cd %v", s.Frame(), s.ActiveChar, n.Typ.String(), f, s.SwapCD)

	return f
}

func (s *Sim) ResetAllNormalCounter() {
	for _, c := range s.Chars {
		c.ResetNormalCounter()
	}
}

func (s *Sim) evalTree(node *ExprTreeNode) bool {
	//recursively evaluate tree nodes
	if node.IsLeaf {
		return s.evalCond(node.Expr)
	}
	//so this is a node, then we want to evalute the left and right
	//and then apply operator on both and return that
	left := s.evalTree(node.Left)
	right := s.evalTree(node.Right)
	s.Log.Debugw("evaluating tree node", "left val", left, "right val", right, "node", node)
	switch node.Op {
	case "||":
		return left || right
	case "&&":
		return left && right
	default:
		//if unrecognized op then return false
		return false
	}

}

func (s *Sim) evalCond(c Condition) bool {

	switch c.Fields[0] {
	case ".buff":

	case ".debuff":
		return s.evalDebuff(c)
	case ".element":
		return s.evalElement(c)
	case ".cd":
		return s.evalCD(c)
	case ".status":
		return s.evalStatus(c)
	case ".tags":
		return s.evalTags(c)
	}
	return false
}

func (s *Sim) evalDebuff(c Condition) bool {
	if len(c.Fields) < 2 {
		panic("unexpected short field")
	}
	d := strings.TrimPrefix(c.Fields[1], ".")
	//expecting the value to be either 0 or not 0; 0 for false
	val := c.Value
	if val > 0 {
		val = 1
	} else {
		val = 0
	}
	active := 0
	if s.Target.HasResMod(d) {
		active = 1
	}
	return compInt(c.Op, active, val)
}

func (s *Sim) evalElement(c Condition) bool {
	if len(c.Fields) < 2 {
		panic("unexpected short field")
	}
	ele := strings.TrimPrefix(c.Fields[1], ".")
	//expecting the value to be either 0 or not 0; 0 for false
	val := c.Value
	if val > 0 {
		val = 1
	} else {
		val = 0
	}
	active := 0
	switch ele {
	case "pyro":
		if s.TargetAura.E() == Pyro {
			active = 1
		}
	case "hydro":
		if s.TargetAura.E() == Hydro {
			active = 1
		}
	case "cryo":
		if s.TargetAura.E() == Cryo {
			active = 1
		}
	case "electro":
		if s.TargetAura.E() == Electro {
			active = 1
		}
	case "frozen":
		if s.TargetAura.E() == Frozen {
			active = 1
		}
	case "electro-charged":
		if s.TargetAura.E() == EC {
			active = 1
		}
	default:
		return false
	}
	return compInt(c.Op, active, val)
}

func (s *Sim) evalCD(c Condition) bool {
	if len(c.Fields) < 3 {
		panic("unexpected short field")
	}
	//check target is valid
	name := strings.TrimPrefix(c.Fields[1], ".")
	ci, ok := s.charPos[name]
	if !ok {
		return false
	}
	x := s.Chars[ci]
	cd := -1 // -1 if cd does not exist
	switch c.Fields[2] {
	case ".skill":
		cd = x.Cooldown(ActionSkill)
	case ".burst":
		cd = x.Cooldown(ActionBurst)
	}
	if cd == -1 {
		return false
	}
	//check vs the conditions
	return compInt(c.Op, cd, c.Value)
}

func (s *Sim) evalStatus(c Condition) bool {
	if len(c.Fields) < 3 {
		panic("unexpected short field")
	}
	//check target is valid
	name := strings.TrimPrefix(c.Fields[1], ".")
	ci, ok := s.charPos[name]
	if !ok {
		return false
	}
	x := s.Chars[ci]
	switch c.Fields[2] {
	case ".energy":
		e := x.CurrentEnergy()
		return compFloat(c.Op, e, float64(c.Value))
	default:
		return false
	}
}

func (s *Sim) evalTags(c Condition) bool {
	if len(c.Fields) < 3 {
		panic("unexpected short field")
	}
	//check target is valid
	name := strings.TrimPrefix(c.Fields[1], ".")
	ci, ok := s.charPos[name]
	if !ok {
		return false
	}
	x := s.Chars[ci]
	tag := strings.TrimPrefix(c.Fields[2], ".")
	v := x.Tag(tag)
	s.Log.Debugw("evaluating tags", "char", x.Name(), "targ", c.Fields[2], "val", v)
	return compInt(c.Op, v, c.Value)
}

func compFloat(op string, a, b float64) bool {
	switch op {
	case "==":
		return a == b
	case "!=":
		return a != b
	case ">":
		return a > b
	case ">=":
		return a >= b
	case "<":
		return a < b
	case "<=":
		return a <= b
	}
	return false
}

func compInt(op string, a, b int) bool {
	switch op {
	case "==":
		return a == b
	case "!=":
		return a != b
	case ">":
		return a > b
	case ">=":
		return a >= b
	case "<":
		return a < b
	case "<=":
		return a <= b
	}
	return false
}

type Action struct {
	Name   string
	Target string //either character or a sequence name

	Exec     []ActionItem //if len > 1 then it's a sequence
	IsSeq    bool         // is this a sequence
	IsStrict bool         //strict sequence?
	Pos      int          //current position in execution, default 0

	ActiveCond string
	SwapTo     string
	SwapLock   int
	PostAction ActionType

	Conditions *ExprTreeNode //conditions to be met
}

type ActionItem struct {
	Typ    ActionType
	Param  int
	Target string
}

type ActionType int

const (
	ActionSequence ActionType = iota
	ActionSequenceStrict
	ActionDelimiter
	ActionSequenceReset
	ActionSkill
	ActionBurst
	ActionAttack
	ActionCharge
	ActionHighPlunge
	ActionLowPlunge
	ActionSpecialProc
	ActionAim
	ActionSwap
	ActionCancellable // delim cancellable action
	ActionDash
	ActionJump
)

var astr = []string{
	"sequence",
	"sequence_strict",
	"",
	"reset_sequence",
	"skill",
	"burst",
	"attack",
	"charge",
	"high_plunge",
	"low_plunge",
	"proc",
	"aim",
	"swap",
	"",
	"dash",
	"jump",
}

func (a ActionType) String() string {
	return astr[a]
}

type ExprTreeNode struct {
	Left   *ExprTreeNode
	Right  *ExprTreeNode
	IsLeaf bool
	Op     string //&& || ( )
	Expr   Condition
}

type Condition struct {
	Fields []string
	Op     string
	Value  int
}

func (c Condition) String() {
	var sb strings.Builder
	for _, v := range c.Fields {
		sb.WriteString(v)
	}
	sb.WriteString(c.Op)
}
