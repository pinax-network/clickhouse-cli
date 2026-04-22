package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ZapLogger *zap.Logger
var SugaredLogger *zap.SugaredLogger

// Encoding selects the log output format.
type Encoding int

const (
	EncodingJSON Encoding = iota
	EncodingConsole
)

// Options configures the global logger.
type Options struct {
	Debug    bool
	Encoding Encoding
	File     *os.File // when set, output is written to this file instead of stdout/stderr
}

// Initialize configures the package-level ZapLogger and SugaredLogger. It must be
// called once at program startup before any logging call.
func Initialize(opts Options) error {
	level := zapcore.InfoLevel
	if opts.Debug {
		level = zapcore.DebugLevel
	}

	encoder := buildEncoder(opts.Encoding, level)

	var core zapcore.Core
	if opts.File != nil {
		core = zapcore.NewCore(encoder, zapcore.Lock(zapcore.AddSync(opts.File)), level)
	} else {
		highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		})
		lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl < zapcore.ErrorLevel && lvl >= level
		})
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, zapcore.Lock(os.Stderr), highPriority),
			zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), lowPriority),
		)
	}

	ZapLogger = zap.New(core)
	SugaredLogger = ZapLogger.Sugar()
	return nil
}

func buildEncoder(encoding Encoding, level zapcore.Level) zapcore.Encoder {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder

	if encoding == EncodingConsole {
		if level <= zapcore.DebugLevel {
			return zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		}
		return zapcore.NewConsoleEncoder(cfg)
	}
	return zapcore.NewJSONEncoder(cfg)
}

func Log(logLevel zapcore.Level, message string, additionalFields ...zap.Field) {
	ZapLogger.Log(logLevel, message, additionalFields...)
}

func Debug(message string, additionalFields ...zap.Field) {
	Log(zapcore.DebugLevel, message, additionalFields...)
}

func Debugf(template string, args ...any) {
	SugaredLogger.Debugf(template, args...)
}

func Info(message string, additionalFields ...zap.Field) {
	Log(zapcore.InfoLevel, message, additionalFields...)
}

func Infof(template string, args ...any) {
	SugaredLogger.Infof(template, args...)
}

func Warn(message string, additionalFields ...zap.Field) {
	Log(zapcore.WarnLevel, message, additionalFields...)
}

func Warnf(template string, args ...any) {
	SugaredLogger.Warnf(template, args...)
}

func Error(message string, additionalFields ...zap.Field) {
	Log(zapcore.ErrorLevel, message, additionalFields...)
}

func Errorf(template string, args ...any) {
	SugaredLogger.Errorf(template, args...)
}

func Panic(message string, additionalFields ...zap.Field) {
	Log(zapcore.PanicLevel, message, additionalFields...)
}

func Panicf(template string, args ...any) {
	SugaredLogger.Panicf(template, args...)
}

func Fatal(message string, additionalFields ...zap.Field) {
	Log(zapcore.FatalLevel, message, additionalFields...)
}

func Fatalf(template string, args ...any) {
	SugaredLogger.Fatalf(template, args...)
}
