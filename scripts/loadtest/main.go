package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	users := flag.Int("users", 200, "concurrent users")
	activityID := flag.String("activity", "", "activity ID")
	host := flag.String("host", "http://localhost:8080", "API host")
	flag.Parse()

	if *activityID == "" {
		fmt.Println("usage: go run ./scripts/loadtest -activity <id>")
		return
	}

	var (
		success atomic.Int64
		fail    atomic.Int64
		wg      sync.WaitGroup
	)

	start := time.Now()

	for i := 0; i < *users; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			body, _ := json.Marshal(map[string]string{
				"user_id":     fmt.Sprintf("user_%d", userID),
				"activity_id": *activityID,
			})
			resp, err := http.Post(
				*host+"/api/seckill",
				"application/json",
				bytes.NewReader(body),
			)
			if err != nil {
				fail.Add(1)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusAccepted {
				success.Add(1)
			} else {
				fail.Add(1)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("\nLoad test complete in %v\n", elapsed)
	fmt.Printf("Success (202): %d\n", success.Load())
	fmt.Printf("Failed  (4xx): %d\n", fail.Load())
	fmt.Printf("QPS: %.0f\n", float64(*users)/elapsed.Seconds())
}
