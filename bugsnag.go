package zapbugsnag

import (
	"github.com/bugsnag/bugsnag-go"
	"go.uber.org/zap/zapcore"
)

// Configuration is a minimal set of parameters for Sentry integration.
type Configuration struct {
	bugsnag.Configuration
	Tags            map[string]string
	Trace           trace
	MinimalLogLevel *zapcore.Level
}

// Build uses the provided configuration to construct a Sentry-backed logging
// core.
func (c Configuration) Build() (zapcore.Core, error) {
	if Core != nil {
		return Core, nil
	}

	bugsnag.Configure(c.Configuration)
	bugsnag.OnBeforeNotify(removeLoggerStackFrames)

	minimalLogLevel := zapcore.ErrorLevel
	if c.MinimalLogLevel != nil {
		minimalLogLevel = *c.MinimalLogLevel
	}

	return newCore(c, minimalLogLevel), nil
}

func removeLoggerStackFrames(event *bugsnag.Event, config *bugsnag.Configuration) error {
	event.Stacktrace = event.Stacktrace[3:]

	return nil
}
