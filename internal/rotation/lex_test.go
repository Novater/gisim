package rotation

import (
	"log"
	"testing"
)

var s = `#chain skill into burst if both skill and burst are ready, stay for at least 100 frames
actions+=sequence_strict 
	target=xingqiu 
	exec=skill[1],burst[2]
	lock=100
	if=.cd.xingqiu.burst<>0;

actions+=sequence_strict 
	target=xingqiu 
	exec=skill[2],burst[1]
	lock=100
	if=.cd.xingqiu.burst<>0&&(.cd.xingqiu.burst<>0||.cd.xingqiu.burst<>0);
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
	a, err := p.parse()
	for _, v := range a {
		log.Println(v)
	}
	if err != nil {
		t.Error(err)
	}
}
