package batcher_test

import (
	"github.com/new-fg/batcher"
	"sync"
	"testing"
	"time"
)

func BenchmarkStructChan(b *testing.B) {
	wg := sync.WaitGroup{}
	_batcher := batcher.New(time.Second*15, 5)
	//Chan to receive batch
	_batcher.Run()
	batches := _batcher.GetBatches()

	wg.Add(1)
	go func() {
		defer wg.Done()
		//start := time.Now()
		for  range batches {
			//fmt.Println(time.Now().Sub(start), batch)
		}
	}()

	b.SetBytes(1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_batcher.Add(i)
	}
	_batcher.Stop()
	wg.Wait()
}