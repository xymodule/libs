package nsqlogger

import (
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	SetPrefix("[TEST]")
	SetLogLevel(TRACE)
	Finest("finest")
	Fine("fine")
	Debug("debug")
	Trace("trace")
	Info("info")
	Warning("warning")
	Error("error")
	Critical("critical")
	<-time.After(time.Minute)
}

func BenchmarkLogger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Debug("benchmark")
	}
}
