package combat

import (
	"strings"

	"github.com/srliao/gisim/internal/rotation"
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
			s.actionQueue = append(s.actionQueue, rotation.ActionItem{
				Target: v.Target,
				Typ:    rotation.ActionSwap,
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
		case rotation.ActionDash:
			s.actionQueue = append(s.actionQueue, rotation.ActionItem{
				Typ: rotation.ActionDash,
			})
			l++
			s.Log.Debugf("\t queueing dash cancel")
		case rotation.ActionJump:
			s.actionQueue = append(s.actionQueue, rotation.ActionItem{
				Typ: rotation.ActionJump,
			})
			l++
			s.Log.Debugf("\t queueing jump cancel")
		}
		//check for any force swaps at the end
		if v.SwapTo != "" {
			if _, ok := s.charPos[v.SwapTo]; ok {
				s.actionQueue = append(s.actionQueue, rotation.ActionItem{
					Target: v.SwapTo,
					Typ:    rotation.ActionSwap,
				})
				l++
				s.Log.Debugw("\t adding swap cancel", "target", v.SwapTo, "swapped from", v.Target)
			}
		}
		s.Log.Infof("[%v] queuing rotation item %v; name: %v, target: %v, seq: %v, strict: %v, exec: %v", s.Frame(), i+1, v.Name, v.Target, v.IsSeq, v.IsStrict, v.Exec)
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
	var n rotation.ActionItem
	//otherwise pop first item on queue and execute
	n, s.actionQueue = s.actionQueue[0], s.actionQueue[1:]
	c := s.Chars[s.ActiveIndex]
	f := 0
	switch n.Typ {
	case rotation.ActionSkill:
		f = c.Skill(n.Param)
	case rotation.ActionBurst:
		s.executeEventHook(PreBurstHook)
		f = c.Burst(n.Param)
		s.executeEventHook(PostBurstHook)
	case rotation.ActionAttack:
		f = c.Attack(n.Param)
	case rotation.ActionCharge:
		f = c.ChargeAttack(n.Param)
	case rotation.ActionHighPlunge:
	case rotation.ActionLowPlunge:
	case rotation.ActionAim:
		f = c.Aimed(n.Param)
	case rotation.ActionSwap:
		f = 20
		//if we're still in cd then forcefully wait up the cd
		if s.SwapCD > 0 {
			f += s.SwapCD
		}
		s.SwapCD = 150
		s.Log.Infof("[%v] swapped from %v to %v", s.Frame(), s.ActiveChar, n.Target)
		ind := s.charPos[n.Target]
		s.ActiveChar = n.Target
		s.ActiveIndex = ind

	case rotation.ActionCancellable:
	case rotation.ActionDash:
		f = 30
	case rotation.ActionJump:
		f = 30
	}

	s.Stats.AbilUsageCountByChar[c.Name()][n.Typ.String()]++

	// s.Log.Infof("[%v] %v executing %v", s.Frame(), s.ActiveChar, a.Action)
	s.Log.Infof("[%v] %v executed %v; animation duration %v; swap cd %v", s.Frame(), s.ActiveChar, n.Typ.String(), f, s.SwapCD)

	return f
}

func (s *Sim) evalTree(node *rotation.ExprTreeNode) bool {
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

func (s *Sim) evalCond(c rotation.Condition) bool {
	if len(c.Fields) < 3 {
		panic("unexpected short field")
	}
	switch c.Fields[0] {
	case ".buff":

	case ".debuff":
	case ".cd":
		return s.evalCD(c)
	case ".status":
		return s.evalStatus(c)
	case ".tags":
		return s.evalTags(c)
	}
	return false
}

func (s *Sim) evalCD(c rotation.Condition) bool {
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
		cd = x.Cooldown(rotation.ActionSkill)
	case ".burst":
		cd = x.Cooldown(rotation.ActionBurst)
	}
	if cd == -1 {
		return false
	}
	//check vs the conditions
	return compInt(c.Op.String(), cd, c.Value)
}

func (s *Sim) evalStatus(c rotation.Condition) bool {
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
		return compFloat(c.Op.String(), e, float64(c.Value))
	default:
		return false
	}
}

func (s *Sim) evalTags(c rotation.Condition) bool {
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
	return compInt(c.Op.String(), v, c.Value)
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
