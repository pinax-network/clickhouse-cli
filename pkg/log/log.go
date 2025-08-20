package log

import (
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ZapLogger *zap.Logger
var SugaredLogger *zap.SugaredLogger

// InitializeGlobalLogger initializes a global json logger with sane defaults.
// If logDebug is set to true, we are going to log debug messages as well, otherwise the min level is going to be Info.
//
// Note: This is the preferred
//
// Example code:
//
//	// We only need to initialize the global logger once in our application
//	_ = log.InitializeGlobalLogger(false)
//	log.Info("successfully initialized the global logger!")
func InitializeGlobalLogger(logDebug bool) (err error) {

	var logger *zap.Logger

	if logDebug {
		logger, err = InitializeJsonLogger(zapcore.DebugLevel)
	} else {
		logger, err = InitializeJsonLogger(zapcore.InfoLevel)
	}

	if err != nil {
		return err
	}

	ZapLogger = logger
	SugaredLogger = logger.Sugar()

	return nil
}

// InitializeGlobalConsoleLogger works like InitializeGlobalLogger, but uses a console encoding format which is designed
// for human consumption. This is useful for local development. For Kubernetes environments, InitializeGlobalLogger
// should be used instead.
// Note that calling this method will overwrite any global logger set by InitializeGlobalLogger.
func InitializeGlobalConsoleLogger(logDebug bool) (err error) {

	var logger *zap.Logger

	if logDebug {
		logger, err = InitializeConsoleLogger(zapcore.DebugLevel)
	} else {
		logger, err = InitializeConsoleLogger(zapcore.InfoLevel)
	}

	if err != nil {
		return err
	}

	ZapLogger = logger
	SugaredLogger = logger.Sugar()

	return nil
}

// InitializeGlobalFileLogger works like InitializeGlobalLogger, but writes to a file instead of the console. Note that
// calling this method will overwrite any global logger set by InitializeGlobalLogger.
func InitializeGlobalFileLogger(logDebug bool, file *os.File) (err error) {

	var logger *zap.Logger

	if logDebug {
		logger, err = InitializeFileLogger(zapcore.DebugLevel, file)
	} else {
		logger, err = InitializeFileLogger(zapcore.InfoLevel, file)
	}

	if err != nil {
		return err
	}

	ZapLogger = logger
	SugaredLogger = logger.Sugar()

	return nil
}

// InitializeConsoleLogger initializes and returns a zap.Logger with sane defaults and the given min level. Note that
// this will not set the global logger provided by this package. To use it you need to store it in a variable and call
// the zap logging methods directly. Example:
//
//	logger, _ := InitializeConsoleLogger(zapcore.Info)
//	logger.Info("initialized console logger!")
func InitializeConsoleLogger(minLevel zapcore.Level) (logger *zap.Logger, err error) {

	var consoleEncoder zapcore.Encoder

	if minLevel <= zapcore.DebugLevel {
		consoleEncoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	} else {
		cfg := zap.NewProductionEncoderConfig()
		cfg.EncodeTime = zapcore.ISO8601TimeEncoder
		consoleEncoder = zapcore.NewConsoleEncoder(cfg)
	}

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel && lvl >= minLevel
	})

	consoleOut := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleOut, lowPriority),
	)

	logger = zap.New(core)

	return
}

// InitializeJsonLogger works like InitializeConsoleLogger but uses structured json output only. This should be the
// preferred way for logging in Kubernetes environments.
func InitializeJsonLogger(minLevel zapcore.Level) (logger *zap.Logger, err error) {

	var consoleEncoder zapcore.Encoder

	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder = zapcore.NewJSONEncoder(cfg)

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel && lvl >= minLevel
	})

	consoleOut := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleOut, lowPriority),
	)

	logger = zap.New(core)

	return
}

// InitializeFileLogger works like InitializeConsoleLogger, but returns a zap.Logger that writes to the given file.
func InitializeFileLogger(minLevel zapcore.Level, file *os.File) (logger *zap.Logger, err error) {

	var consoleEncoder zapcore.Encoder

	if minLevel <= zapcore.DebugLevel {
		consoleEncoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	} else {
		cfg := zap.NewProductionEncoderConfig()
		cfg.EncodeTime = zapcore.ISO8601TimeEncoder
		consoleEncoder = zapcore.NewConsoleEncoder(cfg)
	}

	core := zapcore.NewCore(consoleEncoder, file, minLevel)
	logger = zap.New(core)

	return
}

func Log(logLevel zapcore.Level, message string, additionalFields ...zap.Field) {

	if ZapLogger == nil {
		log.Println("zap logger isn't initialized yet!")
		log.Println(message)

		for _, f := range additionalFields {
			log.Println(fmt.Sprintf("'%s': %+v", f.Key, f))
		}
		if logLevel == zapcore.FatalLevel {
			os.Exit(1)
		}
		if logLevel == zapcore.PanicLevel {
			panic(message)
		}
		return
	}

	ZapLogger.Log(logLevel, message, additionalFields...)
}

func Debug(message string, additionalFields ...zap.Field) {
	Log(zapcore.DebugLevel, message, additionalFields...)
}

func Debugf(template string, args ...interface{}) {
	SugaredLogger.Debugf(template, args)
}

func Info(message string, additionalFields ...zap.Field) {
	Log(zapcore.InfoLevel, message, additionalFields...)
}

func Infof(template string, args ...interface{}) {
	SugaredLogger.Infof(template, args)
}

func Warn(message string, additionalFields ...zap.Field) {
	Log(zapcore.WarnLevel, message, additionalFields...)
}

func Warnf(template string, args ...interface{}) {
	SugaredLogger.Warnf(template, args)
}

func Error(message string, additionalFields ...zap.Field) {
	Log(zapcore.ErrorLevel, message, additionalFields...)
}

func Errorf(template string, args ...interface{}) {
	SugaredLogger.Errorf(template, args)
}

func Panic(message string, additionalFields ...zap.Field) {
	Log(zapcore.PanicLevel, message, additionalFields...)
}

func Panicf(template string, args ...interface{}) {
	SugaredLogger.Panicf(template, args)
}

func Fatal(message string, additionalFields ...zap.Field) {
	Log(zapcore.FatalLevel, message, additionalFields...)
}

func Fatalf(template string, args ...interface{}) {
	SugaredLogger.Fatalf(template, args)
}
