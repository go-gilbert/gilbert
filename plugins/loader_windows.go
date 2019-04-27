package plugins

import (
	"fmt"
	"github.com/x1unix/gilbert/log"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/scope"
	"syscall"
	"unsafe"
)

const (
	nArgs          = 5 // scopePtr, paramsPtr, logPtr, resultPtr, errPtr
	newPluginProc  = "NewPlugin"
	pluginNameProc = "GetPluginName"
)

func wrapPluginDll(fnPtr uintptr) PluginFactory {
	return func(scope *scope.Scope, params manifest.RawParams, logger log.Logger) (plug Plugin, err error) {
		sPtr := (uintptr)(unsafe.Pointer(scope))
		pPtr := (uintptr)(unsafe.Pointer(&params))
		lPtr := (uintptr)(unsafe.Pointer(&logger))

		plug = nil
		err = nil
		plPtr := (uintptr)(unsafe.Pointer(&plug))
		erPtr := (uintptr)(unsafe.Pointer(&err))
		_, _, callErr := syscall.Syscall6(fnPtr, nArgs, sPtr, pPtr, lPtr, plPtr, erPtr, 0)
		if callErr != 0 {
			return nil, fmt.Errorf("failed to invoke DLL method %s(): %s", newPluginProc, err)
		}

		return plug, err
	}
}

func getDllPluginName(handle syscall.Handle) (string, error) {
	var name string
	fnPtr, err := syscall.GetProcAddress(handle, pluginNameProc)
	if err != nil {
		return name, fmt.Errorf("cannot find procedure %s() in plugin DLL (%s)", pluginNameProc, err)
	}

	_, _, callErr := syscall.Syscall(fnPtr, 1, (uintptr)(unsafe.Pointer(&name)), 0, 0)
	if callErr != 0 {
		return name, fmt.Errorf("%s() returned an error: %s", pluginNameProc, callErr)
	}

	return name, nil
}

func loadLibrary(libPath string) (PluginFactory, string, error) {
	lib, err := syscall.LoadLibrary(libPath)
	if err != nil {
		return nil, "", err
	}

	plugName, err := getDllPluginName(lib)
	if err != nil {
		return nil, plugName, err
	}

	fnPtr, err := syscall.GetProcAddress(lib, newPluginProc)
	if err != nil {
		return nil, "", fmt.Errorf("cannot find plugin entrypoint function '%s': %s", newPluginProc, err)
	}

	return wrapPluginDll(fnPtr), plugName, nil
}