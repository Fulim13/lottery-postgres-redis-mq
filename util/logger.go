package util

import (
	"log/slog"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

func InitSlog(logFile string) {
	fout, err := rotatelogs.New(
		logFile+".%Y%m%d%H",                      // path and name of the log file; missing directories are created
		rotatelogs.WithLinkName(logFile),         // symlink pointing at the newest log file
		rotatelogs.WithRotationTime(1*time.Hour), // start a new log file every hour
		rotatelogs.WithMaxAge(7*24*time.Hour),    // keep the last 7 days; WithRotationCount keeps a number of files instead
	)
	if err != nil {
		panic(err)
	}

	handler := slog.NewTextHandler(
		fout, // write to the file
		&slog.HandlerOptions{
			AddSource: true,           // include the file name and line number
			Level:     slog.LevelInfo, // lowest level that gets written
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey { // when the key is "time"
					t := a.Value.Time()
					a.Value = slog.StringValue(t.Format("2006-01-02 15:04:05.000")) // rewrite the value
				}
				return a
			},
		},
	)
	logger := slog.New(handler)

	slog.SetDefault(logger)
}
