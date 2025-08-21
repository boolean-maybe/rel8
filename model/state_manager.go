package model

import (
	"context"
	"rel8/db"
	"sync"
)

type StateTransition struct {
	From  State
	To    State
	IsPop bool
}

type StateChangeCallback func(transition StateTransition)

type ContextualStateManager struct {
	mu            sync.RWMutex
	stateStack    []State
	callbacks     []StateChangeCallback
	syncCallbacks []StateChangeCallback
	maxHistory    int
	server        db.DatabaseServer
}

func NewContextualStateManager(server db.DatabaseServer, maxHistory int) *ContextualStateManager {
	return &ContextualStateManager{
		stateStack:    []State{},
		callbacks:     make([]StateChangeCallback, 0),
		syncCallbacks: make([]StateChangeCallback, 0),
		maxHistory:    maxHistory,
		server:        server,
	}
}

func (csm *ContextualStateManager) AddCallback(callback StateChangeCallback) {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	csm.callbacks = append(csm.callbacks, callback)
}

func (csm *ContextualStateManager) AddSyncCallback(callback StateChangeCallback) {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	csm.syncCallbacks = append(csm.syncCallbacks, callback)
}

func (csm *ContextualStateManager) PushState(ctx context.Context, newState State) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var oldState State
	if len(csm.stateStack) == 0 {
		oldState = newState
	} else {
		oldState = csm.stateStack[len(csm.stateStack)-1]
	}

	csm.stateStack = append(csm.stateStack, newState)

	// Limit history size
	if len(csm.stateStack) > csm.maxHistory {
		csm.stateStack = csm.stateStack[1:]
	}

	// Notify callbacks
	transition := StateTransition{From: oldState, To: newState, IsPop: false}

	// Call synchronous callbacks first (like UI updates)
	for _, callback := range csm.syncCallbacks {
		callback(transition)
	}

	// Then call async callbacks (like logging)
	for _, callback := range csm.callbacks {
		go callback(transition) // Non-blocking callbacks
	}

	return nil
}

func (csm *ContextualStateManager) PopState(ctx context.Context) (State, error) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if len(csm.stateStack) == 1 {
		currentState := csm.stateStack[len(csm.stateStack)-1]
		return currentState, nil
		//todo need to call callbacks??
	}

	currentState := csm.stateStack[len(csm.stateStack)-1]
	csm.stateStack = csm.stateStack[:len(csm.stateStack)-1]
	previousState := csm.stateStack[len(csm.stateStack)-1]

	// Notify callbacks
	transition := StateTransition{From: currentState, To: previousState, IsPop: true}

	// Call synchronous callbacks first (like UI updates)
	for _, callback := range csm.syncCallbacks {
		callback(transition)
	}

	// Then call async callbacks (like logging)
	for _, callback := range csm.callbacks {
		go callback(transition)
	}

	return previousState, nil
}

func (csm *ContextualStateManager) GetCurrentState() State {
	csm.mu.RLock()
	defer csm.mu.RUnlock()
	return csm.stateStack[len(csm.stateStack)-1]
}

func (csm *ContextualStateManager) GetHistory() []State {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	// Return a copy to prevent external modification
	history := make([]State, len(csm.stateStack))
	copy(history, csm.stateStack)
	return history
}

// GetServer returns the database server instance
func (csm *ContextualStateManager) GetServer() db.DatabaseServer {
	csm.mu.RLock()
	defer csm.mu.RUnlock()
	return csm.server
}
