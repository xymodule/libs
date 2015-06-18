package nsqlogger

import (
	"testing"
)

func TestLogger(t *testing.T) {
	SetPrefix("[TEST]")
	Finest("finest")
	Fine("fine")
	Debug("debug")
	Trace("trace")
	Info("info")
	Warning("warning")
	Error("error")
	Critical("critical")
}

func BenchmarkLogger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Debug("benchmark")
	}
}
