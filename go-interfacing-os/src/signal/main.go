package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	fmt.Printf("Waiting for signal, pid: %d\n", os.Getpid())

	sig := <-sigs

	fmt.Println("Program ", sig)
	fmt.Println("Program finished")
}
