//go:build era_wasm
// +build era_wasm

package plugins

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/balaji-balu/margo-hello-world/internal/era/lifecycle"
	"github.com/balaji-balu/margo-hello-world/pkg/era"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// WasmPlugin implements lifecycle.RuntimePlugin
type WasmPlugin struct {
	runtime  wazero.Runtime
	modules  map[string]api.Module
}

func NewRuntimePlugin() lifecycle.RuntimePlugin {
	return &WasmPlugin{
		runtime: wazero.NewRuntime(context.Background()),
		modules: map[string]api.Module{},
	}
}

func (w *WasmPlugin) Install(c era.ComponentSpec) error {
	// For WASM, “install” just checks the file exists
	if _, err := os.Stat(c.Artifact); err != nil {
		return fmt.Errorf("wasm file not found: %v", err)
	}
	return nil
}

func (w *WasmPlugin) Start(c era.ComponentSpec) error {
	ctx := context.Background()
	wasmBytes, err := os.ReadFile(c.Artifact)
	if err != nil {
		return err
	}

	mod, err := w.runtime.Instantiate(ctx, wasmBytes)
	if err != nil {
		return err
	}

	w.modules[c.Name] = mod

	// Call exported "run" function if exists
	if runFn := mod.ExportedFunction("run"); runFn != nil {
		go func() {
			_, err := runFn.Call(ctx)
			if err != nil {
				fmt.Println("WASM run error:", err)
			}
		}()
	}
	return nil
}

func (w *WasmPlugin) Stop(name string) error {
	if mod, ok := w.modules[name]; ok {
		mod.Close(context.Background())
		delete(w.modules, name)
	}
	return nil
}

func (w *WasmPlugin) Remove(name string) error {
	return w.Stop(name)
}

func (w *WasmPlugin) Status(name string) era.ComponentStatus {
	_, running := w.modules[name]
	state := "Stopped"
	if running {
		state = "Running"
	}
	return era.ComponentStatus{
		Name:      name,
		State:     state,
		Message:   "wasm runtime",
		Timestamp: time.Now().Unix(),
	}
}
