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
	s.snapshotHooks[hook] = append(s.snapshotHooks[hook], f)
	s.Log.Debugf("\t[%v] new snapshot hook added %v", s.Frame(), key)
}

//AddHook adds a hook to sim. Hook will be called based on the type of hook
func (s *Sim) AddEventHook(f eventHookFunc, key string, hook eventHookType) {
	s.eventHooks[hook] = append(s.eventHooks[hook], f)
	s.Log.Debugf("\t[%v] new event hook added %v", s.Frame(), key)
}

func (s *Sim) AddEffect(f EffectFunc, key string) {
	s.effects = append(s.effects, f)
	s.Log.Debugf("\t[%v] new effect added %v", s.Frame(), key)
}

func (s *Sim) executeEventHook(t eventHookType) {
	var next []eventHookFunc
	for _, f := range s.eventHooks[t] {
		if !f(s) {
			next = append(next, f)
		}
	}
	s.eventHooks[t] = next
}

func (s *Sim) executeSnapshotHooks(t snapshotHookType, ds *Snapshot) {
	var next []snapshotHookFunc
	for _, f := range s.snapshotHooks[t] {
		if !f(ds) {
			next = append(next, f)
		}
	}
	s.snapshotHooks[t] = next
}

func (s *Sim) runEffects() {
	var next []EffectFunc
	for k, f := range s.effects {
		if !f(s) {
			s.Log.Debugf("\t[%v] action %v expired", s.Frame(), k)
			next = append(next, f)
		}
	}
	s.effects = next
}
