package stdlib_fact

import (
	"fmt"
	"strings"

	"github.com/bonkzero404/uddin-lang/interpreter"
)

type factModule struct{}

func (m *factModule) Name() string { return "fact" }

func (m *factModule) Functions() map[string]interpreter.ModuleFunc {
	return map[string]interpreter.ModuleFunc{
		"assert":  factAssertFunc,
		"retract": factRetractFunc,
		"query":   factQueryFunc,
		"exists":  factExistsFunc,
		"count":   factCountFunc,
		"clear":   factClearFunc,
		"get_all": factGetAllFunc,
	}
}

func init() {
	interpreter.RegisterModule(&factModule{})
}

// factAssertFunc adds a fact to the knowledge base.
func factAssertFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 {
		panic(interpreter.TypeErrorf(pos, "fact.assert() requires at least 2 arguments (category, key, [value])"))
	}
	category, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "fact.assert(): first argument must be a category string"))
	}
	key, ok := args[1].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "fact.assert(): second argument must be a key string"))
	}

	db := ctx.FactDB()
	db.Lock()
	defer db.Unlock()

	fullKey := category + ":" + key
	if len(args) == 2 {
		db.Set(fullKey, true)
	} else {
		db.Set(fullKey, args[2])
	}
	return interpreter.Value(true)
}

// factRetractFunc removes a fact from the knowledge base.
func factRetractFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 1 {
		return interpreter.Value(fmt.Errorf("fact.retract() requires at least 1 argument (category, [key])"))
	}
	category, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fact.retract(): first argument must be a category string"))
	}

	db := ctx.FactDB()
	db.Lock()
	defer db.Unlock()

	if len(args) == 1 {
		prefix := category + ":"
		count := int64(0)
		for k := range db.All() {
			if strings.HasPrefix(k, prefix) {
				db.Delete(k)
				count++
			}
		}
		return interpreter.Value(count)
	}

	key, ok := args[1].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fact.retract(): second argument must be a key string"))
	}
	fullKey := category + ":" + key
	if _, exists := db.Get(fullKey); exists {
		db.Delete(fullKey)
		return interpreter.Value(true)
	}
	return interpreter.Value(false)
}

// factQueryFunc queries facts from the knowledge base.
func factQueryFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 1 {
		return interpreter.Value(fmt.Errorf("fact.query() requires at least 1 argument (category, [key])"))
	}
	category, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fact.query(): first argument must be a category string"))
	}

	db := ctx.FactDB()
	db.RLock()
	defer db.RUnlock()

	if len(args) == 1 {
		prefix := category + ":"
		resultMap := make(map[string]interpreter.Value)
		for k, v := range db.All() {
			if strings.HasPrefix(k, prefix) {
				actualKey := k[len(prefix):]
				resultMap[actualKey] = interpreter.Value(v)
			}
		}
		return interpreter.Value(&resultMap)
	}

	key, ok := args[1].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fact.query(): second argument must be a key string"))
	}
	fullKey := category + ":" + key
	if value, exists := db.Get(fullKey); exists {
		return interpreter.Value(value)
	}
	return interpreter.Value(nil)
}

// factExistsFunc checks if a fact exists in the knowledge base.
func factExistsFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 {
		return interpreter.Value(fmt.Errorf("fact.exists() requires 2 arguments (category, key)"))
	}
	category, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fact.exists(): first argument must be a category string"))
	}
	key, ok := args[1].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fact.exists(): second argument must be a key string"))
	}

	db := ctx.FactDB()
	db.RLock()
	defer db.RUnlock()

	fullKey := category + ":" + key
	_, exists := db.Get(fullKey)
	return interpreter.Value(exists)
}

// factCountFunc returns the number of facts in a category or total.
func factCountFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	db := ctx.FactDB()
	db.RLock()
	defer db.RUnlock()

	if len(args) == 0 {
		return interpreter.Value(len(db.All()))
	}

	category, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "fact.count(): first argument must be a category string"))
	}
	prefix := category + ":"
	count := 0
	for k := range db.All() {
		if strings.HasPrefix(k, prefix) {
			count++
		}
	}
	return interpreter.Value(count)
}

// factClearFunc clears all facts or facts in a specific category.
func factClearFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	db := ctx.FactDB()
	db.Lock()
	defer db.Unlock()

	if len(args) == 0 {
		count := int64(len(db.All()))
		for k := range db.All() {
			db.Delete(k)
		}
		return interpreter.Value(count)
	}

	category, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fact.clear(): first argument must be a category string"))
	}
	prefix := category + ":"
	count := int64(0)
	for k := range db.All() {
		if strings.HasPrefix(k, prefix) {
			db.Delete(k)
			count++
		}
	}
	return interpreter.Value(count)
}

// factGetAllFunc returns all facts in the knowledge base.
func factGetAllFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	db := ctx.FactDB()
	db.RLock()
	defer db.RUnlock()

	results := make(map[string]interpreter.Value)
	for k, v := range db.All() {
		results[k] = interpreter.Value(v)
	}
	return interpreter.Value(&results)
}
