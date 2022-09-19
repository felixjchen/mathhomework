package logging

import (
	"encoding/json"
	"time"

	"go.uber.org/zap"
)

func GetSugar() *zap.SugaredLogger {
	currentTime := time.Now()
	dateString := currentTime.Format("01-02-2006-15:04:05")
	rawJSON := []byte(`{
	  "level": "info",
	  "encoding": "json",
	  "outputPaths": ["stdout", "./logging/logs/log-` + dateString + `.log"],
	  "errorOutputPaths": ["stderr"],
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`)
	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}

	logger := zap.Must(cfg.Build())
	defer logger.Sync()

	sugar := logger.Sugar()
	return sugar
}
