# gisim

## instructions

See [wiki](https://github.com/srliao/gisim/wiki)

## change log

- 04/19/2021: Diluc implemented (not tested)

## cpu profiling

CPU profilling is enabled by default for optimization purposes. If you wish to check it out, run ` go tool pprof -pdf cpu.pprof >| file.pdf` after running the sim (best run the sim for a longer period of time to generate useful cpu profile)

## todo priority list

- [ ] on initial freeze app, both hydro + cryo durability should be reduced = min(remaining dur, new app dur); this should always result in only one "old" aura left; this "old" aura is the one that shows up when shattered
- [ ] add params to weapons
- [ ] add shield status to sim
- [ ] play test xl/xq/ben/fish w/ basic weapons
- [ ] change normal reset to a frame number and if s.F = frame, then reset
- [ ] rand artifact code is bugged
- [ ] change runEffects back to a map; loops can break if one effect adds another effect
- [ ] jump/dash/char switch/burst/skill force reset all char normal counter
- [ ] frames returned per action should have at least 2 number, avg cancellable and avg normal; may actually required more than 2, 1 into each trailling action such as swap, dash, jump, burst, skill, auto

## optimizations

- [ ] change stat look ups to switch instead of maps

## credits

- Most of the % data: https://genshin.honeyhunterworld.com/
- Tons of discussions on KeqingMain (to be added up with ppl's discord tags at some point)
