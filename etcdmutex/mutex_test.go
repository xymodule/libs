package etcdmutex

import (
	"fmt"
	"testing"
)

func TestMutex(t *testing.T) {
	m := Lock("player1")
	if m == nil {
		t.Fatal("cannot lock")
	}

	t.Log(m)

	if !m.Unlock() {
		t.Fatal("cannot unlock")
	}
}

func BenchmarkMutex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := Lock(fmt.Sprintf("test%v", i))
		if m == nil {
			b.Fatal("cannot lock")
		}

		b.Log(m)

		if !m.Unlock() {
			b.Fatal("cannot unlock")
		}
	}
}
