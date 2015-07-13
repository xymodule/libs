package statsdprof

import (
	"fmt"
	log "github.com/gonet2/libs/nsq-logger"
	"github.com/peterbourgon/g2s"
	"os"
	"path"
	"runtime"
	"time"
)

const (
	ENV_STATSD          = "STATSD_HOST"
	DEFAULT_STATSD_HOST = "127.0.0.1:8125"
	COLLECT_DELAY       = 1 * time.Minute
	SERVICE             = "[STATSD_PPROF]"
)

var (
	_statter g2s.Statter
)

func init() {
	addr := DEFAULT_STATSD_HOST
	if env := os.Getenv(ENV_STATSD); env != "" {
		addr = env
	}

	s, err := g2s.Dial("udp", addr)
	if err != nil {
		println(err)
		os.Exit(-1)
	}
	_statter = s

	go pprof_task()
}

// profiling task
func pprof_task() {
	for {
		<-time.After(COLLECT_DELAY)
		collect()
	}
}

// collect & publish to statsd
func collect() {
	tag := ""
	if hostname, err := os.Hostname(); err == nil {
		tag += hostname
	} else {
		log.Warning(SERVICE, err)
	}
	// executable name
	if exe_name, err := os.Readlink("/proc/self/exe"); err == nil {
		tag += "." + path.Base(exe_name)
	} else {
		tag += fmt.Sprintf(".%v", os.Getpid())
		log.Warning(SERVICE, err)
	}

	memstats := &runtime.MemStats{}
	runtime.ReadMemStats(memstats)

	_statter.Counter(1.0, tag+".NumGoroutine", int(runtime.NumGoroutine()))
	_statter.Counter(1.0, tag+".NumCgoCall", int(runtime.NumCgoCall()))
	if memstats.NumGC > 0 {
		_statter.Counter(1.0, tag+".NumGC", int(memstats.NumGC))
		_statter.Timing(1.0, tag+".PauseTotal", time.Duration(memstats.PauseTotalNs))
		_statter.Timing(1.0, tag+".LastPause", time.Duration(memstats.PauseNs[(memstats.NumGC+255)%256]))
		_statter.Counter(1.0, tag+".Alloc", int(memstats.Alloc))
		_statter.Counter(1.0, tag+".TotalAlloc", int(memstats.TotalAlloc))
		_statter.Counter(1.0, tag+".Sys", int(memstats.Sys))
		_statter.Counter(1.0, tag+".Lookups", int(memstats.Lookups))
		_statter.Counter(1.0, tag+".Mallocs", int(memstats.Mallocs))
		_statter.Counter(1.0, tag+".Frees", int(memstats.Frees))
		_statter.Counter(1.0, tag+".HeapAlloc", int(memstats.HeapAlloc))
		_statter.Counter(1.0, tag+".HeapSys", int(memstats.HeapSys))
		_statter.Counter(1.0, tag+".HeapIdle", int(memstats.HeapIdle))
		_statter.Counter(1.0, tag+".HeapInuse", int(memstats.HeapInuse))
		_statter.Counter(1.0, tag+".HeapReleased", int(memstats.HeapReleased))
		_statter.Counter(1.0, tag+".HeapObjects", int(memstats.HeapObjects))
		_statter.Counter(1.0, tag+".StackInuse", int(memstats.StackInuse))
		_statter.Counter(1.0, tag+".StackSys", int(memstats.StackSys))
		_statter.Counter(1.0, tag+".MSpanInuse", int(memstats.MSpanInuse))
		_statter.Counter(1.0, tag+".MSpanSys", int(memstats.MSpanSys))
		_statter.Counter(1.0, tag+".MCacheInuse", int(memstats.MCacheInuse))
		_statter.Counter(1.0, tag+".MCacheSys", int(memstats.MCacheSys))
		_statter.Counter(1.0, tag+".BuckHashSys", int(memstats.BuckHashSys))
		_statter.Counter(1.0, tag+".GCSys", int(memstats.GCSys))
		_statter.Counter(1.0, tag+".OtherSys", int(memstats.OtherSys))
	}
}
