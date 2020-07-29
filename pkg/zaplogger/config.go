package zaplogger

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

const (
	EncoderConsole = "console"
	EncoderJSON    = "json"
)

const (
	TimeEncoderISO8601 = "iso8601"
	TimeEncoderMillis  = "millis"
	TimeEncoderNanos   = "nano"
	TimeEncoderEpoch   = "epoch"
	TimeEncoderDefault = "default"
)

var Config ZapConfig
var config *zapConfig

type ZapConfig struct {
	Development  bool   `json:"development,omitempty" yaml:"development,omitempty"`
	Encoder      string `json:"encoder,omitempty" yaml:"encoder,omitempty"`
	Level        string `json:"level,omitempty" yaml:"level,omitempty"`
	StackLevel   string `json:"stack_level,omitempty" yaml:"stack_level,omitempty"`
	Sample       bool   `json:"sample,omitempty" yaml:"sample,omitempty"`
	TimeEncoding string `json:"time_encoding,omitempty" yaml:"time_encoding,omitempty"`
}

type zapConfig struct {
	level      zap.AtomicLevel
	stackLevel zapcore.Level
	encoder    zapcore.Encoder
	opts       []zap.Option
	sample     bool
}

type encoderConfigFunc func(*zapcore.EncoderConfig)
type encoderFunc func(...encoderConfigFunc) zapcore.Encoder

func NewConfig(c ZapConfig) (*zapConfig, error) {
	var zc zapConfig
	var eFunc encoderFunc

	// If development is enabled, use the default development config;
	// otherwise, use the default production config.
	if c.Development {
		eFunc, _ = getEncoder(EncoderConsole)
		zc.level = zap.NewAtomicLevelAt(zap.DebugLevel)
		zc.opts = append(zc.opts, zap.Development())
		zc.sample = false
		zc.stackLevel = zap.WarnLevel
	} else {
		eFunc, _ = getEncoder(EncoderJSON)
		zc.level = zap.NewAtomicLevelAt(zap.InfoLevel)
		zc.sample = true
		zc.stackLevel = zap.ErrorLevel
	}

	// If Level is set, override the default Level
	if c.Level != "" {
		lvl, err := getLevel(c.Level)
		if err != nil {
			return nil, err
		}
		zc.level = zap.NewAtomicLevelAt(lvl)
	}

	// If StackLevel is set, override the default StackLevel
	if c.StackLevel != "" {
		lvl, err := getLevel(c.StackLevel)
		if err != nil {
			return nil, err
		}
		zc.stackLevel = lvl
	}
	zc.opts = append(zc.opts, zap.AddStacktrace(zc.stackLevel))

	// If Encoder is set, override the default Encoder
	if c.Encoder != "" {
		f, err := getEncoder(c.Encoder)
		if err != nil {
			return nil, err
		}
		eFunc = f
	}

	// Set TimeEncoding, use "2006-01-02 15:04:05" by default
	var ecFuncs []encoderConfigFunc
	if c.TimeEncoding != "" {
		tec, err := getTimeEncoder(c.TimeEncoding)
		if err != nil {
			return nil, err
		}
		ecFuncs = append(ecFuncs, withTimeEncoding(tec))
	} else {
		tec, _ := getTimeEncoder(TimeEncoderDefault)
		ecFuncs = append(ecFuncs, withTimeEncoding(tec))
	}
	zc.encoder = eFunc(ecFuncs...)

	zc.sample = c.Sample
	if zc.level.Level() <= -1 {
		zc.sample = false
	}

	return &zc, nil
}

func getLevel(l string) (zapcore.Level, error) {
	lower := strings.ToLower(l)
	var lvl zapcore.Level
	switch lower {
	case LevelDebug:
		lvl = zapcore.DebugLevel
	case LevelInfo:
		lvl = zapcore.InfoLevel
	case LevelWarn:
		lvl = zapcore.WarnLevel
	case LevelError:
		lvl = zapcore.ErrorLevel
	default:
		return lvl, fmt.Errorf("invalid log level \"%s\"", l)
	}
	return lvl, nil
}

func getEncoder(ec string) (encoderFunc, error) {
	lower := strings.ToLower(ec)
	switch lower {
	case EncoderConsole:
		return func(ecfs ...encoderConfigFunc) zapcore.Encoder {
			encoderConfig := zap.NewDevelopmentEncoderConfig()
			for _, f := range ecfs {
				f(&encoderConfig)
			}
			return zapcore.NewConsoleEncoder(encoderConfig)
		}, nil
	case EncoderJSON:
		return func(ecfs ...encoderConfigFunc) zapcore.Encoder {
			encoderConfig := zap.NewProductionEncoderConfig()
			for _, f := range ecfs {
				f(&encoderConfig)
			}
			return zapcore.NewJSONEncoder(encoderConfig)
		}, nil
	default:
		return nil, fmt.Errorf("invalid encoder \"%s\"", ec)
	}
}

func getTimeEncoder(tec string) (zapcore.TimeEncoder, error) {
	lower := strings.ToLower(tec)
	switch lower {
	case TimeEncoderISO8601:
		return zapcore.ISO8601TimeEncoder, nil
	case TimeEncoderMillis:
		return zapcore.EpochMillisTimeEncoder, nil
	case TimeEncoderNanos:
		return zapcore.EpochNanosTimeEncoder, nil
	case TimeEncoderEpoch:
		return zapcore.EpochTimeEncoder, nil
	case TimeEncoderDefault:
		return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		}, nil
	default:
		return nil, fmt.Errorf("invalid time encoder \"%s\"", tec)
	}
}

func withTimeEncoding(tec zapcore.TimeEncoder) encoderConfigFunc {
	return func(ec *zapcore.EncoderConfig) {
		ec.EncodeTime = tec
	}
}
