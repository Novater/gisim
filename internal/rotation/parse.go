package rotation

import (
	"fmt"
	"strconv"
	"strings"
)

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

	//fields used by parser only
	sourceLine int
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
	"low_lunge",
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

var actionKeys = map[string]ActionType{
	"sequence":        ActionSequence,
	"sequence_strict": ActionSequenceStrict,
	"reset_sequence":  ActionSequenceReset,
	"skill":           ActionSkill,
	"burst":           ActionBurst,
	"attack":          ActionAttack,
	"charge":          ActionCharge,
	"high_plunge":     ActionHighPlunge,
	"low_lunge":       ActionLowPlunge,
	"aim":             ActionAim,
	"dash":            ActionDash,
	"jump":            ActionJump,
	"swap":            ActionSwap,
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
	Op     item
	Value  int
}

func (c Condition) String() {
	var sb strings.Builder
	for _, v := range c.Fields {
		sb.WriteString(v)
	}
	sb.WriteString(c.Op.String())
}

type Parser struct {
	input  string
	l      *lexer
	tokens []item
	pos    int
}

func New(name, input string) *Parser {
	p := &Parser{input: input}
	p.l = lex(name, input)
	p.pos = -1
	return p
}

func (p *Parser) Parse() ([]Action, error) {
	var err error
	var r []Action
	state := 0
	var next Action
	for n := p.next(); n.typ != itemEOF; n = p.next() {
		switch state {
		case 0:
			//the next keyword needs to be action, other wise we have an error
			if n.typ != itemAction {
				return r, fmt.Errorf("<action> bad token at line %v: %v", n.line, n)
			}
			next = Action{}
			next.sourceLine = n.line
			err = p.parseActionItem(&next)
			if err != nil {
				return r, err
			}
			state = 1
		case 1:
			// log.Println(n)
			switch n.typ {
			case itemTarget:
				next.Target, err = p.parseStringIdent()
				if err != nil {
					return r, err
				}
			case itemExec:
				next.Exec, err = p.parseExec()
				if err != nil {
					return r, err
				}
			case itemLock:
				next.SwapLock, err = p.parseLock()
				if err != nil {
					return r, err
				}
			case itemIf:
				next.Conditions, err = p.parseIf()
				if err != nil {
					return r, err
				}
			case itemSwap:
				next.SwapTo, err = p.parseStringIdent()
				if err != nil {
					return r, err
				}
			case itemPost:
				next.PostAction, err = p.parsePostAction()
				if err != nil {
					return r, err
				}
			case itemActive:
				next.ActiveCond, err = p.parseStringIdent()
				if err != nil {
					return r, err
				}
			case itemTerminateLine:
				r = append(r, next)
				state = 0
			default:
				return r, fmt.Errorf("bad token at line %v - %v: %v", n.line, n.pos, n)
			}
		}
	}
	return r, nil
}

func (p *Parser) parseActionItem(next *Action) error {
	n := p.next()
	if n.typ != itemAddToList {
		return fmt.Errorf("<action> bad token at line %v: %v", n.line, n)
	}
	//next should be a keyword
	n = p.next()
	if n.typ != itemIdentifier {
		return fmt.Errorf("<action> bad token at line %v: %v", n.line, n)
	}
	t, ok := actionKeys[n.val]
	if !ok {
		return fmt.Errorf("<action> invalid identifier at line %v: %v", n.line, n)
	}
	a := ActionItem{}
	switch {
	case t == ActionSequence:
		next.IsSeq = true
	case t == ActionSequenceStrict:
		next.IsSeq = true
		next.IsStrict = true
	case t > ActionDelimiter:
		a.Typ = t
		//check for params
		n = p.next()
		if n.typ != itemLeftSquareParen {
			p.backup()
			next.Exec = append(next.Exec, a)
			return nil
		}
		//next should be numbers
		n = p.next()
		if n.typ != itemNumber {
			return fmt.Errorf("<action> invalid number at line %v: %v", n.line, n)
		}
		x, err := strconv.ParseInt(n.val, 10, 64)
		if err != nil {
			return err
		}
		a.Param = int(x)
		//then we have close bracket
		n = p.next()
		if n.typ != itemRightSquareParen {
			return fmt.Errorf("<action> bad token at line %v: %v", n.line, n)
		}

	}

	return nil
}

func (p *Parser) parseStringIdent() (string, error) {
	r := ""
	n := p.next()
	if n.typ != itemAssign {
		return r, fmt.Errorf("<string - assign> bad token at line %v - %v: %v", n.line, n.pos, n)
	}
	n = p.next()
	if n.typ != itemIdentifier {
		return r, fmt.Errorf("<string - id> bad token at line %v - %v: %v %v", n.line, n.pos, n, n.typ)
	}
	r = n.val

	return r, nil
}

func (p *Parser) parsePostAction() (ActionType, error) {
	var t ActionType
	n := p.next()
	if n.typ != itemAssign {
		return t, fmt.Errorf("<post - assign> bad token at line %v - %v: %v", n.line, n.pos, n)
	}
	n = p.next()
	if n.typ != itemIdentifier {
		return t, fmt.Errorf("<post - id> bad token at line %v - %v: %v", n.line, n.pos, n)
	}
	t, ok := actionKeys[n.val]
	if !ok {
		return t, fmt.Errorf("<post - val id> bad token at line %v - %v: %v", n.line, n.pos, n)
	}
	if t <= ActionCancellable {
		return t, fmt.Errorf("<post - cancel> invalid post action at line %v - %v: %v", n.line, n.pos, n)
	}
	return t, nil
}

func (p *Parser) parseExec() ([]ActionItem, error) {
	var r []ActionItem
	n := p.next()
	if n.typ != itemAssign {
		return nil, fmt.Errorf("<exec> bad token at line %v - %v: %v", n.line, n.pos, n)
	}
LOOP:
	for {
		n = p.next()
		// log.Printf("exec item %v\n", n)
		if n.typ != itemIdentifier {
			return nil, fmt.Errorf("<exec> bad token at line %v - %v: %v", n.line, n.pos, n)
		}
		t, ok := actionKeys[n.val]
		if !ok {
			return nil, fmt.Errorf("<exec> bad token at line %v - %v: %v", n.line, n.pos, n)
		}
		if t <= ActionDelimiter {
			return nil, fmt.Errorf("<exec> bad token at line %v - %v: %v", n.line, n.pos, n)
		}

		a := ActionItem{}
		a.Typ = t
		//check for params
		n = p.next()

		if n.typ == itemLeftSquareParen {
			//next should be numbers
			n = p.next()
			if n.typ != itemNumber {
				return nil, fmt.Errorf("<action> invalid number at line %v: %v", n.line, n)
			}
			x, err := strconv.ParseInt(n.val, 10, 64)
			if err != nil {
				return nil, err
			}
			a.Param = int(x)

			//then we have close bracket
			n = p.next()
			if n.typ != itemRightSquareParen {
				return nil, fmt.Errorf("<action> bad token at line %v: %v", n.line, n)
			}
		} else {
			p.backup()
		}

		r = append(r, a)
		n = p.next()
		if n.typ != itemComma {
			p.backup()
			break LOOP
		}
	}

	return r, nil
}

func (p *Parser) parseIf() (*ExprTreeNode, error) {
	n := p.next()
	if n.typ != itemAssign {
		return nil, fmt.Errorf("<if> bad token at line %v - %v: %v", n.line, n.pos, n)
	}
	parenDepth := 0
	var queue []*ExprTreeNode
	var stack []*ExprTreeNode
	var x *ExprTreeNode
	var root *ExprTreeNode

	//operands are conditions
	//operators are &&, ||, (, )
LOOP:
	for {
		//we expect either brackets, or a field
		n = p.next()
		switch {
		case n.typ == itemLeftParen:
			parenDepth++
			stack = append(stack, &ExprTreeNode{
				Op: "(",
			})
			//expecting a condition after a paren
			c, err := p.parseCondition()
			if err != nil {
				return nil, err
			}
			queue = append(queue, &ExprTreeNode{
				Expr:   c,
				IsLeaf: true,
			})
		case n.typ == itemRightParen:

			parenDepth--
			if parenDepth < 0 {
				return nil, fmt.Errorf("unmatched right paren")
			}
			/**
			Else if token is a right parenthesis
				Until the top token (from the stack) is left parenthesis, pop from the stack to the output buffer
				Also pop the left parenthesis but don’t include it in the output buffe
			**/

			for {
				x, stack = stack[len(stack)-1], stack[:len(stack)-1]
				if x.Op == "(" {
					break
				}
				queue = append(queue, x)
			}

		case n.typ == itemField:
			p.backup()
			//scan for fields
			c, err := p.parseCondition()
			if err != nil {
				return nil, err
			}
			queue = append(queue, &ExprTreeNode{
				Expr:   c,
				IsLeaf: true,
			})
		}

		//check if any logical ops
		n = p.next()
		switch {
		case n.typ > itemLogicOP && n.typ < itemCompareOp:
			//check if top of stack is an operator
			if len(stack) > 0 && stack[len(stack)-1].Op != "(" {
				//pop and add to output
				x, stack = stack[len(stack)-1], stack[:len(stack)-1]
				queue = append(queue, x)
			}
			//append current operator to stack
			stack = append(stack, &ExprTreeNode{
				Op: n.val,
			})
		case n.typ == itemRightParen:
			p.backup()
		default:
			p.backup()
			break LOOP
		}
	}

	if parenDepth > 0 {
		return nil, fmt.Errorf("unmatched left paren")
	}

	for i := len(stack) - 1; i >= 0; i-- {
		queue = append(queue, stack[i])
	}

	var ts []*ExprTreeNode
	//convert this into a tree
	for _, v := range queue {
		if v.Op != "" {
			// fmt.Printf("%v ", v.Op)
			//pop 2 nodes from tree
			if len(ts) < 2 {
				panic("tree stack less than 2 before operator")
			}
			v.Left, ts = ts[len(ts)-1], ts[:len(ts)-1]
			v.Right, ts = ts[len(ts)-1], ts[:len(ts)-1]
			ts = append(ts, v)
		} else {
			// fmt.Printf("%v ", v.Expr)
			ts = append(ts, v)
		}
	}
	// fmt.Printf("\n")

	root = ts[0]
	return root, nil
}

func (p *Parser) parseCondition() (Condition, error) {
	var c Condition
	var n item
LOOP:
	for {
		//look for a field
		n = p.next()
		if n.typ != itemField {
			return c, fmt.Errorf("<if - field> bad token at line %v - %v: %v", n.line, n.pos, n)
		}
		c.Fields = append(c.Fields, n.val)
		//see if any more fields
		n = p.peek()
		if n.typ != itemField {
			break LOOP
		}
	}

	//scan for comparison op
	n = p.next()
	if n.typ <= itemCompareOp || n.typ >= itemKeyword {
		return c, fmt.Errorf("<if - comp> bad token at line %v - %v: %v", n.line, n.pos, n)
	}
	c.Op = n
	//scan for value
	n = p.next()
	if n.typ != itemNumber {
		return c, fmt.Errorf("<if - num> bad token at line %v - %v: %v", n.line, n.pos, n)
	}
	val, err := strconv.ParseInt(n.val, 10, 64)
	if err != nil {
		return c, fmt.Errorf("<if - strconv> bad token at line %v - %v: %v", n.line, n.pos, n)
	}
	c.Value = int(val)
	return c, nil
}

func (p *Parser) parseLock() (int, error) {
	n := p.next()
	if n.typ != itemAssign {
		return -1, fmt.Errorf("<target> bad token at line %v - %v: %v", n.line, n.pos, n)
	}
	n = p.next()
	if n.typ != itemNumber {
		return -1, fmt.Errorf("<target> bad token at line %v - %v: %v", n.line, n.pos, n)
	}
	r, err := strconv.ParseInt(n.val, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("<target> bad token at line %v - %v: %v", n.line, n.pos, n)
	}
	return int(r), nil
}

func isActionValid(a Action) bool {
	if a.Target == "" {
		return false
	}
	if len(a.Exec) == 0 {
		return false
	}
	return true
}

func (p *Parser) next() item {
	p.pos++
	if p.pos == len(p.tokens) {
		t := p.l.nextItem()
		p.tokens = append(p.tokens, t)
	}
	// log.Printf("next token: %v", p.tokens[p.pos])
	return p.tokens[p.pos]
}

func (p *Parser) backup() {
	if p.pos > 0 {
		p.pos--
	}
}

func (p *Parser) peek() item {
	next := p.next()
	p.backup()
	return next
}
