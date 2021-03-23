# gisim

## instructions

download Go for your operating system

clone the repo; `cd gisim`

either `go build` and then run the executable (in command line) or use `go run main.go`

Flags:

| flag | acceptable            | info                                                         |
| ---- | --------------------- | ------------------------------------------------------------ |
| d    | `debug` `info` `warn` | level out log output, defaults to `warn`                     |
| p    | `whatever.yaml`       | config file to use; defaults to `config.yaml`                |
| s    | `100`                 | number representing seconds to run sim for, defaults to `60` |

Example `go run main.go -d=debug -s=20`

Typically I would run the command in bash with output piped to a file for debugging purposes

`go run main.go -d=debug -s=60 &> out.log`

## bugs/issues

- there should be an animation lock if normal attacks are not chained
- sim assumes A4 is available even if lvl specified has not unlocked A4
- status such as gouba, ganyu lotus etc... should be kept track of under Sim in order for us to have action conditions
- not sure how to implement Amos bow yet? maybe add post arrow fire hook? [could just add an initial frame to snapshot, then diff between that and current frame would be travel time]
- crimson witch does not track stacks yet
- current overload formula seems to give higher damage than actual; actual 2482, got 2594.3517401129357
- we don't track self auras. this is a problem for swirl w elemental absorption
- frozen interaction isn't quite accurate. need better understanding of elemental system before this can be improved. However the current implement SHOULD result in lower than max theoritical damage (as frozen is removed completely so you can't have multiple reacts and the duration is probably a bit shorter than actual) so it's acceptable

## todo list

- [ ] aura ICD
- [ ] xingqiu
- [ ] EC testing
- [ ] fischl
- [ ] refactor weapon and artifact equip code into character; this way weapons can keep track of their own internal icd etc..
- [ ] jump/dash/char switch/burst/skill force reset all char normal counter
- [ ] ningguang
- [ ] crystallize
- [ ] sucrose
- [ ] swirl
- [ ] resonance
- [ ] rotation conditions
- [ ] field effects
- [ ] frames returned per action should have at least 2 number, avg cancellable and avg normal; may actually required more than 2, 1 into each trailling action such as swap, dash, jump, burst, skill, auto

## on auras

tinfoil hat theory:

- auras are stored in a collection (array perhaps)
- when a new element is applied, it is applied to the existing auras in this list in order
- example:

  - crystallize on EC => electro only
  - pyro (supervape bug) on EC => electro first
  - shatter on frozen => frozen first
  - pyro on frozen? i suspect this will trigger melt THEN vaporize

- so geo and pyro both react with electro first but ice reacted first w
- looks like hydro mage reapplys hydro at regular intervals.. so it could just be which ever has the longer gauge left reacts first
- looks like single target usually that's going to be electro if you're constantly reapplying it
- if you melt a frozen 1A with 2B there's nothing left :( i guess you have to melt with a lower threshold?

Apparently also this...

Artesians, Hu Tao: Zhongli/Xiangling/Hu Tao have seperate CA ICD
[10:34 PM] Artesians, Hu Tao: uhhh
[10:35 PM] Artesians, Hu Tao: all catalysts have separate CA ICD

## brainstorm - OUT DATED

hooks should have both the trigger character, and the sim data

type of hooks

- ones that affect the character
  modify char hp, char stats
- ones that affect the snapshot
  modify snapshot stats, snapshot element
- ones that affect the field
  constructs
- ones that affect the unit

actions:

- can read everything in the sim including char, unit, field

field

- geo constructs
  albedo E
  zhongli E
  ningguang E
  geo mc E

char.statMods <- coded as action that adds to this and checked every tick

- Bennett ult
  register an action
  on tick:
  check if hp needs to be healed
  remove stat buff from non active char if present
  check if hp > 70, if so, add buff to active char stat buff
- Albedo A4
  register an action
  on register, add buff to all char
  on expiry, remove buff from all char
- Jean Q
  same as bennet ult
  how to infuse with elements??

onAuraApplication

afterAuraApplication

- fischl A4

onAttack? where to put this hook?

- fischl c1

onDamage

- Chongyun E? Somehow convert all physical dmg for spear/sword/2h sword to cryo

afterDamage

- Amber weak point proc / crescent proc; should specify if use weak point in settings
- Albedo E -> triggers additional action
- Beidou burst discharge

onOrb

- Barbara A4 E cd reduction

Implemented in the skill

- Amber C1
- Barbara e on hit healing
