package batcher

import (
	"time"
)

type Batcher interface {
	Add(data interface{})
	Run()
	GetBatches() <-chan []interface{}
	Stop()
}
type batcher struct {
	dataChan        chan interface{}
	isFullChan      chan struct{}
	batches         chan []interface{}
	isTerminateChan chan interface{}
	maxTime         time.Duration
	maxItems        int
	keepGoing       bool
}

func New(maxTime time.Duration, maxItems int) Batcher {
	return &batcher{
		dataChan:        make(chan interface{}, maxItems),
		isFullChan:      make(chan struct{}, 1),
		batches:         make(chan []interface{}, 1),
		isTerminateChan: make(chan interface{}, 1),
		maxTime:         maxTime,
		maxItems:        maxItems,
		keepGoing:       true,
	}
}

func (b *batcher) Add(data interface{}) {
	//add data to data channel
	//Send signal when data reach full capacity for batching immediately
	if len(b.dataChan) < b.maxItems {
		b.dataChan <- data
		if len(b.dataChan) == b.maxItems {
			<-b.isFullChan
		}

	}else if len(b.dataChan) == b.maxItems {
		<-b.isFullChan
		b.dataChan <- data
	}
}

func (b *batcher) Stop() {
	//Send signal for close all opened channels
	b.isTerminateChan <- struct{}{}
}


func (b *batcher) GetBatches() <-chan []interface{} {
	//Channel for delivery batch
	return b.batches
}

func (b *batcher) Run() {
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

func (b *batcher) terminateChan() {
	b.keepGoing = false
	close(b.dataChan)
	close(b.isFullChan)
	close(b.isTerminateChan)
	close(b.batches)
}

func (b *batcher) processBatch(timer *time.Timer)  {
	var batch []interface{}
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