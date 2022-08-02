package tool

import (
	"github.com/spf13/viper"
	"os"

	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

// UTCFormatter ...
type UTCFormatter struct {
	easy.Formatter
}

// Format ...
func (u UTCFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return u.Formatter.Format(e)
}

// GetLoggerByRequestID ...
func GetLoggerByRequestID(rid string) *logrus.Entry {
	level, ok := viper.Get("log-level").(logrus.Level)
	if !ok {
		panic("Get log-level error")
	}
	l := &logrus.Logger{
		Out: os.Stderr,
		// set logger level
		Level: level,
		Formatter: &UTCFormatter{
			easy.Formatter{
				TimestampFormat: "2006-01-02T15:04:05.999Z",
				LogFormat:       "%time%: %requestId% [%lvl%]  %msg%\n",
			},
		},
	}
	le2 := l.WithField("requestId", rid)
	return le2
}
