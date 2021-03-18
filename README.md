# gisim

## todo list

- [ ] character orb receiving function
- [ ] sim handle orb
- [ ] finish implementing ganyu (up to c6?)

  - [ ] normal attack
  - [ ] aimed attack
  - [ ] aimed attack lvl 1
  - [ ] plunge attack
  - [ ] cooldown on abil
  - [ ] energy use on abil
  - [ ] particles on skill
  - [ ] C1: ffa regen 2 energy/ ffa reduce target cryo res
  - [ ] C2: extra charge
  - [ ]

- [ ] reactions
- [ ] add ningguang
- resonance
- priority based rotation
- rotation conditions
- field effects

## brainstorm

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

## other notes?

amos bow

- need 2 hooks:
  - after normal/charge attack if arrow
  - after normal/charge attack damage (for flight time)
