package loader

import (
	"context"
	"encoding/json"
	"errors"
	sdk "github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/log"
	"github.com/go-gilbert/gilbert/support/ipc"
)

const (
	pluginActionsProc = "GetPluginActions"
	pluginNameProc    = "GetPluginName"
)

var loadedPlugins = make(map[string]*pluginBridge)

type pluginBridge struct {
	*ipc.Client
	path        string
	cancelFn    context.CancelFunc
	mainSession *ipc.Session
}

func newPluginBridge(parentCtx context.Context, libPath string) (b *pluginBridge, err error) {
	ctx, cancelFn := context.WithCancel(parentCtx)
	client := ipc.NewClient(ctx, libPath)
	b = &pluginBridge{
		path:        libPath,
		Client:      client,
		mainSession: client.NewSession(),
		cancelFn:    cancelFn,
	}

	if err = b.Connect(); err != nil {
		return nil, err
	}

	log.Default.Debugf("bridge: start new plugin process '%s'", libPath)

	if err = b.mainSession.Open(); err != nil {
		log.Default.Debugf("bridge: connect error, killing plugin process '%s' (errors: %v)", client.Close())
		return nil, err
	}

	log.Default.Debugf("bridge: open session '%s'", b.mainSession.ID().String())
	return b, nil
}

func (b *pluginBridge) PluginName() (name string, err error) {
	err = b.mainSession.Call(&name, pluginNameProc)
	return name, err
}

func (b *pluginBridge) PluginActions() (sdk.Actions, error) {
	actions := make([]string, 0)
	if err := b.mainSession.Call(&actions, pluginActionsProc); err != nil {
		return nil, err
	}

	out := make(sdk.Actions, len(actions))
	for _, action := range actions {
		out[action] = b.wrapHandlerFactory(action)
	}

	return out, nil
}

func (b *pluginBridge) Dispose() {
	b.cancelFn()
	if err := b.Close(); err != nil {
		log.Default.Debugf("bridge: failed to kill plugin '%s', %s", b.path, err)
	}

	log.Default.Debugf("bridge: killed '%s'", b.path)
}

func (b *pluginBridge) wrapHandlerFactory(actionName string) sdk.HandlerFactory {
	return func(scope sdk.ScopeAccessor, params sdk.ActionParams) (sdk.ActionHandler, error) {
		s := b.NewSession()

		if err := s.Open(); err != nil {
			return nil, err
		}

		return &actionHandlerBridge{
			session: s,
			params:  params,
			scope:   scope,
		}, nil
	}
}

type actionHandlerBridge struct {
	session *ipc.Session
	params  sdk.ActionParams
	scope   sdk.ScopeAccessor
}

func (b *actionHandlerBridge) Call(ctx sdk.JobContextAccessor, runner sdk.JobRunner) error {
	b.proxyScopeCalls(b.scope)
	b.proxyJobRunner(ctx, runner)
	ctx.Log().Debugf("wrapper: wrap scope, job context and runner")
	return nil
}

func (b *actionHandlerBridge) Cancel(ctx sdk.JobContextAccessor) error {
	return nil
}

func (b *actionHandlerBridge) proxyContext(ctx sdk.JobContextAccessor) {

}

func (b *actionHandlerBridge) proxyJobRunner(ctx sdk.JobContextAccessor, runner sdk.JobRunner) {
	// TODO: implement
	b.session.HandleFunc("runner.RunJob", stubHandler)
	b.session.HandleFunc("runner.ActionByName", stubHandler)
}

func stubHandler(args json.RawMessage) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (b *actionHandlerBridge) proxyScopeCalls(scope sdk.ScopeAccessor) {
	b.session.HandleFunc("scope.AppendVariables", func(args json.RawMessage) (interface{}, error) {
		var v sdk.Vars
		if err := json.Unmarshal(args, v); err != nil {
			return nil, err
		}

		scope.AppendVariables(v)
		return nil, nil
	})

	b.session.HandleFunc("scope.AppendGlobals", func(args json.RawMessage) (interface{}, error) {
		var v sdk.Vars
		if err := json.Unmarshal(args, v); err != nil {
			return nil, err
		}

		scope.AppendGlobals(v)
		return nil, nil
	})

	b.session.HandleFunc("scope.Global", func(args json.RawMessage) (interface{}, error) {
		var varName string
		if err := json.Unmarshal(args, &varName); err != nil {
			return nil, err
		}

		out, ok := scope.Global(varName)
		return ipc.Values{
			"out": out,
			"ok":  ok,
		}, nil
	})

	b.session.HandleFunc("scope.Var", func(args json.RawMessage) (interface{}, error) {
		var varName string
		if err := json.Unmarshal(args, &varName); err != nil {
			return nil, err
		}

		local, out, ok := scope.Var(varName)
		return ipc.Values{
			"out":     out,
			"ok":      ok,
			"isLocal": local,
		}, nil
	})

	b.session.HandleFunc("scope.Vars", func(_ json.RawMessage) (interface{}, error) {
		return scope.Vars(), nil
	})

	b.session.HandleFunc("scope.ExpandVariables", func(args json.RawMessage) (interface{}, error) {
		var exp string
		if err := json.Unmarshal(args, &exp); err != nil {
			return nil, err
		}
		return scope.ExpandVariables(exp)
	})

	b.session.HandleFunc("scope.ExpandVariables", func(args json.RawMessage) (interface{}, error) {
		var exp string
		if err := json.Unmarshal(args, &exp); err != nil {
			return nil, err
		}
		return scope.ExpandVariables(exp)
	})

	b.session.HandleFunc("scope.Scan", func(args json.RawMessage) (interface{}, error) {
		var values []*string
		if err := json.Unmarshal(args, &values); err != nil {
			return nil, err
		}

		err := scope.Scan(values...)
		return values, err
	})

	b.session.HandleFunc("scope.Environment", func(_ json.RawMessage) (interface{}, error) {
		return scope.Environment(), nil
	})

	b.session.HandleFunc("scope.Environ", func(_ json.RawMessage) (interface{}, error) {
		return scope.Environ(), nil
	})
}
