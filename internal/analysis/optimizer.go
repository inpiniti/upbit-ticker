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
			result := RunSimulation(ticks, config)
			result.IntervalDuration = d

			resultsChan <- result
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

// RunSimulation 단일 시뮬레이션 (사이클 통계 포함)
func RunSimulation(ticks []types.RawTick, config types.BotConfiguration) types.OptimizationResult {
	// 초기 상태
	state := types.MarketState{
		IntervalBuffer:    []types.RawTick{},
		IntervalStartTime: 0,
		PrevAverage:       nil,
		PrevSlope:         nil,
		IsHolding:         false,
	}

	totalProfit := 0.0
	entryPrice := 0.0

	// 사이클 통계
	cycleCount := 0
	winCount := 0
	lossCount := 0
	totalWin := 0.0
	totalLoss := 0.0

	for _, tick := range ticks {
		res := ProcessTick(state, tick, config)

		if res.TradeSignal == "BUY" {
			entryPrice = ApplyCost("BUY", tick.Price, config)
		} else if res.TradeSignal == "SELL" {
			exitPrice := ApplyCost("SELL", tick.Price, config)
			cycleProfit := exitPrice - entryPrice
			totalProfit += cycleProfit
			cycleCount++

			// 수익/손실 분류
			if cycleProfit > 0 {
				winCount++
				totalWin += cycleProfit
			} else {
				lossCount++
				totalLoss += cycleProfit // 음수값
			}
		}

		state = res.NewState
	}

	// 승률 계산
	winRate := 0.0
	if cycleCount > 0 {
		winRate = float64(winCount) / float64(cycleCount)
	}

	// 평균 수익/손실 계산
	avgWin := 0.0
	if winCount > 0 {
		avgWin = totalWin / float64(winCount)
	}
	avgLoss := 0.0
	if lossCount > 0 {
		avgLoss = totalLoss / float64(lossCount)
	}

	return types.OptimizationResult{
		Profit:     totalProfit,
		CycleCount: cycleCount,
		WinCount:   winCount,
		LossCount:  lossCount,
		WinRate:    winRate,
		AvgWin:     avgWin,
		AvgLoss:    avgLoss,
	}
}
