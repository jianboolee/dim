package logger

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

func Init() {
	logDir := os.Getenv("LOG_DIR")
	level := os.Getenv("LOG_LEVEL")
	if logDir == "" {
		logDir = "logs"
	}
	if level == "" {
		level = "debug"
	}

	// 创建 logs 目录
	os.MkdirAll(logDir, os.ModePerm)

	// 按日期生成文件名
	date := time.Now().Format("2006-01-02")
	logFile := fmt.Sprintf("%s/%s.log", logDir, date)

	// 设置 lumberjack 日志轮转
	writerSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100, // MB
		MaxBackups: 0,   // 0 表示不限制备份文件数
		MaxAge:     30,  // 保留 30 天
		Compress:   true,
	})

	// 日志等级
	logLevel := zapcore.InfoLevel
	if err := logLevel.Set(level); err != nil {
		logLevel = zapcore.InfoLevel
	}

	// 编码配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // 终端彩色
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewTee(
		// 输出到文件，json 格式
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), writerSyncer, logLevel),
		// 输出到终端，彩色
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(os.Stdout), logLevel),
	)

	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
}

// 简单封装
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

func Sync() {
	_ = Logger.Sync()
}
