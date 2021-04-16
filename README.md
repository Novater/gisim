# gisim

## instructions

download Go for your operating system

clone the repo; `cd gisim`

either `go build` and then run the executable (in command line) or use `go run main.go`

Flags:

| flag | acceptable            | info                                                          |
| ---- | --------------------- | ------------------------------------------------------------- |
| d    | `debug` `info` `warn` | level out log output, defaults to `warn`                      |
| p    | `whatever.yaml`       | config file to use; defaults to `config.yaml`                 |
| s    | any number            | number representing seconds to run sim for, defaults to `600` |
| o    | any string            | log file to write to; if blank no log. defaults to no log     |
| c    | `true/false`          | whether or not print caller function in debug log             |

Example `go run main.go -d=debug -s=20`

To save the log to file (easier to read) `go run main.go -d=debug -o=out.log`

## bugs/issues

- there should be an animation lock if normal attacks are not chained
- sim assumes A4 is available even if lvl specified has not unlocked A4
- status such as gouba, ganyu lotus etc... should be kept track of under Sim in order for us to have action conditions
- not sure how to implement Amos bow yet? maybe add post arrow fire hook? [could just add an initial frame to snapshot, then diff between that and current frame would be travel time]
- current overload formula seems to give higher damage than actual; actual 2482, got 2594.3517401129357
- we don't track self auras. this is a problem for swirl w elemental absorption
- swirl EC is incorrect; currently only trigger swirlelectro but need to trigger both; need to restructure reacitontype to take an array... but that becomes a mess :(
- anemo resonance reduced CD not yet implemented

## todo list

- [ ] play test xl/xq/ben/fish w/ basic weapons
- [ ] change normal reset to a frame number and if s.F = frame, then reset
- [ ] rand artifact code is bugged
- [ ] change runEffects back to a map; loops can break if one effect adds another effect
- [ ] aura ICD
- [ ] jump/dash/char switch/burst/skill force reset all char normal counter
- [ ] frames returned per action should have at least 2 number, avg cancellable and avg normal; may actually required more than 2, 1 into each trailling action such as swap, dash, jump, burst, skill, auto

## rotation syntax ideas

action list example

```
actions+=/burst,char=diluc

#reset attack chain if E goes on CD
actions+=/restart_sequence,name=attackp,if=diluc.cd.skill>0
actions+=/restart_sequence,name=attack,if=diluc.cd.skill>0

#execute full chain if skill is not yet used and burst buff (pyro) is on
actions+=/sequence_strict,name=attackp,char=diluc,auto:skill:auto:skill:auto:skill:auto,if=diluc.buff.burst=1&diluc.status.skill_count=0

#otherwise just execute through the chain like normal
actions+=/sequence,name=attack,char=diluc,auto:skill:auto:skill:auto:skill:auto

#autoattack as filler
actions+=/attack,char=diluc
```

## optimizations

- [ ] change stat look ups to switch instead of maps

## credits

- Most of the % data: https://genshin.honeyhunterworld.com/
- Tons of discussions on KeqingMain (to be added up with ppl's discord tags at some point)
