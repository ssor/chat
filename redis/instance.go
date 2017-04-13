package redis

import (
	"github.com/ssor/log"
	redigo "github.com/ssor/redigo/redis"
)

// Instance is a redis pool instance
type Instance struct {
	*redigo.Pool
}

func NewInstance(host string) *Instance {
	redisPool := &redigo.Pool{
		MaxIdle:   80,
		MaxActive: 12000, // max number of connections
		Dial: func() (redigo.Conn, error) {
			c, err := redigo.Dial("tcp", host)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}

	return &Instance{redisPool}
}

func (instance *Instance) LoadScript(script *Script) error {
	conn := instance.Get()
	defer conn.Close()

	return script.Load(conn)
}

func (instance *Instance) NewScript(keyCount int, src string) *Script {
	return &Script{redigo.NewScript(keyCount, src)}
}

func (instance *Instance) DoScript(script *Script, keysAndArgs ...interface{}) (interface{}, error) {
	if instance == nil {
		panic("redis instance nil")
	}
	conn := instance.Get()
	defer conn.Close()

	return script.Do(conn, keysAndArgs...)
}

func (instance *Instance) Do(cmd string, args ...interface{}) (interface{}, error) {
	conn := instance.Get()
	defer conn.Close()
	res, err := conn.Do(cmd, args...)
	return res, err
}

func (instance *Instance) DoMulti(cmds *Commands) error {
	if cmds.flush == true {
		return instance.flushCmds(cmds)
	}
	return instance.doCmdOneByOne(cmds)
}

func (instance *Instance) doCmdOneByOne(cmds *Commands) error {
	conn := instance.Get()
	defer conn.Close()
	for _, cmd := range cmds.cmds {
		_, err := conn.Do(cmd.CommandName, cmd.Args...)
		if err != nil {
			log.SysF("doCmdOneByOne (%s %s) error: %s \r\n", cmd.CommandName, cmd.Args, err)
			return err
		}
		log.TraceF("doCmdOneByOne (%s %s) ", cmd.CommandName, cmd.Args)
	}
	return nil
}

func (instance *Instance) flushCmds(cmds *Commands) error {
	conn := instance.Get()
	defer conn.Close()
	for _, cmd := range cmds.cmds {
		err := conn.Send(cmd.CommandName, cmd.Args...)
		if err != nil {
			log.SysF("flushCmds (%s %s) error: %s \r\n", cmd.CommandName, cmd.Args, err)
			return err
		}
		log.TraceF("flushCmds (%s %s) ", cmd.CommandName, cmd.Args)
	}
	err := conn.Flush()
	if err != nil {
		log.SysF("flushCmds flush error: %s \r\n", err)
		return err
	}
	return nil
}
