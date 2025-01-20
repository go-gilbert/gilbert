/*
Package actions contains operations related to action handlers
*/
package actions

import (
	"github.com/go-gilbert/gilbert/internal/actions/build"
	"github.com/go-gilbert/gilbert/internal/actions/cover"
	"github.com/go-gilbert/gilbert/internal/actions/cover/html"
	"github.com/go-gilbert/gilbert/internal/actions/pkgget"
	"github.com/go-gilbert/gilbert/internal/actions/shell"
	"github.com/go-gilbert/gilbert/internal/actions/watch"
	"github.com/go-gilbert/gilbert/internal/runner"
)

// BuiltinHandlers contains standard action handlers.
var BuiltinHandlers = runner.ActionHandlers{
	"get-package": pkgget.NewAction,
	"build":       build.NewAction,
	"shell":       shell.NewAction,
	"watch":       watch.NewAction,
	"cover":       cover.NewAction,
	"cover:html":  html.NewAction,
}
