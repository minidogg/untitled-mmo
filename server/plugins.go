package main

import (
	"fmt"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

type EventBus struct {
	L        *lua.LState
	handlers map[string][]lua.LValue
}

func NewEventBus(L *lua.LState) *EventBus {
	return &EventBus{
		L:        L,
		handlers: make(map[string][]lua.LValue),
	}
}

func (b *EventBus) luaOn(L *lua.LState) int {
	event := L.CheckString(1)
	fn := L.CheckFunction(2)

	b.handlers[event] = append(b.handlers[event], fn)
	return 0
}

func (b *EventBus) RegisterAPI() {
	b.L.SetGlobal("on", b.L.NewFunction(b.luaOn))
}

func (b *EventBus) Emit(event string, args ...lua.LValue) {
	if handlers, ok := b.handlers[event]; ok {
		for _, fn := range handlers {
			err := b.L.CallByParam(lua.P{
				Fn:      fn,
				NRet:    0,
				Protect: true,
			}, args...)

			if err != nil {
				fmt.Println("lua event error:", err)
			}
		}
	}
}

type PluginManager struct {
	L   *lua.LState
	Bus *EventBus
}

func NewPluginManager() *PluginManager {
	L := lua.NewState()

	bus := NewEventBus(L)
	bus.RegisterAPI()

	return &PluginManager{
		L:   L,
		Bus: bus,
	}
}

func (pm *PluginManager) Close() {
	pm.L.Close()
}

func (pm *PluginManager) LoadPluginFile(path string) error {
	if err := pm.L.DoFile(path); err != nil {
		return fmt.Errorf("plugin %s failed: %w", path, err)
	}
	return nil
}

func (pm *PluginManager) LoadPlugins(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".lua" {
			if err := pm.LoadPluginFile(path); err != nil {
				return err
			}
		}
		return nil
	})
}
