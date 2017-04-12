package redis

type Command struct {
	CommandName string
	Args        []interface{}
}

func NewCommand(cmd string, args ...interface{}) *Command {
	return &Command{
		CommandName: cmd,
		Args:        args,
	}
}

type Commands struct {
	cmds  []*Command
	flush bool
}

func NewCommands(flush bool) *Commands {
	return &Commands{
		cmds:  []*Command{},
		flush: flush,
	}
}

func (rc *Commands) AddCmd(cmds ...*Command) *Commands {
	rc.cmds = append(rc.cmds, cmds...)
	return rc
}

func (rc *Commands) Add(cmd string, args ...interface{}) *Commands {
	rc.cmds = append(rc.cmds, NewCommand(cmd, args...))
	return rc
}
