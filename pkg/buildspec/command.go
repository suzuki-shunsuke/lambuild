package buildspec

import (
	"errors"
	"fmt"

	"github.com/suzuki-shunsuke/lambuild/pkg/expr"
)

type Command struct {
	Command string
	If      expr.Bool
}

type Commands []Command

func (commands *Commands) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var val []interface{}
	if err := unmarshal(&val); err != nil {
		return err
	}
	arr := make(Commands, len(val))
	for i, a := range val {
		if s, ok := a.(string); ok {
			arr[i] = Command{
				Command: s,
			}
			continue
		}
		b, ok := a.(map[interface{}]interface{})
		if !ok {
			return errors.New("command must be string or map[interface{}]interface{}")
		}
		cmd := Command{}
		for k, v := range b {
			ks, ok := k.(string)
			if !ok {
				return errors.New("key of command must be string")
			}
			switch ks {
			case "command":
				c, ok := v.(string)
				if !ok {
					return errors.New("command of command must be string")
				}
				cmd.Command = c
			case "if":
				vs, ok := v.(string)
				if !ok {
					return errors.New("if of buildspec must be string")
				}
				prog, err := expr.NewBool(vs)
				if err != nil {
					return fmt.Errorf("compile an expression: %s: %w", vs, err)
				}
				cmd.If = prog
			default:
				return errors.New("invalid key of command: " + ks)
			}
		}
		arr[i] = cmd
	}
	*commands = arr
	return nil
}

func (commands *Commands) Filter(param interface{}) ([]string, error) {
	cmds := []string{}
	for _, command := range *commands {
		if command.If.Empty() {
			cmds = append(cmds, command.Command)
			continue
		}
		f, err := command.If.Run(param)
		if err != nil {
			return nil, fmt.Errorf("evaluate command.if: %w", err)
		}
		if f {
			cmds = append(cmds, command.Command)
			continue
		}
	}
	return cmds, nil
}
