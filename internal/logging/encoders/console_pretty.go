package encoders

import (
	"bytes"

	"github.com/fatih/color"
	"github.com/nwidger/jsoncolor"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"

	"github.com/spacelift-io/spacelift-operator/internal/logging"
)

var (
	bufferpool = buffer.NewPool()
	levels     = func(level zapcore.Level) string {
		switch -level {
		case logging.Level4:
			return color.MagentaString("DEBUG")
		case logging.Level5:
			return color.MagentaString("TRACE")
		default:
			return color.BlueString("INFO")
		}
	}
)

type prettyConsoleEncoder struct {
	formatter      *jsoncolor.Formatter
	consoleEncoder zapcore.Encoder
	zapcore.Encoder
	*zapcore.EncoderConfig
}

func NewPrettyConsoleEncoder(cfg zapcore.EncoderConfig) *prettyConsoleEncoder {
	if cfg.ConsoleSeparator == "" {
		cfg.ConsoleSeparator = "\t"
	}
	cfg.EncodeLevel = func(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(levels(level))
	}

	jsonConfig := cfg
	jsonConfig.TimeKey = ""
	jsonConfig.LevelKey = ""
	jsonConfig.NameKey = ""
	jsonConfig.CallerKey = ""
	jsonConfig.MessageKey = ""
	jsonConfig.StacktraceKey = ""

	consoleConfig := cfg
	consoleConfig.StacktraceKey = ""
	consoleConfig.NameKey = ""
	consoleConfig.MessageKey = ""

	f := jsoncolor.NewFormatter()
	f.NumberColor = color.New(color.FgCyan)
	f.TrueColor = color.New(color.FgYellow)
	f.FalseColor = f.TrueColor
	return &prettyConsoleEncoder{
		Encoder:        zapcore.NewConsoleEncoder(jsonConfig),
		consoleEncoder: zapcore.NewConsoleEncoder(consoleConfig),
		EncoderConfig:  &cfg,
		formatter:      f,
	}
}

func (e *prettyConsoleEncoder) Clone() zapcore.Encoder {
	return &prettyConsoleEncoder{
		Encoder:        e.Encoder.Clone(),
		consoleEncoder: e.consoleEncoder.Clone(),
		EncoderConfig:  e.EncoderConfig,
		formatter:      e.formatter,
	}
}

func (e *prettyConsoleEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	line, err := e.consoleEncoder.EncodeEntry(ent, nil)
	if err != nil {
		return line, err
	}
	line.TrimNewline()
	line.AppendString(e.ConsoleSeparator)

	if ent.LoggerName != "" {
		line.AppendString(color.New(color.FgCyan).Sprintf("[%s]", ent.LoggerName))
		line.AppendByte(' ')
	}

	line.AppendString(ent.Message)

	jsonRawBuf, err := e.Encoder.EncodeEntry(ent, fields)
	if err != nil {
		return line, err
	}
	defer jsonRawBuf.Free()

	var b bytes.Buffer
	err = e.formatter.Format(&b, jsonRawBuf.Bytes())
	if err != nil {
		return line, err
	}
	line.AppendString(e.ConsoleSeparator)
	line.AppendString(b.String())

	if ent.Stack != "" && e.StacktraceKey != "" {
		line.AppendByte('\n')
		line.AppendString(ent.Stack)
	}

	if e.LineEnding != "" {
		line.AppendString(e.LineEnding)
	} else {
		line.AppendString(zapcore.DefaultLineEnding)
	}
	return line, err
}
