```go
package main
func main() {
	b := batching.NewBatching(time.Second*5, 3)
	//Chan to receive batch
	batches := b.Run()
	// Calculate execution time and batch data
	for batch := range batches {
		fmt.Println(batch)
	}
}
```