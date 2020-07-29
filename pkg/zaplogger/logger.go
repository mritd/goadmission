package zaplogger

import (
	"log"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var confOnce sync.Once

func NewLogger(c *zapConfig) *zap.Logger {
	syncer := zapcore.AddSync(os.Stdout)
	if c.sample {
		c.opts = append(c.opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
		}))
	}
	c.opts = append(c.opts, zap.AddCallerSkip(1), zap.ErrorOutput(syncer))
	return zap.New(zapcore.NewCore(c.encoder, syncer, c.level)).WithOptions(c.opts...)
}

func New(name string) *zap.Logger {
	return NewLogger(config).Named(name)
}

func NewSugar(name string) *zap.SugaredLogger {
	return New(name).Sugar()
}

func Setup() {
	confOnce.Do(func() {
		zc, err := NewConfig(Config)
		if err != nil {
			log.Fatalf("Failed to create zap logger: %v", err)
		}
		config = zc
	})
}
