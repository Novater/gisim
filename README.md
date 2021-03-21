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
- ICD for aura application not yet tracked
- status such as gouba, ganyu lotus etc... should be kept track of under Sim
- not sure how to implement Amos bow yet? maybe add post arrow fire hook? [could just add an initial frame to snapshot, then diff between that and current frame would be travel time]
- crimson witch does not track stacks yet

## todo list

- [ ] add fischl E only (up to c4?)
  - [ ] skill
- [ ] add xingqiu up to c6
  - [ ] e
  - [ ] q
- [ ] refactor weapon and artifact equip code into character; this way weapons can keep track of their own internal icd etc..
- [ ] finish implementing ganyu (up to c6?)
  - [ ] normal attack
  - [ ] aimed attack
  - [ ] aimed attack lvl 1
  - [ ] plunge attack
- [ ] implement xiangling
  - [ ] pyro application
  - [ ] charged attack
  - [ ] plunge attack
  - [ ] xiangling a4 (when to pick up?)
- [ ] jump/dash/char switch/burst/skill force reset all char normal counter
- [ ] reactions
- [ ] add ningguang
- [ ] resonance
- [ ] priority based rotation
- [ ] rotation conditions
- [ ] field effects
- [ ] frames returned per action should have at least 2 number, avg cancellable and avg normal; may actually required more than 2, 1 into each trailling action such as swap, dash, jump, burst, skill, auto

## done

- [x] melt reaction
- [x] character orb receiving function
- [x] sim handle orb
- [x] refactor character into an interface, let the character itself decide what to use to keep track of stuff
- [x] rotations to check CD
- [x] swap should trigger a 150f cooldown
- Ganyu
  - [x] cooldown on abil
  - [x] energy use on abil
  - [x] particles on skill
  - [x] C1: ffa regen 2 energy/ ffa reduce target cryo res
  - [x] C2: extra charge
- Xiangling
  - [x] normal attack
  - [x] guoba
  - [x] pyronado
  - [x] c1
  - [x] c2
  - [x] c4
  - [x] c6

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
