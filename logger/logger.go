package logger

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

// FileConfig for logger
type FileConfig struct {
	// Level is the min level in this log file
	Level Level
	// Logger config for this log file
	Logger *lumberjack.Logger
}

// Logger with multi file
type Logger []logger

type logger struct {
	level  Level
	logger *zap.Logger
}

const (
	// DebugLevel logger
	DebugLevel = Level(zapcore.DebugLevel)
	// InfoLevel logger
	InfoLevel = Level(zapcore.InfoLevel)
	// WarnLevel logger
	WarnLevel = Level(zapcore.WarnLevel)
	// ErrorLevel logger
	ErrorLevel = Level(zapcore.ErrorLevel)
	// DPanicLevel logger
	DPanicLevel = Level(zapcore.DPanicLevel)
	// PanicLevel logger
	PanicLevel = Level(zapcore.PanicLevel)
	// FatalLevel logger
	FatalLevel = Level(zapcore.FatalLevel)
)

// Level for logger
type Level zapcore.Level

// InitConfig with debug, info, warn, error 4 levels
func InitConfig(filePath string, maxSize, maxBackups, maxAge int) []FileConfig {
	dl := &lumberjack.Logger{
		Filename:   filePath + ".debug.log",
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
	}
	il := &lumberjack.Logger{
		Filename:   filePath + ".info.log",
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
	}
	wl := &lumberjack.Logger{
		Filename:   filePath + ".warn.log",
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
	}
	el := &lumberjack.Logger{
		Filename:   filePath + ".error.log",
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
	}
	return []FileConfig{
		{
			Level:  DebugLevel,
			Logger: dl,
		},
		{
			Level:  InfoLevel,
			Logger: il,
		},
		{
			Level:  WarnLevel,
			Logger: wl,
		},
		{
			Level:  ErrorLevel,
			Logger: el,
		},
	}
}

// New multi level logger
func New(configs []FileConfig) *Logger {
	cores := make([]zapcore.Core, len(configs))
	for i := range configs {
		config := configs[i]
		ec := zap.NewProductionEncoderConfig()
		ec.TimeKey = "time"
		ec.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(time.RFC3339))
		}
		ec.EncodeCaller = zapcore.ShortCallerEncoder
		ec.EncodeName = zapcore.FullNameEncoder
		en := zapcore.NewJSONEncoder(ec)

		maxLevel := FatalLevel
		if i < len(configs)-1 {
			maxLevel = configs[i+1].Level
		}
		lef := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.Level(config.Level) && lvl < zapcore.Level(maxLevel)
		})

		cores[i] = zapcore.NewCore(en, zapcore.AddSync(config.Logger), lef)
	}
	loggers := make([]logger, len(cores))
	for i := range cores {
		loggers[i] = logger{
			level:  configs[i].Level,
			logger: zap.New(zapcore.NewTee(cores[i:]...), zap.AddCaller()),
		}
	}
	lg := Logger(loggers)
	return &lg
}

// GetLogger logger by level
func (l *Logger) GetLogger(level Level) *zap.SugaredLogger {
	if lg := l.Get(level); lg != nil {
		return lg.Sugar()
	}
	return nil
}

// Get logger by level
func (l *Logger) Get(level Level) *zap.Logger {
	loggers := []logger(*l)
	for i := len(loggers) - 1; i >= 0; i-- {
		lg := loggers[i]
		if level >= lg.level {
			return lg.logger
		}
	}
	return nil
}
