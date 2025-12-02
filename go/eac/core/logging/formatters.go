package logging

import (
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// rawEncoder outputs only the message - no timestamp, level, or fields
// Perfect for clean CLI output
type rawEncoder struct {
	zapcore.Encoder
}

// newRawEncoder creates an encoder that outputs only the log message
func newRawEncoder() zapcore.Encoder {
	return &rawEncoder{
		Encoder: zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			MessageKey: "msg",
		}),
	}
}

// EncodeEntry outputs just the message followed by newline
func (e *rawEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf := buffer.NewPool().Get()
	buf.AppendString(entry.Message)
	buf.AppendString("\n")
	return buf, nil
}

// Clone creates a copy of the encoder
func (e *rawEncoder) Clone() zapcore.Encoder {
	return &rawEncoder{Encoder: e.Encoder.Clone()}
}

// timestampedEncoder outputs "HH:MM:SS.mmm  LEVEL  module:message"
type timestampedEncoder struct {
	zapcore.Encoder
	module string
}

// newTimestampedEncoder creates an encoder with timestamp, level, and module prefix
func newTimestampedEncoder(module string) zapcore.Encoder {
	cfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		EncodeTime:     zapcore.TimeEncoderOfLayout("15:04:05.000"),
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	return &timestampedEncoder{
		Encoder: zapcore.NewConsoleEncoder(cfg),
		module:  module,
	}
}

// EncodeEntry outputs timestamped format: "HH:MM:SS.mmm  LEVEL  module:message"
func (e *timestampedEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf := buffer.NewPool().Get()

	// Time: HH:MM:SS.mmm
	buf.AppendString(entry.Time.Format("15:04:05.000"))
	buf.AppendString("  ")

	// Level: padded to 7 chars (DEBUG, INFO, WARN, ERROR)
	levelStr := entry.Level.CapitalString()
	buf.AppendString(levelStr)
	// Pad to 7 characters
	for i := len(levelStr); i < 7; i++ {
		buf.AppendByte(' ')
	}
	buf.AppendString(" ")

	// Module and message
	if e.module != "" {
		buf.AppendString(e.module)
		buf.AppendString(":")
	}
	buf.AppendString(entry.Message)
	buf.AppendString("\n")

	return buf, nil
}

// Clone creates a copy of the encoder
func (e *timestampedEncoder) Clone() zapcore.Encoder {
	return &timestampedEncoder{
		Encoder: e.Encoder.Clone(),
		module:  e.module,
	}
}

// newJSONEncoder creates a standard JSON encoder for file logging
func newJSONEncoder() zapcore.Encoder {
	cfg := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	return zapcore.NewJSONEncoder(cfg)
}

// CreateEncoder creates an encoder based on the formatter type
func CreateEncoder(formatter FormatterType, module string) zapcore.Encoder {
	switch formatter {
	case FormatterRaw:
		return newRawEncoder()
	case FormatterTimestamped:
		return newTimestampedEncoder(module)
	case FormatterJSON:
		return newJSONEncoder()
	default:
		// Default to raw for console-like behavior
		return newRawEncoder()
	}
}
