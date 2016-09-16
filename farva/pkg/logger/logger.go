package logger

import (
	"github.com/Sirupsen/logrus"
)

var Log = logrus.New()

func init() {
	Log.Formatter = new(logrus.TextFormatter)
	Log.Level = logrus.DebugLevel
}
