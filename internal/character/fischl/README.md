# Fischl

## Todo

- a2 (are we ever going to sim aiming at bird?)
- normal attack
- aimed mode no charge
- aimed mode charge
- plunge attack
- c2 aoe (single target mode only right now)
- c4 (characters dont take dmg atm)

## Assumptions

- every charge attack lvl 2 applies cryo twice (arrow + bloom)
- Q only hits target once every so often (single target)
- Fischl A4 is assumed to not apply electro
- only initial application of EC will trigger Fischl A4 (need to test this, does ticks trigger A4?)
- ReactedTo field is not implemented; Fischl A4 on swirl will not work

## Issues/Bugs

- Oz ICD not implemented correctly yet
- Fischl turbo interaction not implemented

## Notes

muakasan: Finding:
Determining Fischl's Elemental skill's ICD. Fischl's E (Oz) will apply electro every 4 hits or when a timer reaches 5 seconds after the first electro application. When the 5 second timer reaches zero, the oz's next hit will apply electro regardless of the hit counter. When the timer reaches 0 the hit counter will be reset and a new 5 sec timer will begin. This is very similar to auto attack ICDs (https://library.keqingmains.com/mechanics/combat/elemental-reactions/internal-cooldown-of-elemental-application), except instead of 3 auto attacks it is 4 oz hits, and instead of a 2.5 sec timer, it is a 5 sec timer. 