package distmutex

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
	//"github.com/xymodule/libs/distmutex"
)

var endpoints []string = []string{"http://172.16.10.222:2379", "http://172.16.10.223:2379", "http://172.16.10.224:2379"}

func init() {
	Init("/backends", endpoints)
}

func TestDistMutex(t *testing.T) {
	DistMutexLockDo("100", func() {
		log.Info("lockDo OK")
	})
}

func BenchmarkDistMutex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DistMutexLockDo(fmt.Sprint(i), func() {
			//log.Infof("lockDo %v", i)
		})
	}
}
