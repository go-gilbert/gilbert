/*
Package actions contains operations related to action handlers
*/
package actions

import (
	"fmt"
	"sync"

	"github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/actions/build"
	"github.com/go-gilbert/gilbert/actions/cover"
	"github.com/go-gilbert/gilbert/actions/cover/html"
	"github.com/go-gilbert/gilbert/actions/pkgget"
	"github.com/go-gilbert/gilbert/actions/shell"
	"github.com/go-gilbert/gilbert/actions/watch"
)

// mutex for concurrent read from actionsHandlers map
var m sync.RWMutex

// actionsHandlers is list of known action handlers
//
// By-default it contains a list of built-in action handlers
var actionsHandlers = sdk.Actions{
	"get-package": pkgget.NewAction,
	"build":       build.NewAction,
	"shell":       shell.NewAction,
	"watch":       watch.NewAction,
	"cover":       cover.NewAction,
	"cover:html":  html.NewAction,
}

// HandleFunc registers action handler for specified action name
//
// Returns an error if action is already handled by another handler
func HandleFunc(actionName string, handler sdk.HandlerFactory) error {
	if _, actionExists := actionsHandlers[actionName]; actionExists {
		return fmt.Errorf("action handler '%s' is already registered", actionName)
	}

	actionsHandlers[actionName] = handler
	return nil
}

// GetHandler returns handler for specified action
//
// Returns an error if no action handler was registered
func GetHandler(actionName string) (sdk.HandlerFactory, error) {
	m.RLock()
	defer m.RUnlock()

	handler, ok := actionsHandlers[actionName]
	if !ok {
		return nil, fmt.Errorf("no such action handler: '%s'", actionName)
	}

	return handler, nil
}
