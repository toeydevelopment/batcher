```go
package main
func main() {
	b := batcher.New(time.Second*5, 3)
	//Start to receive data 
	b.Run()
	//Chan to receive batch 
	batches := b.GetBatches() 
	// Calculate execution time and batch data
	for batch := range batches {
		fmt.Println(batch)
	}
}
```
## benchmark batcher
```go
go test -bench=. -benchmem -benchtime=1s
```