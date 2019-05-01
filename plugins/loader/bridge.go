package loader

import (
	"context"
	"encoding/json"
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

		proxyScopeCalls(s, scope)

		return nil, nil
	}
}

func proxyScopeCalls(sess *ipc.Session, scope sdk.ScopeAccessor) {
	sess.HandleFunc("scope.AppendVariables", func(args json.RawMessage) (interface{}, error) {
		var v sdk.Vars
		if err := json.Unmarshal(args, v); err != nil {
			return nil, err
		}

		scope.AppendVariables(v)
		return nil, nil
	})

	sess.HandleFunc("scope.AppendGlobals", func(args json.RawMessage) (interface{}, error) {
		var v sdk.Vars
		if err := json.Unmarshal(args, v); err != nil {
			return nil, err
		}

		scope.AppendGlobals(v)
		return nil, nil
	})

	sess.HandleFunc("scope.Global", func(args json.RawMessage) (interface{}, error) {
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

	sess.HandleFunc("scope.Var", func(args json.RawMessage) (interface{}, error) {
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

	sess.HandleFunc("scope.Vars", func(_ json.RawMessage) (interface{}, error) {
		return scope.Vars(), nil
	})

	sess.HandleFunc("scope.ExpandVariables", func(args json.RawMessage) (interface{}, error) {
		var exp string
		if err := json.Unmarshal(args, &exp); err != nil {
			return nil, err
		}
		return scope.ExpandVariables(exp)
	})

	sess.HandleFunc("scope.ExpandVariables", func(args json.RawMessage) (interface{}, error) {
		var exp string
		if err := json.Unmarshal(args, &exp); err != nil {
			return nil, err
		}
		return scope.ExpandVariables(exp)
	})

	sess.HandleFunc("scope.Scan", func(args json.RawMessage) (interface{}, error) {
		var values []*string
		if err := json.Unmarshal(args, &values); err != nil {
			return nil, err
		}

		err := scope.Scan(values...)
		return values, err
	})

	sess.HandleFunc("scope.Environment", func(_ json.RawMessage) (interface{}, error) {
		return scope.Environment(), nil
	})

	sess.HandleFunc("scope.Environ", func(_ json.RawMessage) (interface{}, error) {
		return scope.Environ(), nil
	})
}
