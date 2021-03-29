package combat

type snapshotHookType string

const (
	PostSnapshot   snapshotHookType = "POST_SNAPSHOT"
	PreDamageHook  snapshotHookType = "PRE_DAMAGE"
	PreReaction    snapshotHookType = "PRE_REACTION"
	PostDamageHook snapshotHookType = "POST_DAMAGE"
	OnCritDamage   snapshotHookType = "CRIT_DAMAGE"
	PostReaction   snapshotHookType = "POST_REACTION"
)

type eventHookType string

const (
	PreBurstHook  eventHookType = "PRE_BURST_HOOK"
	PostBurstHook eventHookType = "POSt_BURST_HOOK"
)

type snapshotHookFunc func(s *Snapshot) bool
type eventHookFunc func(s *Sim) bool

//AddCombatHook adds a hook to sim. Hook will be called based on the type of hook
func (s *Sim) AddSnapshotHook(f snapshotHookFunc, key string, hook snapshotHookType) {
	if _, ok := s.snapshotHooks[hook]; !ok {
		s.snapshotHooks[hook] = make(map[string]snapshotHookFunc)
	}
	s.snapshotHooks[hook][key] = f
	s.Log.Debugf("\t[%v] new snapshot hook added %v", s.Frame(), key)
}

//AddHook adds a hook to sim. Hook will be called based on the type of hook
func (s *Sim) AddEventHook(f eventHookFunc, key string, hook eventHookType) {
	if _, ok := s.eventHooks[hook]; !ok {
		s.eventHooks[hook] = make(map[string]eventHookFunc)
	}
	s.eventHooks[hook][key] = f
	s.Log.Debugf("\t[%v] new event hook added %v", s.Frame(), key)
}

func (s *Sim) AddEffect(f EffectFunc, key string) {
	s.effects = append(s.effects, f)
	s.Log.Debugf("\t[%v] new effect added %v", s.Frame(), key)
}

func (s *Sim) executeEventHook(t eventHookType) {
	for k, f := range s.eventHooks[t] {
		if f(s) {
			s.Log.Debugf("[%v] event hook %v returned true, deleting", s.Frame(), k)
			delete(s.eventHooks[t], k)
		}
	}
}

func (s *Sim) executeSnapshotHooks(t snapshotHookType, ds *Snapshot) {
	for k, f := range s.snapshotHooks[t] {
		if f(ds) {
			s.Log.Debugf("[%v] event hook %v returned true, deleting", s.Frame(), k)
			delete(s.snapshotHooks[t], k)
		}
	}
}

func (s *Sim) runEffects() {
	next := make([]EffectFunc, 0, len(s.effects))
	for _, f := range s.effects {
		if !f(s) {
			next = append(next, f)
		}
	}
	s.effects = next
}
