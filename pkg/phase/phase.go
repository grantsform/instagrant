package phase

import (
	"github.com/grantios/instagrant/pkg/config"
	"github.com/grantios/instagrant/pkg/logger"
	"github.com/grantios/instagrant/pkg/util"
)

// Context holds the execution context for phases
type Context struct {
	Config    *config.Config
	Logger    *logger.Logger
	DryRun    bool
	TargetDir string // Mount point for the new system (usually /mnt)
	State     *State
	Exec      *util.Executor    // Command executor
	Helper    *util.PhaseHelper // Common operation helpers
}

// Phase represents a single installation step
type Phase interface {
	// Name returns the unique identifier for this phase
	Name() string
	
	// Description returns a human-readable description
	Description() string
	
	// Execute runs the phase
	Execute(ctx *Context) error
	
	// Validate checks if the phase can run
	Validate(ctx *Context) error
	
	// IsChroot returns true if this phase must run in chroot
	IsChroot() bool
}

// BasePhase provides default implementations for common phase methods
type BasePhase struct {
	name        string
	description string
	isChroot    bool
}

// NewBasePhase creates a new base phase with the given properties
func NewBasePhase(name, description string, isChroot bool) BasePhase {
	return BasePhase{
		name:        name,
		description: description,
		isChroot:    isChroot,
	}
}

func (b BasePhase) Name() string {
	return b.name
}

func (b BasePhase) Description() string {
	return b.description
}

func (b BasePhase) IsChroot() bool {
	return b.isChroot
}

func (b BasePhase) Validate(ctx *Context) error {
	return nil // Default: no validation
}

// State tracks installation progress
type State struct {
	CurrentPhase string
	Completed    []string
	Failed       bool
	ErrorMessage string
}

// Registry holds all available phases
type Registry struct {
	phases []Phase
}

// NewRegistry creates a new phase registry
func NewRegistry() *Registry {
	return &Registry{
		phases: make([]Phase, 0),
	}
}

// Register adds one or more phases to the registry
func (r *Registry) Register(phases ...Phase) {
	r.phases = append(r.phases, phases...)
}

// GetPhases returns all registered phases in order
func (r *Registry) GetPhases() []Phase {
	return r.phases
}

// GetPhase returns a phase by name
func (r *Registry) GetPhase(name string) Phase {
	for _, p := range r.phases {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

// GetPhasesFrom returns phases starting from the named phase
func (r *Registry) GetPhasesFrom(name string) []Phase {
	for i, p := range r.phases {
		if p.Name() == name {
			return r.phases[i:]
		}
	}
	return nil
}
