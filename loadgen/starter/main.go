package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"go.temporal.io/sdk/client"

	loadgen "temporal-server-playground/loadgen"
)

func main() {
	workflows  := flag.Int("workflows", 10, "number of workflows to start")
	activities := flag.Int("activities", 3, "number of concurrent activities per workflow")
	children   := flag.Int("children", 3, "number of concurrent child workflows per workflow")
	intervalMs := flag.Int("interval", 0, "ms between each workflow start (0 = start all concurrently)")
	flag.Parse()

	c, err := client.Dial(client.Options{
		MetricsHandler: loadgen.NewMetricsHandler("temporal-starter"),
	})
	if err != nil {
		log.Fatalln("unable to create Temporal client:", err)
	}
	defer c.Close()

	params := loadgen.Params{
		NumActivities:     *activities,
		NumChildWorkflows: *children,
	}

	start := func(i int) {
		_, err := c.ExecuteWorkflow(
			context.Background(),
			client.StartWorkflowOptions{
				ID:        fmt.Sprintf("load-gen-%d-%d", time.Now().UnixMilli(), i),
				TaskQueue: loadgen.TaskQueue,
			},
			loadgen.ParentWorkflow,
			params,
		)
		if err != nil {
			log.Printf("failed to start workflow %d: %v", i, err)
		}
	}

	if *intervalMs > 0 {
		// Start one workflow every intervalMs milliseconds
		ticker := time.NewTicker(time.Duration(*intervalMs) * time.Millisecond)
		defer ticker.Stop()
		for i := 0; i < *workflows; i++ {
			<-ticker.C
			go start(i)
		}
	} else {
		// Start all concurrently
		var wg sync.WaitGroup
		for i := 0; i < *workflows; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				start(i)
			}(i)
		}
		wg.Wait()
	}

	fmt.Printf("started %d workflows (%d activities + %d child workflows each)\n",
		*workflows, *activities, *children)
}
