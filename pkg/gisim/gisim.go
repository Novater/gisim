package gisim

import (
	"math/rand"
	"time"

	"go.uber.org/zap"
)

type Sim struct {
	Log    *zap.SugaredLogger
	Status map[string]int

	chars   []Character
	actions []Action
}

func (s *Sim) Run(duration int) {
	rand.Seed(time.Now().UnixNano())
	skip := 0
	for f := 0; f < 60*duration; f++ {
		//execute tick functions
		s.decrementStatusDuration()
		s.executeCharacterTicks()
		s.executeActions()
		if skip > 0 {
			skip--
			continue
		}
		//find out what the next ability is
		next := s.findNextValidAbility()
		//execute action
		skip += s.executeAbility()
	}

}

type Character interface {
	Tick()
}

type Rotation struct{}

type Combo struct{}

type ComboItem struct{}

func (s *Sim) findNextValidAbility() ComboItem {
	return ComboItem{}
}

func (s *Sim) executeAbility(a ComboItem) int {
	return 0
}

type Action struct {
	F        ActionFunc
	Duration int
}

type ActionFunc func(s *Sim)

func (s *Sim) executeCharacterTicks() {
	for _, c := range s.chars {
		c.Tick()
	}
}

func (s *Sim) executeActions() {
	next := make([]Action, 0, len(s.actions))
	for _, a := range s.actions {
		if a.Duration == 0 {
			a.F(s)
		} else {
			a.Duration--
			next = append(next, a)
		}
	}
	s.actions = next
}

func (s *Sim) decrementStatusDuration() {
	for k, v := range s.Status {
		if v == 0 {
			delete(s.Status, k)
		} else {
			s.Status[k]--
		}
	}
}
