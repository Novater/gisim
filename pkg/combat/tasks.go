package combat

import (
	"math/rand"
	"strings"
	"time"
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
	for k, a := range s.tasks {
		if a.Delay == 0 {
			s.Log.Debugf("\t [%v] executing task %v, originated from frame %v", s.Frame(), k, a.originFrame)
			a.F(s)
			delete(s.tasks, k)
		} else {
			a.Delay--
			s.tasks[k] = a
		}
	}
}

func (s *Sim) AddTask(f TaskFunc, name string, delay int) {
	key := RandStringBytesMaskImprSrcSB(10, name)
	s.tasks[key] = Task{
		Name:        name,
		Delay:       delay,
		F:           f,
		originFrame: s.f,
	}
	s.Log.Debugf("\t task added: %v", key)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrcSB(n int, name string) string {
	sb := strings.Builder{}
	sb.Grow(n + len(name))
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
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
