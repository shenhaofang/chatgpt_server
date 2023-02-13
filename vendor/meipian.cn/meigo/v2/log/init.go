package log

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"meipian.cn/meigo/v2/config"
)

var Logger = logrus.New()

func init() {
	cfg := config.LoadLogConfig()
	Logger = NewLogger(cfg)
}

func NewLogger(cfg *config.LogConfig) *logrus.Logger {
	var l = logrus.New()
	setOut(cfg, l)

	switch cfg.Format {
	case "json":
		l.Formatter = new(logrus.JSONFormatter)
	default:
		l.Formatter = new(logrus.TextFormatter)
	}

	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		Err("Log Level: ", err.Error())
	} else {
		l.Level = level
	}
	return l
}

// setOut 设置日志出口；
// 如果未设置，默认在控制台输出；
func setOut(cfg *config.LogConfig, logger *logrus.Logger) {

	logFile := cfg.Path
	if logFile == "" {
		cfg.Level = "debug" // 无日志文件时，log level为debug模式，便于单元测试
		return
	}

	if cfg.Daily {
		date := time.Now().Format("20060102")
		logFile = cfg.Path + date

		// 每五分钟检查一次
		go func() {
			for {
				time.Sleep(5 * time.Minute)
				today := time.Now().Format("20060102")

				if date != today {
					date = today
					logFile = cfg.Path + date

					file, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
					if err != nil {
						logger.Error("Log: Failed to log to file, ", logFile, err.Error())
						break
					}
					logger.Out = file
				}
			}
		}()
	}

	file, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err == nil {
		logger.Out = file
	} else {
		Err("Log: Failed to log to file, ", logFile, err.Error())
	}
}
