# gisim

See [wiki](https://github.com/srliao/gisim/wiki) for detailed explanation and instructions.

## roadmap

- [ ] skeleton code for every character (i.e. all the abilities coded but no frame count)
- [ ] implementation for all weapons and artifacts
- [ ] indepth implementation/testing for each character one at a time
- [ ] revamp input config to use custom parser and/or make yaml more readable
- [ ] add artifact simulation and stat weighting calcs

## how you can contribute

The following are some ways you can contribute, listed in order of most time consuming/to least:

1. Come up with benchmark config files i.e. for this given config file, the expected dps should be xxx. Help reconcile difference. (These benchmarks can then be used as "unit" tests as new code gets implemented to make sure nothing breaks)
2. Taking over a character's implementation in terms of gathering all necessary frame counts and detailed testing of the character including (no coding required):
   1. comparing sim damage line by line vs actual in game
   2. making sure elemental applications are applying the correct gauge at the correct time
   3. test different combos and making sure interaction behaves as expected
   4. test different weapon + artifacts used by the character and make sure interaction behaves as expected
3. Running your own character setup in the sim and compare output. Provide feedback on any discrepancy
4. Generally playing around with the sim, providing any sort of feedback/feature request/etc... 

These are just some ideas. Any bug fixes/pull requests etc... are always welcome.

## change log

- 04/19/2021: Diluc implemented (not tested)

## cpu profiling

CPU profilling is enabled by default for optimization purposes. If you wish to check it out, run ` go tool pprof -pdf cpu.pprof >| file.pdf` after running the sim (best run the sim for a longer period of time to generate useful cpu profile)


## misc todo

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
- [ ] use int for map keys instead of strings

## credits

- Most of the % data: https://genshin.honeyhunterworld.com/
- Tons of discussions on KeqingMain (to be added up with ppl's discord tags at some point)
