package batcher

import (
	"time"
)

type Butcher interface {
	AddData(data interface{})
	Run() chan []interface{}
}
type Batching struct {
	dataChan chan interface{}
	isFullChan chan interface{}
	maxTime time.Duration
	maxItems int
}

func NewBatching( maxTime time.Duration, maxItems int) Butcher {
	return &Batching{dataChan: make(chan interface{}, maxItems), isFullChan: make(chan interface{}), maxTime: maxTime, maxItems: maxItems}
}

func (b *Batching) AddData(data interface{}) {
	// Add data to channel if not reach maximum length
	if len(b.dataChan) < b.maxItems {
		b.dataChan <- data
	} else {
		// Signal full item to process
		<- b.isFullChan
		// Add data to channel
		b.dataChan <- data
	}
}

func (b *Batching) Run() chan []interface{} {
	//Channel for delivery batch
	batches := make(chan []interface{})

	go func() {
		defer close(b.dataChan)
		defer close(b.isFullChan)
		var batch []interface{}
		// Timer for create batch if data not reach max items
		timer := time.NewTimer(b.maxTime)
		for {
			select {
			//Time condition reach making batch
			case <- timer.C:
				currentLenChan := len(b.dataChan)
				for i := 0; i < currentLenChan; i++ {
					data := <- b.dataChan
					batch = append(batch, data)
				}
				timer.Reset(b.maxTime)
				if len(batch) > 0 {
					batches <- batch
					batch = []interface{}{}
				}
			//Data full capacity making batch
			case b.isFullChan <- struct {}{}:
				if len(b.dataChan) == cap(b.dataChan) {
					currentLenChan := len(b.dataChan)
					//Append data to batch array
					for i := 0; i < currentLenChan; i++ {
						data := <- b.dataChan
						batch = append(batch, data)
					}
					//Reset timer to re countdown
					timer.Reset(b.maxTime)
					//Send data to be further process
					batches <- batch
					batch = []interface{}{}
				}
			}
		}
	}()
	return batches
}