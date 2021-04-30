package parse

import (
	"log"
	"testing"

	"github.com/srliao/gisim/pkg/combat"
)

var s = `
char+=sucrose ele=anemo lvl=60 hp=6501 atk=112 def=463 cr=0.05 cd=0.50 anemo%=.12 cons=2 talent=1,1,1;
weapon+=sucrose label="sacrificial fragments" atk=99 refine=1 em=85;

target+="blazing axe mitachurl" lvl=88 pyro=0.1 dendro=0.1 hydro=0.1 electro=0.1 geo=0.1 anemo=0.1 physical=.3;
active+=sucrose;

actions+=burst target=sucrose;
actions+=skill target=sucrose;
actions+=attack target=sucrose;
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
	log.Println("characters:")
	for _, v := range a.Characters {
		log.Println(v.Base.Name)
		//basic stats:
		log.Println("\t basics", v.Base)
		log.Println("\t weapons", v.Weapon)
		log.Println("\t talents", v.Talents)
		log.Println("\t sets", v.Sets)
		//pretty print stats
		log.Println("\t stats", combat.PrettyPrintStats(v.Stats))
	}
	log.Println("rotations:")
	for _, v := range a.Rotation {
		log.Println(v)
	}
	log.Println(a.Enemy)
	if err != nil {
		t.Error(err)
	}

}
