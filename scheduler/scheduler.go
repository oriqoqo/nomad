package scheduler

import (
	"fmt"

	"github.com/hashicorp/nomad/nomad/structs"
)

// BuiltinSchedulers contains the built in registered schedulers
// which are available
var BuiltinSchedulers = map[string]Factory{}

// NewScheduler is used to instantiate and return a new scheduler
// given the scheduler name, initial state, and planner.
func NewScheduler(name string, state State, planner Planner) (Scheduler, error) {
	// Lookup the factory function
	factory, ok := BuiltinSchedulers[name]
	if !ok {
		return nil, fmt.Errorf("unknown scheduler '%s'", name)
	}

	// Instantiate the scheduler
	sched := factory(state, planner)
	return sched, nil
}

// Factory is used to instantiate a new Scheduler
type Factory func(State, Planner) Scheduler

// Scheduler is the top level instance for a scheduler. A scheduler is
// meant to only encapsulate business logic, pushing the various plumbing
// into Nomad itself. They are invoked to process a single evaluation at
// a time. The evaluation may result in task allocations which are computed
// optimistically, as there are many concurrent evaluations being processed.
// The task allocations are submitted as a plan, and the current leader will
// coordinate the commmits to prevent oversubscription or improper allocations
// based on stale state.
type Scheduler interface {
	// Process is used to handle a new evaluation. The scheduler is free to
	// apply any logic necessary to make the task placements. The state and
	// planner will be provided prior to any invocations of process.
	Process(*structs.Evaluation) error
}

// State is an immutable view of the global state. This allows schedulers
// to make intelligent decisions based on allocations of other schedulers
// and to enforce complex constraints that require more information than
// is available to a local state scheduler.
type State interface {
	// Nodes returns an iterator over all the nodes.
	// The type of each result is *structs.Node
	Nodes() (ResultIterator, error)
}

// ResultIterator is used to iterate over a list of results
// from a query. The type of the response depends on the query,
// and perfectly represents when generics would be useful.
type ResultIterator interface {
	Next() interface{}
}

// Planner interface is used to submit a task allocation plan.
type Planner interface {
	// SubmitPlan is used to submit a plan for consideration.
	// This will return a PlanResult or an error. It is possible
	// that this will result in a state refresh as well.
	SubmitPlan(*structs.Plan) (*structs.PlanResult, State, error)
}
