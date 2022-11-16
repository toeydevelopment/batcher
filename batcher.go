package batcher

import (
	"time"
)

type Batcher[T any] interface {
	Add(data T)
	Run()
	GetBatches() <-chan []T
	Stop()
}
type batcher[T any] struct {
	dataChan        chan T
	isFullChan      chan struct{}
	batches         chan []T
	isTerminateChan chan struct{}
	maxTime         time.Duration
	maxItems        int
	keepGoing       bool
}

func New[T any](maxTime time.Duration, maxItems int) Batcher[T] {
	return &batcher[T]{
		dataChan:        make(chan T, maxItems),
		isFullChan:      make(chan struct{}, 1),
		batches:         make(chan []T, 1),
		isTerminateChan: make(chan struct{}, 1),
		maxTime:         maxTime,
		maxItems:        maxItems,
		keepGoing:       true,
	}
}

func (b *batcher[T]) Add(data T) {
	//add data to data channel
	//Send signal when data reach full capacity for batching immediately
	if len(b.dataChan) < b.maxItems {
		b.dataChan <- data
		if len(b.dataChan) == b.maxItems {
			<-b.isFullChan
		}

	} else if len(b.dataChan) == b.maxItems {
		<-b.isFullChan
		b.dataChan <- data
	}
}

func (b *batcher[T]) Stop() {
	//Send signal for close all opened channels
	b.isTerminateChan <- struct{}{}
}

func (b *batcher[T]) GetBatches() <-chan []T {
	//Channel for delivery batch
	return b.batches
}

func (b *batcher[T]) Run() {
	go func() {
		// Timer for create batch if data not reach max items
		timer := time.NewTimer(b.maxTime)
		for b.keepGoing {
			select {
			//Time condition reach making batch
			case <-timer.C:
				b.processBatch(timer)
			case <-b.isTerminateChan:
				b.processBatch(timer)
				timer.Stop()
				b.terminateChan()
			//Data full capacity making batch
			case b.isFullChan <- struct{}{}:
				b.processBatch(timer)
			}
		}
	}()
}

func (b *batcher[T]) terminateChan() {
	b.keepGoing = false
	close(b.dataChan)
	close(b.isFullChan)
	close(b.isTerminateChan)
	close(b.batches)
}

func (b *batcher[T]) processBatch(timer *time.Timer) {
	var batch []T

	currentLenChan := len(b.dataChan)
	for i := 0; i < currentLenChan; i++ {
		data := <-b.dataChan
		batch = append(batch, data)
	}

	if len(batch) > 0 {
		b.batches <- batch
	}
	timer.Reset(b.maxTime)
}
