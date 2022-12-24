package logs

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/rs/zerolog"
)

type Event struct {
	logger     *Logger
	level      zerolog.Level
	fields     map[string]interface{}
	message    string
	err        error
	ctx        context.Context
	filename   string
	lineNumber int
}

func (event *Event) Value(key string, v interface{}) *Event {
	event.fields[key] = v
	return event
}

func (event *Event) Err(err error) *Event {
	event.err = err
	return event
}

func (event *Event) Context(ctx context.Context) *Event {
	event.ctx = ctx
	return event
}

func (event *Event) Caller(fieventname string, line int) *Event {
	event.filename = fieventname
	event.lineNumber = line
	return event
}

func (event *Event) Msg(msg string) {
	event.message = msg
	event.flush()
}

func (event *Event) Msgf(msg string, a ...interface{}) {
	event.message = fmt.Sprintf(msg, a...)
	event.flush()
}

func (event Event) flush() {
	if event.logger == nil {
		return
	}

	formattedMsg := event.message
	if event.err != nil {
		formattedMsg = event.err.Error() + ": " + formattedMsg
	}

	event.logger.logger.WithLevel(event.level).Fields(map[string]interface{}{
		"timestamp": zerolog.TimestampFunc().UTC().Format(time.RFC3339),
		"content": map[string]interface{}{
			"level":      event.level.String(),
			"message":    formattedMsg,
			"filename":   event.filename,
			"linenumber": event.lineNumber,
		},
	}).Msg(formattedMsg)

}

func Debug() *Event {
	return caller(logs, zerolog.DebugLevel)
}

func Info() *Event {
	return caller(logs, zerolog.InfoLevel)
}

func Warn() *Event {
	return caller(logs, zerolog.WarnLevel)
}

func Error() *Event {
	return caller(logs, zerolog.ErrorLevel)
}

func caller(logger *Logger, level zerolog.Level) *Event {
	_, fievent, line, _ := runtime.Caller(2)
	logs := &Event{
		logger: logger,
		level:  level,
		fields: map[string]interface{}{},
	}
	return logs.Caller(fievent, line)
}
