//go:build js && wasm

package meta

import (
	"encoding/json"
	"errors"
	"fmt"
	"syscall/js"
)

const wasmSaveKey = "girlbusAG_save_v2"

func getLocalStorage() (js.Value, bool) {
	ls := js.Global().Get("localStorage")
	if ls.IsUndefined() || ls.IsNull() || ls.Type() != js.TypeObject {
		return js.Undefined(), false
	}
	return ls, true
}

func readLocalStorage(key string) (raw string, ok bool) {
	defer func() {
		if recover() != nil {
			raw = ""
			ok = false
		}
	}()

	ls, valid := getLocalStorage()
	if !valid {
		return "", false
	}
	val := ls.Call("getItem", key)
	if val.IsUndefined() || val.IsNull() || val.Type() != js.TypeString {
		return "", false
	}
	return val.String(), true
}

// Load reads SaveData from browser localStorage, returning defaults if unavailable or corrupt.
func Load() SaveData {
	data := DefaultSaveData()
	raw, ok := readLocalStorage(wasmSaveKey)
	if !ok || raw == "" {
		return data
	}

	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return DefaultSaveData()
	}
	data.Normalize()
	return data
}

// Save normalizes and writes SaveData to browser localStorage.
func Save(data SaveData) (err error) {
	data.Normalize()

	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("localStorage unavailable: %v", r)
		}
	}()

	ls, valid := getLocalStorage()
	if !valid {
		return errors.New("localStorage unavailable")
	}
	ls.Call("setItem", wasmSaveKey, string(raw))
	return nil
}
