package terraform

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/config"
)

// EvalDeleteOutput is an EvalNode implementation that deletes an output
// from the state.
type EvalDeleteOutput struct {
	Name string
}

// TODO: test
func (n *EvalDeleteOutput) Eval(ctx EvalContext) (interface{}, error) {
	state, lock := ctx.State()
	if state == nil {
		return nil, nil
	}

	// Get a write lock so we can access this instance
	lock.Lock()
	defer lock.Unlock()

	// Look for the module state. If we don't have one, create it.
	mod := state.ModuleByPath(ctx.Path())
	if mod == nil {
		return nil, nil
	}

	delete(mod.Outputs, n.Name)

	return nil, nil
}

// EvalWriteOutput is an EvalNode implementation that writes the output
// for the given name to the current state.
type EvalWriteOutput struct {
	Name      string
	Sensitive bool
	Value     *config.RawConfig
	// ContinueOnErr allows interpolation to fail during Input
	ContinueOnErr bool
}

// TODO: test
func (n *EvalWriteOutput) Eval(ctx EvalContext) (interface{}, error) {
	// This has to run before we have a state lock, since interpolation also
	// reads the state
	cfg, err := ctx.Interpolate(n.Value, nil)
	// handle the error after we have the module from the state

	state, lock := ctx.State()
	if state == nil {
		return nil, fmt.Errorf("cannot write state to nil state")
	}

	// Get a write lock so we can access this instance
	lock.Lock()
	defer lock.Unlock()
	// Look for the module state. If we don't have one, create it.
	mod := state.ModuleByPath(ctx.Path())
	if mod == nil {
		mod = state.AddModule(ctx.Path())
	}

	// handling the interpolation error
	if err != nil {
		if n.ContinueOnErr || flagWarnOutputErrors {
			log.Printf("[ERROR] Output interpolation %q failed: %s", n.Name, err)
			// if we're continuing, make sure the output is included, and
			// marked as unknown
			mod.Outputs[n.Name] = &OutputState{
				Type:  "string",
				Value: config.UnknownVariableValue,
			}
			return nil, EvalEarlyExitError{}
		}
		return nil, err
	}

	// Get the value from the config
	var valueRaw interface{} = config.UnknownVariableValue
	if cfg != nil {
		var ok bool
		valueRaw, ok = cfg.Get("value")
		if !ok {
			valueRaw = ""
		}
		if cfg.IsComputed("value") {
			// replace computed with fully qualified variable name
			if len(mod.Path) > 1 {
				switch valueRaw.(type) {
				case string:
					valueRaw = fmt.Sprintf("${module.%s.%s", strings.Join(mod.Path[1:], "."), valueRaw.(string)[2:])
				case []interface{}:
					for i, v := range valueRaw.([]interface{}) {
						valueRaw.([]interface{})[i] = fmt.Sprintf("${module.%s.%s", strings.Join(mod.Path[1:], "."), v.(string)[2:])
					}
				case map[string]interface{}:
					for k, v := range valueRaw.(map[string]interface{}) {
						valueRaw.(map[string]interface{})[k] = fmt.Sprintf("${module.%s.%s", strings.Join(mod.Path[1:], "."), v.(string)[2:])
					}
				case []map[string]interface{}:
					for i, v := range valueRaw.([]map[string]interface{}) {
						for k, _v := range v {
							valueRaw.([]map[string]interface{})[i][k] = fmt.Sprintf("${module.%s.%s", strings.Join(mod.Path[1:], "."), _v.(string)[2:])
						}
					}
				}
			}
		}
	}

	switch valueTyped := valueRaw.(type) {
	case string:
		mod.Outputs[n.Name] = &OutputState{
			Type:      "string",
			Sensitive: n.Sensitive,
			Value:     valueTyped,
		}
	case []interface{}:
		mod.Outputs[n.Name] = &OutputState{
			Type:      "list",
			Sensitive: n.Sensitive,
			Value:     valueTyped,
		}
	case map[string]interface{}:
		mod.Outputs[n.Name] = &OutputState{
			Type:      "map",
			Sensitive: n.Sensitive,
			Value:     valueTyped,
		}
	case []map[string]interface{}:
		// an HCL map is multi-valued, so if this was read out of a config the
		// map may still be in a slice.
		if len(valueTyped) == 1 {
			mod.Outputs[n.Name] = &OutputState{
				Type:      "map",
				Sensitive: n.Sensitive,
				Value:     valueTyped[0],
			}
			break
		}
		return nil, fmt.Errorf("output %s type (%T) with %d values not valid for type map",
			n.Name, valueTyped, len(valueTyped))
	default:
		return nil, fmt.Errorf("output %s is not a valid type (%T)\n", n.Name, valueTyped)
	}

	return nil, nil
}
