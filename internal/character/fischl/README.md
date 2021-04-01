# Fischl

## Todo

- a2 (are we ever going to sim aiming at bird?)
- aimed mode no charge
- aimed mode charge
- plunge attack
- c2 aoe (single target mode only right now)
- c4 (characters dont take dmg atm)

## Assumptions

- burst dmg is not tagged re ICD since nothing else uses it right now

## Issues/Bugs

- Fischl ICD not tested yet
- Fischl turbo interaction not implemented

## Notes

muakasan: Finding:
Determining Fischl's Elemental skill's ICD. Fischl's E (Oz) will apply electro every 4 hits or when a timer reaches 5 seconds after the first electro application. When the 5 second timer reaches zero, the oz's next hit will apply electro regardless of the hit counter. When the timer reaches 0 the hit counter will be reset and a new 5 sec timer will begin. This is very similar to auto attack ICDs (https://library.keqingmains.com/mechanics/combat/elemental-reactions/internal-cooldown-of-elemental-application), except instead of 3 auto attacks it is 4 oz hits, and instead of a 2.5 sec timer, it is a 5 sec timer. 

ICD tags:

no cooldown
Fischl	Nightrider (Summoning)	Element Skill	None	Fischl
Fischl	Nightrider (Summoning with Devourer of All Sins)	Element Skill	None	Fischl


Fischl	Nightrider (Per Hit by Oz Attacks)	Element Skill	Element Skill	Fischl
Fischl	Midnight Phantasmagoria (Per Hit by Oz Attacks)	Element Skill	Element Skill	Fischl
Fischl	Undone Be Thy Sinful Hex	Element Skill	Element Skill	Fischl
Fischl	Evernight Raven	Element Skill	Element Skill	Fischl <- C6


summon + c4
Fischl	Midnight Phantasmagoria (Falling Thunder)	Element Burst	Element Burst	Fischl
Fischl	Her Pilgrimage of Bleak	Element Burst	Element Burst	Fischl


normal
Fischl	Gaze of the Deep	Nomral Attack 1	None	Default



not doing
Fischl	Stellar Predator	Charged Attack	None	Default
