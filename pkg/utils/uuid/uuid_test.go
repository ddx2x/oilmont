package uuid

import (
	"fmt"
	"sync"
	"testing"
)

func TestSUID(t *testing.T) {
	u := NewSUID()

	fmt.Printf("[%s]\r\n[%s]\r\n", u.String(), u.String())
}

func BenchmarkUUID(b *testing.B) {
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		uuid := NewSUID()
		_ = uuid
		wg.Done()
	}
	wg.Wait()
}

var mapx = make(map[string]struct{})

func BenchmarkTestSUUIDX(b *testing.B) {
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		uuid := NewSUID().String()
		if _, exist := mapx[uuid]; exist {
			b.Fatal("this uuid is not unique")
		}
		mapx[uuid] = struct{}{}
		wg.Done()
	}
	wg.Wait()
}
