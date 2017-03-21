package lua

import (
	"errors"
	"fmt"
	"log"
	"xsbPro/common"
)

type LuaScriptSet struct {
	Scripts map[string]*common.Script
}

func NewLuaScriptSet() *LuaScriptSet {
	return &LuaScriptSet{
		Scripts: make(map[string]*common.Script),
	}
}

func (lss *LuaScriptSet) Add(name string, script *common.Script) error {
	if _, exists := lss.Scripts[name]; exists {
		return errors.New("alreay exists")
	} else {
		lss.Scripts[name] = script
	}
	return nil
}

func (lss *LuaScriptSet) Load(loader func(*common.Script) error) error {

	for name, script := range lss.Scripts {
		err := loader(script)
		if err != nil {
			return fmt.Errorf("load script %s err: %s", name, err.Error())
		}
		log.Println("load script  ", name, " ", script.Hash())
	}

	return nil
}
