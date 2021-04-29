package parse

import (
	"log"
	"testing"

	"github.com/srliao/gisim/pkg/combat"
)

var s = `
##XIANGLING
char+=xiangling ele=pyro lvl=80 hp=9638 atk=200 def=593 cr=0.05 cd=0.50 em=72 cons=6 talent=6,6,8;
weapon+=xiangling label="skyward spine" atk=674 refine=1 er=0.368;
art+=xiangling label="gladiator's finale" count=2;
art+=xiangling label="crimson witch of flames" count=2;
stats+=xiangling label=flower hp=4780 hp%=.047 atk%=.152 em=21 cd=.264;
stats+=xiangling label=feather atk=311 def=37 def%=.057 cr=.097 cd=.218;
stats+=xiangling label=sands atk%=0.466 cd=.124 cr=.078 hp=209 atk=51;
stats+=xiangling label=goblet pyro%=.466 cr=.089 er=0.052 atk%=.093 atk=43;
stats+=xiangling label=circlet cr=.311 atk%=.087 hp%=0.058 cd=.132 def=76;
##XINGQIU
char+=xingqiu ele=hydro lvl=70 hp=8352 atk=165 def=619 cr=0.05 cd=0.50 atk%=.18 cons=6 talent=1,8,8;
weapon+=xingqiu label="sacrificial sword" atk=401 refine=3 er=.559;
art+=xingqiu label="gladiator's finale" count=2;
art+=xingqiu label="noblesse oblige" count=2;
stats+=xingqiu label=flower hp=4780 def=44 er=.065 cr=.097 cd=.124;
stats+=xingqiu label=feather atk=311 cd=.218 def=19 atk=.117 em=40;
stats+=xingqiu label=sands atk%=0.466 cd=.124 def%=.175 er=.045 hp=478;
stats+=xingqiu label=goblet hydro%=.466 cd=.202 atk=.14 hp=299 atk=39;
stats+=xingqiu label=circlet cr=.311 cd=0.062 atk%=.192 hp%=.082 atk=39;
##BENNETT
char+=bennett ele=pyro lvl=70 hp=10129 atk=156 def=630 cr=0.05 cd=0.50 er=.2 cons=2 talent=2,3,6;
weapon+=bennett label="festering desire" atk=510 er=0.459 refine=5;
art+=bennett label="noblesse oblige" count=4;
stats+=bennett label=flower hp=3967 atk=45 cd=.148 def=39 atk%=.058;
stats+=bennett label=feather atk=258 atk%=.117 def=16 er=.104 em=42;
stats+=bennett label=sands er=.43 atk%=.163 hp%=.058 cd=.117 hp=209;
stats+=bennett label=goblet pyro%=.387 hp=657 atk=19 cd=.14 cr=.035;
stats+=bennett label=circlet cr=.232 cd=.056 atk=28 em=30 def%=.053;
##FISCHL
char+=fischl ele=electro lvl=70 hp=7508 atk=200 def=485 cr=0.05 cd=0.50 atk%=.18 cons=4 talent=4,8,2;
weapon+=fischl label="favonius warbow" atk=401 refine=5;
art+=fischl label="gladiator's finale" count=2;
art+=fischl label="wanderer's troupe" count=2;
stats+=fischl label=flower hp=4780 cd=.109 atk%=.087 cr=.109 atk=33;
stats+=fischl label=feather atk=311 cd=.179 er=.091 atk%=.058 cr=.062;
stats+=fischl label=sands atk%=0.387 em=35 cr=.039 atk=49 hp=508;
stats+=fischl label=goblet electro%=.466 atk%=.105 em=61 cr=.027 er=.11;
stats+=fischl label=circlet cr=.258 atk%=.041 atk=53 er=.058 hp=568;
##ENEMY
target+="blazing axe mitachurl" lvl=88 pyro=0.1 dendro=0.1 hydro=0.1 electro=0.1 geo=0.1 anemo=0.1 physical=.3;
active+=xingqiu;
##ROTATION
actions+=sequence_strict target=xingqiu exec=skill,burst lock=100;
actions+=skill target=xingqiu if=.status.xingqiu.energy<80 lock=100;
actions+=burst target=xingqiu;
actions+=burst target=bennett;
actions+=sequence_strict target=xiangling exec=skill,burst;
actions+=skill target=xiangling active=xiangling;
actions+=burst target=fischl if=.status.xiangling.energy<70&&.tags.fischl.oz==0 swap=xiangling;
actions+=skill target=fischl if=.status.xiangling.energy<70&&.tags.fischl.oz==0 swap=xiangling;
actions+=burst target=fischl if=.tags.fischl.oz==0;
actions+=skill target=fischl if=.tags.fischl.oz==0;
actions+=skill target=bennett if=.status.xiangling.energy<40 swap=xiangling;
actions+=skill target=bennett;
actions+=attack target=xiangling;
actions+=attack target=xingqiu active=xingqiu;
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
	if err != nil {
		t.Error(err)
	}

}
