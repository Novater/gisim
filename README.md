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
```

`sequence` and `sequence_rest` not yet properly implemented

## optimizations

- [ ] change stat look ups to switch instead of maps

## credits

- Most of the % data: https://genshin.honeyhunterworld.com/
- Tons of discussions on KeqingMain (to be added up with ppl's discord tags at some point)
