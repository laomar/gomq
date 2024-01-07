package log

import (
	"github.com/gin-gonic/gin"
	. "github.com/laomar/gomq/config"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var log *zap.SugaredLogger

// Init log
func Init() {
	ws := make([]zapcore.WriteSyncer, 1)
	ws[0] = zapcore.AddSync(file())
	if Cfg.Env == "dev" {
		ws = append(ws, os.Stdout)
	}

	core := zapcore.NewCore(encoder(), zapcore.NewMultiWriteSyncer(ws...), level())
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	log = logger.Sugar()
}

func encoder() zapcore.Encoder {
	c := zap.NewProductionEncoderConfig()
	if Cfg.Env == "dev" {
		c = zap.NewDevelopmentEncoderConfig()
	}
	c.TimeKey = "time"
	c.EncodeTime = func(t time.Time, e zapcore.PrimitiveArrayEncoder) {
		e.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	c.EncodeCaller = func(c zapcore.EntryCaller, e zapcore.PrimitiveArrayEncoder) {
		e.AppendString(c.TrimmedPath())
	}

	e := zapcore.NewJSONEncoder(c)
	if strings.ToLower(Cfg.Log.Format) == "text" || Cfg.Env == "dev" {
		e = zapcore.NewConsoleEncoder(c)
	}
	return e
}

func file() *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   Cfg.DataDir + "/log/" + filepath.Base(os.Args[0]) + ".log",
		MaxAge:     Cfg.Log.MaxAge,
		MaxSize:    Cfg.Log.MaxSize,
		MaxBackups: Cfg.Log.MaxCount,
		LocalTime:  true,
		Compress:   false,
	}
}

// Get log level
func level() (l zapcore.Level) {
	switch strings.ToLower(Cfg.Log.Level) {
	case "debug":
		l = zapcore.DebugLevel
	case "warning":
		l = zapcore.WarnLevel
	case "error":
		l = zapcore.ErrorLevel
	case "panic":
		l = zapcore.PanicLevel
	case "fatal":
		l = zapcore.FatalLevel
	default:
		l = zapcore.InfoLevel
	}
	return l
}

// Log function

func Debug(args ...any) {
	log.Debug(args...)
}
func Debugf(tpl string, args ...any) {
	log.Debugf(tpl, args)
}
func Info(args ...any) {
	log.Info(args...)
}
func Infof(tpl string, args ...any) {
	log.Infof(tpl, args)
}
func Warn(args ...any) {
	log.Warn(args...)
}
func Warnf(tpl string, args ...any) {
	log.Warnf(tpl, args)
}
func Error(args ...any) {
	log.Error(args...)
}
func Errorf(tpl string, args ...any) {
	log.Errorf(tpl, args)
}
func Fatal(args ...any) {
	log.Fatal(args...)
}
func Fatalf(tpl string, args ...any) {
	log.Fatalf(tpl, args)
}

// Logger Gin log middleware
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}
