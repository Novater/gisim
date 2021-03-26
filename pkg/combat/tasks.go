package combat

type Task struct {
	Name  string
	F     TaskFunc
	Delay int
}

func (t Task) String() string {
	return t.Name
}

type TaskFunc func(s *Sim)

func (s *Sim) runTasks() {
	next := make([]Task, 0, len(s.tasks))
	for _, a := range s.tasks {
		if a.Delay == 0 {
			a.F(s)
		} else {
			a.Delay--
			next = append(next, a)
		}
	}
	s.tasks = next
}

func (s *Sim) AddTask(f TaskFunc, name string, delay int) {
	s.tasks = append(s.tasks, Task{
		Name:  name,
		Delay: delay,
		F:     f,
	})
}
