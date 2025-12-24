package analysis

import (
	"sort"
	"sync"
	"time"
	"upbit-ticker/types"
)

// FindSweetSpot 주어진 틱 데이터로 최적의 구간을 병렬로 탐색
func FindSweetSpot(ticks []types.RawTick, baseConfig types.BotConfiguration) []types.OptimizationResult {
	intervals := generateIntervals()
	resultsChan := make(chan types.OptimizationResult, len(intervals))
	var wg sync.WaitGroup

	// Concurrency Control (Worker Pool Pattern)
	maxConcurrency := 10 // Worker 개수 조정
	sem := make(chan struct{}, maxConcurrency)

	for _, duration := range intervals {
		wg.Add(1)
		sem <- struct{}{} // Acquire token

		go func(d time.Duration) {
			defer wg.Done()
			defer func() { <-sem }() // Release token

			// Run Simulation
			config := baseConfig
			config.IntervalDuration = d
			profit, count := RunSimulation(ticks, config)

			resultsChan <- types.OptimizationResult{
				IntervalDuration: d,
				Profit:           profit,
				TradeCount:       count,
			}
		}(duration)
	}

	wg.Wait()
	close(resultsChan)

	// Collect Results
	var results []types.OptimizationResult
	for res := range resultsChan {
		results = append(results, res)
	}

	// Sort by Profit Descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Profit > results[j].Profit
	})

	// Return Top 5 or all? Prompt says Top 5 list.
	if len(results) > 5 {
		return results[:5]
	}
	return results
}

// generateIntervals 1초~24시간 구간 생성
func generateIntervals() []time.Duration {
	var list []time.Duration
	// 1초 ~ 59초
	for s := 1; s < 60; s++ {
		list = append(list, time.Duration(s)*time.Second)
	}
	// 1분 ~ 24시간 (분 단위)
	for m := 1; m <= 60*24; m++ {
		list = append(list, time.Duration(m)*time.Minute)
	}
	return list
}

// RunSimulation 단일 시뮬레이션 (In-Memory)
func RunSimulation(ticks []types.RawTick, config types.BotConfiguration) (float64, int) {
	// 초기 상태
	state := types.MarketState{
		IntervalBuffer:    []types.RawTick{},
		IntervalStartTime: 0,
		PrevAverage:       nil,
		PrevSlope:         nil,
		IsHolding:         false,
	}

	totalProfit := 0.0
	tradeCount := 0
	entryPrice := 0.0

	for _, tick := range ticks {
		res := ProcessTick(state, tick, config)

		if res.TradeSignal == "BUY" {
			entryPrice = ApplyCost("BUY", tick.Price, config)
		} else if res.TradeSignal == "SELL" {
			exitPrice := ApplyCost("SELL", tick.Price, config)
			totalProfit += (exitPrice - entryPrice)
			tradeCount++
		}

		state = res.NewState
	}

	return totalProfit, tradeCount
}
