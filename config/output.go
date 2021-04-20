package config

import (
	"context"

	"github.com/tsaikd/KDGoLib/errutil"
	"golang.org/x/sync/errgroup"
	"log-aggregation/config/goglog"
	"log-aggregation/config/logevent"
)

// errors
var (
	ErrorUnknownOutputType1 = errutil.NewFactory("unknown output config type: %q")
	ErrorInitOutputFailed1  = errutil.NewFactory("initialize output module failed: %v")
)

// TypeOutputConfig is interface of output module
type TypeOutputConfig interface {
	TypeCommonConfig
	Output(ctx context.Context, event logevent.LogEvent) (err error)
}

// OutputConfig is basic output config struct
type OutputConfig struct {
	CommonConfig
}

// OutputHandler is a handler to regist output module
type OutputHandler func(ctx context.Context, raw *ConfigRaw) (TypeOutputConfig, error)

var (
	mapOutputHandler = map[string]OutputHandler{}
)

// RegistOutputHandler regist a output handler
func RegistOutputHandler(name string, handler OutputHandler) {
	mapOutputHandler[name] = handler
}

// GetOutputs get outputs from config
func GetOutputs(ctx context.Context, outputRaw []ConfigRaw) (outputs []TypeOutputConfig, err error) {
	var output TypeOutputConfig
	for _, raw := range outputRaw {
		handler, ok := mapOutputHandler[raw["type"].(string)]
		if !ok {
			return outputs, ErrorUnknownOutputType1.New(nil, raw["type"])
		}

		if output, err = handler(ctx, &raw); err != nil {
			return outputs, ErrorInitOutputFailed1.New(err, raw)
		}

		outputs = append(outputs, output)
	}
	return
}

func (t *Config) getOutputs() (outputs []TypeOutputConfig, err error) {
	return GetOutputs(t.ctx, t.OutputRaw)
}

func (t *Config) startOutputs() (err error) {
	outputs, err := t.getOutputs()
	if err != nil {
		return
	}

	t.eg.Go(func() error {
		for {
			select {
			case <-t.ctx.Done():
				if len(t.chFilterOut) < 1 {
					return nil
				}
			case event := <-t.chFilterOut:
				eg, ctx := errgroup.WithContext(t.ctx)
				for _, output := range outputs {
					func(output TypeOutputConfig) {
						eg.Go(func() error {
							if err2 := output.Output(ctx, event); err2 != nil {
								goglog.Logger.Errorf("output module %q failed: %v\n", output.GetType(), err2)
							}
							return nil
						})
					}(output)
				}
				if err := eg.Wait(); err != nil {
					return err
				}
				if t.chOutDebug != nil {
					t.chOutDebug <- event
				}
			}
		}
	})

	return
}
