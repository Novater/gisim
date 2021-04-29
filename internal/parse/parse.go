package parse

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/srliao/gisim/pkg/combat"
)

var actionKeys = map[string]combat.ActionType{
	"sequence":        combat.ActionSequence,
	"sequence_strict": combat.ActionSequenceStrict,
	"reset_sequence":  combat.ActionSequenceReset,
	"skill":           combat.ActionSkill,
	"burst":           combat.ActionBurst,
	"attack":          combat.ActionAttack,
	"charge":          combat.ActionCharge,
	"high_plunge":     combat.ActionHighPlunge,
	"low_lunge":       combat.ActionLowPlunge,
	"aim":             combat.ActionAim,
	"dash":            combat.ActionDash,
	"jump":            combat.ActionJump,
	"swap":            combat.ActionSwap,
}

type Parser struct {
	input  string
	l      *lexer
	tokens []item
	pos    int
	chars  map[string]*combat.CharacterProfile
	target *combat.EnemyProfile
}

func New(name, input string) *Parser {
	p := &Parser{input: input}
	p.l = lex(name, input)
	p.pos = -1
	return p
}

func (p *Parser) Parse() (combat.Profile, error) {
	var r combat.Profile
	p.target = &combat.EnemyProfile{}
	p.target.Resist = make(map[combat.EleType]float64)
	p.chars = make(map[string]*combat.CharacterProfile)
	for n := p.next(); n.typ != itemEOF; n = p.next() {
		switch n.typ {
		case itemAction:
			next, err := p.parseAction()
			if err != nil {
				return r, err
			}
			r.Rotation = append(r.Rotation, next)
		case itemChar: //char basics
			err := p.parseChar()
			if err != nil {
				return r, err
			}
		case itemStats: //add stats
			err := p.parseStats()
			if err != nil {
				return r, err
			}
		case itemWeapon: //weapon data
			err := p.parseWeapon()
			if err != nil {
				return r, err
			}
		case itemArt: //artifact sets
			err := p.parseArt()
			if err != nil {
				return r, err
			}
		case itemTarget: //enemy related
			err := p.parseTarget()
			if err != nil {
				return r, err
			}
		case itemActive: //active char
			_, err := p.consume(itemAddToList)
			if err != nil {
				return r, err
			}
			n = p.next()
			if n.typ != itemIdentifier {
				return r, fmt.Errorf("<active> bad token at line %v: %v", n.line, n)
			}
			r.InitialActive = n.val
			n = p.next()
			if n.typ != itemTerminateLine {
				return r, fmt.Errorf("<active> bad token at line %v: %v", n.line, n)
			}
		default:
			return r, fmt.Errorf("parse> bad token at line %v: %v", n.line, n)
		}
	}
	for _, v := range p.chars {
		r.Characters = append(r.Characters, *v)
	}
	return r, nil
}

func (p *Parser) newChar(name string) {
	r := combat.CharacterProfile{}
	r.Base.Name = name
	r.Stats = make([]float64, len(combat.StatTypeString))
	r.Sets = make(map[string]int)
	p.chars[name] = &r
}

func (p *Parser) parseChar() error {
	var err error
	//char+=bennett ele=pyro lvl=70 hp=8352 atk=165 def=619 cr=0.05 cd=0.50 er=.2 cons=6 talent=1,8,8;
	_, err = p.consume(itemAddToList)
	if err != nil {
		return err
	}
	//next should be an identifier representing the char
	n, err := p.consume(itemIdentifier)
	if err != nil {
		return err
	}
	name := n.val
	if _, ok := p.chars[name]; !ok {
		p.newChar(name)
	}
	c := p.chars[name]
loop:
	for n := p.next(); n.typ != itemEOF; n = p.next() {
		switch n.typ {
		case itemEle:
			_, err = p.consume(itemAssign)
			if err != nil {
				return err
			}
			n := p.next()
			if n.typ <= eleTypeKeyword {
				return fmt.Errorf("<char> expecting element, got bad token at line %v - %v: %v", n.line, n.pos, n)
			}
			c.Base.Element = combat.EleType(n.val)
		case itemLvl:
			_, err = p.consume(itemAssign)
			if err != nil {
				return err
			}
			n := p.next()
			if n.typ != itemNumber {
				return fmt.Errorf("<char> expecting lvl to be a number, got bad token at line %v - %v: %v", n.line, n.pos, n)
			}
			lvl, err := strconv.Atoi(n.val)
			if err != nil {
				return fmt.Errorf("<char> expecting integer lvl: %v", err)
			}
			c.Base.Level = lvl
		case itemCons:
			_, err = p.consume(itemAssign)
			if err != nil {
				return err
			}
			n := p.next()
			if n.typ != itemNumber {
				return fmt.Errorf("<char> expecting cons to be a number, got bad token at line %v - %v: %v", n.line, n.pos, n)
			}
			lvl, err := strconv.Atoi(n.val)
			if err != nil {
				return fmt.Errorf("<char> expecting integer cons: %v", err)
			}
			c.Base.Cons = lvl
		case statHP:
			_, err = p.consume(itemAssign)
			if err != nil {
				return err
			}
			n := p.next()
			if n.typ != itemNumber {
				return fmt.Errorf("<char> expecting hp to be a number, got bad token at line %v - %v: %v", n.line, n.pos, n)
			}
			amt, err := strconv.ParseFloat(n.val, 64)
			if err != nil {
				return fmt.Errorf("<char> expecting float hp: %v", err)
			}
			c.Base.HP = amt
		case statATK:
			_, err = p.consume(itemAssign)
			if err != nil {
				return err
			}
			n := p.next()
			if n.typ != itemNumber {
				return fmt.Errorf("<char> expecting atk to be a number, got bad token at line %v - %v: %v", n.line, n.pos, n)
			}
			amt, err := strconv.ParseFloat(n.val, 64)
			if err != nil {
				return fmt.Errorf("<char> expecting float atk: %v", err)
			}
			c.Base.Atk = amt
		case statDEF:
			_, err = p.consume(itemAssign)
			if err != nil {
				return err
			}
			n := p.next()
			if n.typ != itemNumber {
				return fmt.Errorf("<char> expecting def to be a number, got bad token at line %v - %v: %v", n.line, n.pos, n)
			}
			amt, err := strconv.ParseFloat(n.val, 64)
			if err != nil {
				return fmt.Errorf("<char> expecting float def: %v", err)
			}
			c.Base.Def = amt
		case itemTalent:
			_, err = p.consume(itemAssign)
			if err != nil {
				return err
			}
			//number = attack
			n := p.next()
			if n.typ != itemNumber {
				return fmt.Errorf("<char> expecting atk to be a number, got bad token at line %v - %v: %v", n.line, n.pos, n)
			}
			atk, err := strconv.Atoi(n.val)
			if err != nil {
				return fmt.Errorf("<char> expecting integer atk talent lvl: %v", err)
			}
			c.Talents.Attack = atk
			_, err = p.consume(itemComma)
			if err != nil {
				return err
			}
			//number = skill
			n = p.next()
			if n.typ != itemNumber {
				return fmt.Errorf("<char> expecting skill to be a number, got bad token at line %v - %v: %v", n.line, n.pos, n)
			}
			skill, err := strconv.Atoi(n.val)
			if err != nil {
				return fmt.Errorf("<char> expecting integer skill talent lvl: %v", err)
			}
			c.Talents.Skill = skill
			_, err = p.consume(itemComma)
			if err != nil {
				return err
			}
			//number = burst
			n = p.next()
			if n.typ != itemNumber {
				return fmt.Errorf("<char> expecting burst to be a number, got bad token at line %v - %v: %v", n.line, n.pos, n)
			}
			burst, err := strconv.Atoi(n.val)
			if err != nil {
				return fmt.Errorf("<char> expecting integer burst talent lvl: %v", err)
			}
			c.Talents.Burst = burst
		case itemTerminateLine:
			break loop
		default:
			if n.typ > statKeyword && n.typ < eleTypeKeyword {
				s := n.val
				//just random stats i guess
				_, err = p.consume(itemAssign)
				if err != nil {
					return err
				}
				n := p.next()
				if n.typ != itemNumber {
					return fmt.Errorf("<char> expecting stat to be a number, got bad token at line %v - %v: %v", n.line, n.pos, n)
				}
				amt, err := strconv.ParseFloat(n.val, 64)
				if err != nil {
					return fmt.Errorf("<char> expecting float stats: %v", err)
				}
				pos := combat.StrToStatType(s)
				c.Stats[pos] += amt
			} else {
				return fmt.Errorf("<char> bad token at line %v - %v: %v", n.line, n.pos, n)
			}
		}
	}
	return nil
}

func (p *Parser) parseStats() error {
	var err error
	//stats+=bennett label=flower hp=4780 def=44 er=.065 cr=.097 cd=.124;
	_, err = p.consume(itemAddToList)
	if err != nil {
		return err
	}
	//next should be an identifier representing the char
	n, err := p.consume(itemIdentifier)
	if err != nil {
		return err
	}
	name := n.val
	if _, ok := p.chars[name]; !ok {
		p.newChar(name)
	}
	c := p.chars[name]
loop:
	for n := p.next(); n.typ != itemEOF; n = p.next() {
		switch {
		case n.typ == itemLabel:
			_, err = p.consume(itemAssign)
			if err != nil {
				return err
			}
			_, err = p.consume(itemIdentifier)
			if err != nil {
				return err
			}
		case n.typ == itemTerminateLine:
			break loop
		case n.typ > statKeyword && n.typ < eleTypeKeyword:
			s := n.val
			//just random stats i guess
			_, err = p.consume(itemAssign)
			if err != nil {
				return err
			}
			n := p.next()
			if n.typ != itemNumber {
				return fmt.Errorf("<stats> expecting stats to be a number, got bad token at line %v - %v: %v", n.line, n.pos, n)
			}
			amt, err := strconv.ParseFloat(n.val, 64)
			if err != nil {
				return fmt.Errorf("<stats> expecting float stats: %v", err)
			}
			pos := combat.StrToStatType(s)
			c.Stats[pos] += amt
		default:
			return fmt.Errorf("<stats> bad token at line %v - %v: %v", n.line, n.pos, n)

		}
	}
	return nil
}

func (p *Parser) parseWeapon() error {
	var err error
	//weapon+=bennett label="festering desire" atk=401 er=0.559 refine=3;
	_, err = p.consume(itemAddToList)
	if err != nil {
		return err
	}
	//next should be an identifier representing the char
	n, err := p.consume(itemIdentifier)
	if err != nil {
		return err
	}
	name := n.val
	if _, ok := p.chars[name]; !ok {
		p.newChar(name)
	}
	c := p.chars[name]
loop:
	for n := p.next(); n.typ != itemEOF; n = p.next() {
		switch n.typ {
		case itemLabel:
			_, err = p.consume(itemAssign)
			if err != nil {
				return err
			}
			n, err = p.consume(itemString)
			if err != nil {
				return err
			}
			s := n.val
			if len(s) > 0 && s[0] == '"' {
				s = s[1:]
			}
			if len(s) > 0 && s[len(s)-1] == '"' {
				s = s[:len(s)-1]
			}
			c.Weapon.Name = s
		case itemRefine:
			_, err = p.consume(itemAssign)
			if err != nil {
				return err
			}
			n := p.next()
			if n.typ != itemNumber {
				return fmt.Errorf("<weapon> expecting refine to be a number, got bad token at line %v - %v: %v", n.line, n.pos, n)
			}
			lvl, err := strconv.Atoi(n.val)
			if err != nil {
				return fmt.Errorf("<weapon> expecting integer refine: %v", err)
			}
			c.Weapon.Refine = lvl
		case statATK:
			_, err = p.consume(itemAssign)
			if err != nil {
				return err
			}
			n := p.next()
			if n.typ != itemNumber {
				return fmt.Errorf("<weapon> expecting atk to be a number, got bad token at line %v - %v: %v", n.line, n.pos, n)
			}
			amt, err := strconv.ParseFloat(n.val, 64)
			if err != nil {
				return fmt.Errorf("<weapon> expecting float atk: %v", err)
			}
			c.Weapon.Atk = amt
		case itemTerminateLine:
			break loop
		default:
			if n.typ > statKeyword && n.typ < eleTypeKeyword {
				s := n.val
				//just random stats i guess
				_, err = p.consume(itemAssign)
				if err != nil {
					return err
				}
				n := p.next()
				if n.typ != itemNumber {
					return fmt.Errorf("<weapon> expecting stats to be a number, got bad token at line %v - %v: %v", n.line, n.pos, n)
				}
				amt, err := strconv.ParseFloat(n.val, 64)
				if err != nil {
					return fmt.Errorf("<weapon> expecting float stats: %v", err)
				}
				pos := combat.StrToStatType(s)
				c.Stats[pos] += amt
			} else {
				return fmt.Errorf("<weapon> bad token at line %v - %v: %v", n.line, n.pos, n)
			}
		}
	}
	return nil
}

func (p *Parser) parseArt() error {
	var err error
	//art+=xiangling label="gladiator's finale" count=2;
	_, err = p.consume(itemAddToList)
	if err != nil {
		return err
	}
	//next should be an identifier representing the char
	n, err := p.consume(itemIdentifier)
	if err != nil {
		return err
	}
	name := n.val
	if _, ok := p.chars[name]; !ok {
		p.newChar(name)
	}
	c := p.chars[name]
	var label string
	var count int
loop:
	for n := p.next(); n.typ != itemEOF; n = p.next() {
		switch n.typ {
		case itemLabel:
			_, err = p.consume(itemAssign)
			if err != nil {
				return err
			}
			n, err = p.consume(itemString)
			if err != nil {
				return err
			}
			s := n.val
			if len(s) > 0 && s[0] == '"' {
				s = s[1:]
			}
			if len(s) > 0 && s[len(s)-1] == '"' {
				s = s[:len(s)-1]
			}
			label = s
		case itemCount:
			_, err = p.consume(itemAssign)
			if err != nil {
				return err
			}
			n := p.next()
			if n.typ != itemNumber {
				return fmt.Errorf("<art> expecting count to be a number, got bad token at line %v - %v: %v", n.line, n.pos, n)
			}
			i, err := strconv.Atoi(n.val)
			if err != nil {
				return fmt.Errorf("<art> expecting integer count: %v", err)
			}
			count = i
		case itemTerminateLine:
			break loop
		default:
			return fmt.Errorf("<art> bad token at line %v - %v: %v", n.line, n.pos, n)
		}
	}
	c.Sets[label] = count
	return nil
}

func (p *Parser) parseTarget() error {
	var err error
	//stats+=bennett label=flower hp=4780 def=44 er=.065 cr=.097 cd=.124;
	_, err = p.consume(itemAddToList)
	if err != nil {
		return err
	}
	//next should be an identifier or a string
	n := p.next()
	if n.typ != itemIdentifier && n.typ != itemString {
		return fmt.Errorf("<target> expecting a string name after target, got bad token at line %v - %v: %v", n.line, n.pos, n)
	}
loop:
	for n := p.next(); n.typ != itemEOF; n = p.next() {
		switch {
		case n.typ == itemLvl:
			_, err = p.consume(itemAssign)
			if err != nil {
				return err
			}
			n := p.next()
			if n.typ != itemNumber {
				return fmt.Errorf("<target> expecting lvl to be a number, got bad token at line %v - %v: %v", n.line, n.pos, n)
			}
			lvl, err := strconv.Atoi(n.val)
			if err != nil {
				return fmt.Errorf("<target> expecting integer lvl: %v", err)
			}
			p.target.Level = lvl
		case n.typ > eleTypeKeyword:
			s := n.val
			//just random stats i guess
			_, err = p.consume(itemAssign)
			if err != nil {
				return err
			}
			n := p.next()
			if n.typ != itemNumber {
				return fmt.Errorf("<target> expecting stat to be a number, got bad token at line %v - %v: %v", n.line, n.pos, n)
			}
			amt, err := strconv.ParseFloat(n.val, 64)
			if err != nil {
				return fmt.Errorf("<target> expecting float lvl: %v", err)
			}
			p.target.Resist[combat.EleType(s)] += amt
		case n.typ == itemTerminateLine:
			break loop
		default:
			return fmt.Errorf("<target> bad token at line %v - %v: %v", n.line, n.pos, n)

		}
	}
	return nil
}

func (p *Parser) parseAction() (combat.Action, error) {
	var err error
	var r combat.Action
	err = p.parseActionItem(&r)
	if err != nil {
		return r, err
	}
READLOOP:
	for n := p.next(); n.typ != itemEOF; n = p.next() {
		switch n.typ {
		case itemTarget:
			r.Target, err = p.parseStringIdent()
			if err != nil {
				return r, err
			}
		case itemExec:
			r.Exec, err = p.parseExec()
			if err != nil {
				return r, err
			}
		case itemLock:
			r.SwapLock, err = p.parseLock()
			if err != nil {
				return r, err
			}
		case itemIf:
			r.Conditions, err = p.parseIf()
			if err != nil {
				return r, err
			}
		case itemSwap:
			r.SwapTo, err = p.parseStringIdent()
			if err != nil {
				return r, err
			}
		case itemPost:
			r.PostAction, err = p.parsePostAction()
			if err != nil {
				return r, err
			}
		case itemActive:
			r.ActiveCond, err = p.parseStringIdent()
			if err != nil {
				return r, err
			}
		case itemTerminateLine:
			if err := isActionValid(r); err != nil {
				return r, fmt.Errorf("bad action: %v", err)
			}
			break READLOOP
		default:
			return r, fmt.Errorf("bad token at line %v - %v: %v", n.line, n.pos, n)
		}
	}
	return r, nil
}

func (p *Parser) parseActionItem(next *combat.Action) error {
	_, err := p.consume(itemAddToList)
	if err != nil {
		return err
	}
	//next should be a keyword
	n := p.next()
	// log.Println(n)
	if n.typ != itemIdentifier {
		return fmt.Errorf("<action> bad token at line %v: %v", n.line, n)
	}
	t, ok := actionKeys[n.val]
	if !ok {
		return fmt.Errorf("<action> invalid identifier at line %v: %v", n.line, n)
	}
	a := combat.ActionItem{}
	switch {
	case t == combat.ActionSequence:
		next.IsSeq = true
	case t == combat.ActionSequenceStrict:
		next.IsSeq = true
		next.IsStrict = true
	case t > combat.ActionDelimiter:
		a.Typ = t
		//check for params
		n = p.next()
		if n.typ != itemLeftSquareParen {
			p.backup()
			next.Exec = append(next.Exec, a)
			return nil
		}
		// log.Println(n)
		//next should be numbers
		n = p.next()
		// log.Println(n)
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
		// log.Println(n)
		next.Exec = append(next.Exec, a)

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

func (p *Parser) parsePostAction() (combat.ActionType, error) {
	var t combat.ActionType
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
	if t <= combat.ActionCancellable {
		return t, fmt.Errorf("<post - cancel> invalid post action at line %v - %v: %v", n.line, n.pos, n)
	}
	return t, nil
}

func (p *Parser) parseExec() ([]combat.ActionItem, error) {
	var r []combat.ActionItem
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
		if t <= combat.ActionDelimiter {
			return nil, fmt.Errorf("<exec> bad token at line %v - %v: %v", n.line, n.pos, n)
		}

		a := combat.ActionItem{}
		a.Typ = t
		//check for params
		n = p.next()

		if n.typ == itemLeftSquareParen {
			//next should be numbers
			n = p.next()
			if n.typ != itemNumber {
				return nil, fmt.Errorf("<exec - num> invalid number at line %v: %v", n.line, n)
			}
			x, err := strconv.ParseInt(n.val, 10, 64)
			if err != nil {
				return nil, err
			}
			a.Param = int(x)

			//then we have close bracket
			n = p.next()
			if n.typ != itemRightSquareParen {
				return nil, fmt.Errorf("<exec - right paren> bad token at line %v: %v", n.line, n)
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

func (p *Parser) parseIf() (*combat.ExprTreeNode, error) {
	n := p.next()
	if n.typ != itemAssign {
		return nil, fmt.Errorf("<if> bad token at line %v - %v: %v", n.line, n.pos, n)
	}
	parenDepth := 0
	var queue []*combat.ExprTreeNode
	var stack []*combat.ExprTreeNode
	var x *combat.ExprTreeNode
	var root *combat.ExprTreeNode

	//operands are conditions
	//operators are &&, ||, (, )
LOOP:
	for {
		//we expect either brackets, or a field
		n = p.next()
		switch {
		case n.typ == itemLeftParen:
			parenDepth++
			stack = append(stack, &combat.ExprTreeNode{
				Op: "(",
			})
			//expecting a condition after a paren
			c, err := p.parseCondition()
			if err != nil {
				return nil, err
			}
			queue = append(queue, &combat.ExprTreeNode{
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
			queue = append(queue, &combat.ExprTreeNode{
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
			stack = append(stack, &combat.ExprTreeNode{
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

	var ts []*combat.ExprTreeNode
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

func (p *Parser) parseCondition() (combat.Condition, error) {
	var c combat.Condition
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
	c.Op = n.val
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

func isActionValid(a combat.Action) error {
	if a.Target == "" {
		return errors.New("missing target")
	}
	if len(a.Exec) == 0 {
		return errors.New("missing actions")
	}
	return nil
}

func (p *Parser) consume(i ItemType) (item, error) {
	n := p.next()
	if n.typ != i {
		return n, fmt.Errorf("expecting %v, got bad token at line %v - %v: %v", i, n.line, n.pos, n)
	}
	return n, nil
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
