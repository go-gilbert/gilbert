/*
Package actions contains operations related to action handlers
*/
package actions

import (
	"github.com/go-gilbert/gilbert/internal/legacy/actions/build"
	"github.com/go-gilbert/gilbert/internal/legacy/actions/cover"
	"github.com/go-gilbert/gilbert/internal/legacy/actions/cover/html"
	"github.com/go-gilbert/gilbert/internal/legacy/actions/pkgget"
	"github.com/go-gilbert/gilbert/internal/legacy/actions/shell"
	"github.com/go-gilbert/gilbert/internal/legacy/actions/watch"
	"github.com/go-gilbert/gilbert/internal/legacy/runner"
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
