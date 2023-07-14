package bot

import (
	"fmt"
	"reflect"
	"strconv"
)

type Action struct {
	Raw     interface{}
	Wrapper func(...interface{}) (string, error)
}

type Routine struct {
	Params   []string
	Action   Action
	Result   string
	ErrorMsg string
}

func NewRoutine(action Action) *Routine {
	return &Routine{
		Action: action,
	}
}

func (cmd *Routine) Execute(args []string) (string, error) {
	castArgs, err := cmd.CastArgs(args)
	if err != nil {
		return "", err
	}

	return cmd.Action.Wrapper(castArgs...)
}

func (cmd *Routine) CastArgs(args []string) ([]interface{}, error) {
	fnType := reflect.TypeOf(cmd.Action.Raw)
	numParams := fnType.NumIn()
	if numParams != len(args) {
		return nil, fmt.Errorf("wrong number of args. Given: %d, Takes: %d", len(args), numParams)
	}
	castParams := make([]interface{}, numParams)
	for i := 0; i < numParams; i++ {
		paramType := fnType.In(i)
		switch paramType.Kind() {
		case reflect.Int:
			asInt, err := strconv.Atoi(args[i])
			if err != nil {
				fmt.Println("Wrong type")
				return nil, fmt.Errorf("wrong type of args")
			}
			castParams[i] = asInt
		case reflect.String:
			castParams[i] = args[i]
		case reflect.Float64:
			asFloat, err := strconv.ParseFloat(args[i], 64)
			if err != nil {
				fmt.Println("Wrong type")
				return nil, fmt.Errorf("wrong type of args")
			}
			castParams[i] = asFloat
		case reflect.Float32:
			asFloat, err := strconv.ParseFloat(args[i], 32)
			if err != nil {
				fmt.Println("Wrong type")
				return nil, fmt.Errorf("wrong type of args")
			}
			castParams[i] = asFloat
		default:
			fmt.Println("default hit")
			return nil, fmt.Errorf("function has unsupported param type")
		}
	}

	return castParams, nil
}
