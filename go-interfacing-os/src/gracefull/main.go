package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const workerPoolSize = 4

func main() {

	ingestChan := make(chan int, 1)
	consumer := Consumer{
		ingestChan: ingestChan,
		jobsChan:   make(chan int, workerPoolSize),
	}

	// Simulate some input
	producer := Producer{ingestChan: ingestChan}
	go producer.start()

	// Set up cancellation context and waitgroup
	ctx, cancelFunc := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	// Start control loop with cancel context
	go consumer.startConsumer(ctx)

	// Start workers
	wg.Add(workerPoolSize)
	for i := 0; i < workerPoolSize; i++ {
		go consumer.worker(wg, i)
	}

	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	<-termChan // Blocks here until interrupted

	fmt.Println("\n==== Shutdown received ====")
	cancelFunc() // Signal cancel to context
	wg.Wait()    // Wait for workers

	fmt.Println("All workers done, shutting down!")
}

type Consumer struct {
	ingestChan chan int
	jobsChan   chan int
}

func (c Consumer) worker(wg *sync.WaitGroup, index int) {
	defer wg.Done()

	fmt.Printf("Worker %d start\n", index)
	for jobId := range c.jobsChan {
		// simulate work
		fmt.Printf("Worker %d started job %d\n", index, jobId)
		time.Sleep(time.Millisecond * time.Duration(1000+rand.Intn(2000)))
		fmt.Printf("Worker %d finished job %d\n", index, jobId)
	}
	fmt.Printf("Worker %d stop\n", index)
}

func (c Consumer) startConsumer(ctx context.Context) {
	for {
		select {
		case job := <-c.ingestChan:
			c.jobsChan <- job
		case <-ctx.Done():
			fmt.Println("Cancel received, closing jobsChan!")
			close(c.jobsChan)
			fmt.Println("Closed jobsChan")
			return
		}
	}
}

type Producer struct {
	ingestChan chan int
}

func (p Producer) start() {
	jobId := 0
	for {
		p.ingestChan <- jobId
		jobId++
		time.Sleep(time.Millisecond * 300)
	}
}
