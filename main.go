package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World")
	})

	urls := []string{}
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "ALIVE_") {
			value := strings.Split(e, "=")
			urls = append(urls, value[1])
		}
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var wg sync.WaitGroup

	go func() {
		for range ticker.C {
			wg.Add(1)
			go fetchURLs(urls, &wg)
		}
	}()

	server := &http.Server{Addr: ":8001"}

	// Channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("ListenAndServe(): %v\n", err)
		}
	}()

	<-stop // Wait for a signal

	fmt.Println("Shutting down gracefully...")
	ticker.Stop()

	// Create a context with a timeout for the shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server Shutdown Failed:%+v", err)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	fmt.Println("Server gracefully stopped")
}

func fetchURLs(urls []string, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error fetching URL:", url, err)
			return
		}
		defer resp.Body.Close()
		fmt.Println("Fetched URL:", url, "Status Code:", resp.StatusCode)
	}
}
