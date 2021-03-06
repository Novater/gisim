package combat

import (
	"strings"
)

type Task struct {
	Name        string
	F           TaskFunc
	Delay       int
	originFrame int
}

func (t Task) String() string {
	return t.Name
}

type TaskFunc func(s *Sim)

func (s *Sim) runTasks() {

	for _, t := range s.tasks[s.F] {
		t.F(s)
	}

	delete(s.tasks, s.F)
}

func (s *Sim) AddTask(f TaskFunc, name string, delay int) {
	s.tasks[s.F+delay] = append(s.tasks[s.F+delay], Task{
		Name:        name,
		Delay:       delay,
		F:           f,
		originFrame: s.F,
	})
	s.Log.Debugf("\t task added: %v", name)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func (s *Sim) RandStringBytesMaskImprSrcSB(n int, name string) string {
	sb := strings.Builder{}
	sb.Grow(n + len(name))
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, s.Rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = s.Rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	sb.WriteString(name)

	return sb.String()
}
