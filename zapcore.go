package zapbugsnag

import (
	"errors"

	"github.com/bugsnag/bugsnag-go"
	"go.uber.org/zap/zapcore"
)

type severity interface{}

func bugsnagSeverity(lvl zapcore.Level) severity {
	switch lvl {
	case zapcore.DebugLevel:
		return bugsnag.SeverityInfo
	case zapcore.InfoLevel:
		return bugsnag.SeverityInfo
	case zapcore.WarnLevel:
		return bugsnag.SeverityWarning
	case zapcore.ErrorLevel:
		return bugsnag.SeverityError
	case zapcore.DPanicLevel:
		return bugsnag.SeverityError
	case zapcore.PanicLevel:
		return bugsnag.SeverityError
	case zapcore.FatalLevel:
		return bugsnag.SeverityError
	default:
		// Unrecognized levels are fatal.
		return bugsnag.SeverityError
	}
}

type client interface {
	Capture(*bugsnag.Event, map[string]string) (string, chan error)
	Wait()
}

type trace struct {
	Disabled bool
}

type core struct {
	zapcore.Level
	enc zapcore.Encoder
	encoderConfig zapcore.EncoderConfig
	trace

	fields map[string]interface{}
	tags   map[string]string
}

var Core *core = nil

func newCore(cfg Configuration, enab zapcore.Level) *core {
	Core = &core{
		Level:  enab,
		trace:  cfg.Trace,
		fields: make(map[string]interface{}),
		tags:   cfg.Tags,
		encoderConfig: cfg.EncoderConfig,
	}
	return Core
}

func (c *core) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *core) Write(ent zapcore.Entry, fs []zapcore.Field) error {
	meta := bugsnag.MetaData{}
	passedError := errors.New(ent.Message)

	if fs != nil {
		for i := 0; i < len(fs); i++ {
			field := fs[i]

			if field.Type == zapcore.ErrorType {
				passedError = field.Interface.(error)
				continue
			}

			enc := zapcore.NewMapObjectEncoder()
			field.AddTo(enc)

			meta.AddStruct(field.Key,  enc.Fields)
		}
	}

	err := bugsnag.Notify(passedError, bugsnagSeverity(ent.Level), meta)
	if err != nil {
		return err
	}

	return nil
}

func (c *core) Sync() error {
	return nil
}

func (c *core) With(fs []zapcore.Field) zapcore.Core {
	return c.with(fs)
}

func (c *core) with(fs []zapcore.Field) *core {
	// Copy our map.
	m := make(map[string]interface{}, len(c.fields))
	for k, v := range c.fields {
		m[k] = v
	}

	// Add fields to an in-memory encoder.
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range fs {
		f.AddTo(enc)
	}

	// Merge the two maps.
	for k, v := range enc.Fields {
		m[k] = v
	}

	return &core{
		Level:  c.Level,
		trace:  c.trace,
		fields: m,
	}
}
