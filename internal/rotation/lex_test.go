package rotation

import (
	"log"
	"testing"
)

var s = `#chain skill into burst if both skill and burst are ready, stay for at least 100 frames
actions+=sequence_strict 
	target=Xingqiu 
	exec=skill[1],burst[2]
	lock=100
	if=.cd.Xingqiu.burst<>0;

actions+=sequence_strict 
	target=Xingqiu 
	exec=skill[2],burst[1]
	lock=100
	if=.cd.Xingqiu.burst<>0&&(.cd.Xingqiu.burst<>0||.cd.Xingqiu.burst<>0);

actions+=sequence_strict target=xingqiu exec=skill,burst lock=100;
actions+=skill target=xingqiu if=.status.xingqiu.energy<80;
actions+=burst target=xingqiu;
actions+=burst target=bennett;
actions+=sequence_strict target=xiangling exec=skill,burst;
actions+=skill target=xiangling active=xiangling;
actions+=skill target=bennett if=.status.xiangling.energy<70&&.cd.xiangling.burst<120 swap=xiangling;
actions+=burst target=fischl if=.status.xiangling.energy<70&&.buff.fischl.oz==0 swap=xiangling;
actions+=skill target=fischl if=.status.xiangling.energy<70&&.buff.fischl.oz==0 swap=xiangling;
actions+=burst target=fischl if=.buff.fischl.oz==0;
actions+=skill target=fischl if=.buff.fischl.oz==0;
actions+=attack target=xingqiu active=xingqiu;
actions+=attack target=xiangling active=xiangling;
actions+=attack target=bennett active=bennett;
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
