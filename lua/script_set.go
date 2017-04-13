package lua

import (
	"errors"
	"fmt"
	"log"

	"github.com/ssor/chat/redis"
)

type luaScriptSet struct {
	Scripts map[string]*redis.Script
}

func newLuaScriptSet() *luaScriptSet {
	return &luaScriptSet{
		Scripts: make(map[string]*redis.Script),
	}
}

func (lss *luaScriptSet) Add(name string, keyCount int, src string) error {
	script := redis.NewScript(keyCount, src)
	return lss.AddScript(name, script)
}
func (lss *luaScriptSet) AddScript(name string, script *redis.Script) error {
	if _, exists := lss.Scripts[name]; exists {
		return errors.New("alreay exists")
	}
	lss.Scripts[name] = script

	return nil
}

func (lss *luaScriptSet) Load(loader func(*redis.Script) error) error {

	for name, script := range lss.Scripts {
		err := loader(script)
		if err != nil {
			return fmt.Errorf("load script %s err: %s", name, err.Error())
		}
		log.Println("load script  ", name, " ", script.Hash())
	}

	return nil
}
