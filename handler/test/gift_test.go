package test

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const url = "http://localhost:5678/lucky"
const P = 100 // 100 simulated users hammering the draw button

func TestLottery(t *testing.T) {
	hitMap := make(map[string]int, 10) // how many times each prize was drawn
	giftCh := make(chan string, 10000) // ids of the prizes that were drawn
	counterCh := make(chan struct{})   // signals that the counting goroutine is done

	// count the draws off the hot path
	go func() {
		for giftId := range giftCh {
			hitMap[giftId]++
		}
		counterCh <- struct{}{} // counting is finished
	}()

	wg := sync.WaitGroup{}
	wg.Add(P)
	begin := time.Now()
	var totalCall int64    // total number of requests
	var totalUseTime int64 // total time those requests took
	for i := 0; i < P; i++ {
		go func() {
			defer wg.Done()
			for {
				t1 := time.Now()
				resp, err := http.Get(url)
				atomic.AddInt64(&totalUseTime, time.Since(t1).Milliseconds())
				atomic.AddInt64(&totalCall, 1)
				if err != nil {
					fmt.Println(err)
					break
				}
				bs, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
					break
				}
				resp.Body.Close()
				giftId := string(bs)
				if len(giftId) > 0 {
					if giftId == "0" { // a prize id of 0 means everything has been handed out
						break
					}
					giftCh <- giftId // record the prize that was drawn
				} else {
					fmt.Println("empty giftId")
				}
			}
		}()
	}
	wg.Wait()
	close(giftCh)
	<-counterCh // wait until hitMap is complete

	totalTime := int64(time.Since(begin).Seconds())
	if totalTime > 0 && totalCall > 0 {
		qps := totalCall / totalTime
		avgTime := totalUseTime / totalCall
		fmt.Printf("totalCall %d / totalTime %ds = QPS %d\n", totalCall, totalTime, qps)
		fmt.Printf("totalUseTime %dms / totalCall %d = avg time %dms\n", totalUseTime, totalCall, avgTime)
		//QPS 6600, avg time 26ms
		// Every successful request consumes exactly one unit (ReduceInventory decrements once per returned gift id), so it takes 6500 requests to drain the warehouse. Then each of the 100 workers needs one more request to see "0" and exit its loop:
		// 6500 + 100 = 6600 calls
		total := 0
		for giftId, count := range hitMap {
			fmt.Printf("%s\t%d\n", giftId, count)
			total += count
		}
		fmt.Printf("%d prizes in total\n", total)
	}
}

// go test -v ./handler/test -run="^TestLottery$" -count=1
