package parse

import (
	"log"
	"testing"
)

var s = `
actions+=burst target=fischl if=.tags.fischl.oz==0;
actions+=skill target=fischl if=.tags.fischl.oz==0;
actions+=sequence_strict
  target=eula
  exec=skill,burst,skill[1],attack,attack,attack,attack,dash,attack,attack,attack,attack;
#use hold E, but only if burst is not coming off cd in 10s
actions+=skill[1]
  target=eula
  if=.cd.eula.burst>600&&.tags.eula.grimheart==2;
#use e, but only if burst is not coming off cd in 4 seconds
actions+=skill
  target=eula
  if=.cd.eula.burst>240;
actions+=attack target=eula;
actions+=attack target=fischl active=fischl;
`

func testLex(t *testing.T) {
	log.Println("testing lex")

	l := lex("test", s)
	last := "roar"
	// stop := false
	for n := l.nextItem(); n.typ != itemEOF; n = l.nextItem() {
		if n.typ == itemAction {
			log.Printf("action start at line %v\n", n.line)
		}
		log.Println(n)
		if n.val == last {
			t.FailNow()
		}
		last = n.val
	}

}

func TestParse(t *testing.T) {
	p := New("test", s)
	a, err := p.Parse()
	for _, v := range a {
		log.Println(v)
	}
	if err != nil {
		t.Error(err)
	}
}
