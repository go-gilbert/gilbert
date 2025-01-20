package runner

import (
	"fmt"
)

type HandlerNotFoundError struct {
	ActionName string
}

func NewHandlerNotFoundError(actionName string) *HandlerNotFoundError {
	return &HandlerNotFoundError{ActionName: actionName}
}

func (err *HandlerNotFoundError) Error() string {
	return fmt.Sprintf("no such action handler: %q", err.ActionName)
}

// HandlerResolver is abstract handlers registry.
type HandlerResolver interface {
	// GetHandler resolves action handler by name.
	GetHandler(actionName string) (HandlerFactory, error)
}

// HandlerSet is a static handler resolver.
type HandlerSet struct {
	handlers ActionHandlers
}

func NewHandlerSet(handlers ActionHandlers) *HandlerSet {
	return &HandlerSet{
		handlers: handlers,
	}
}

// HandleFunc registers action handler for specified action name
//
// Returns an error if action is already handled by another handler
func (r *HandlerSet) HandleFunc(actionName string, handler HandlerFactory) error {
	if _, ok := r.handlers[actionName]; ok {
		return fmt.Errorf("action handler %q is already registered", actionName)
	}

	r.handlers[actionName] = handler
	return nil
}

// GetHandler returns handler for specified action
//
// Returns an error if no action handler was registered
func (r *HandlerSet) GetHandler(actionName string) (HandlerFactory, error) {
	handler, ok := r.handlers[actionName]
	if !ok {
		return nil, NewHandlerNotFoundError(actionName)
	}

	return handler, nil
}
