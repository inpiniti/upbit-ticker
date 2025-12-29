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

// RunSimulation 단일 시뮬레이션 (일반 + 마틴게일 전략 비교)
func RunSimulation(ticks []types.RawTick, config types.BotConfiguration) types.OptimizationResult {
	// 초기 상태
	state := types.MarketState{
		IntervalBuffer:    []types.RawTick{},
		IntervalStartTime: 0,
		PrevAverage:       nil,
		PrevSlope:         nil,
		IsHolding:         false,
	}

	// 일반 전략 변수
	totalProfit := 0.0
	entryPrice := 0.0

	// 사이클 통계
	cycleCount := 0
	winCount := 0
	lossCount := 0
	totalWin := 0.0
	totalLoss := 0.0

	// 마틴게일 전략 변수
	martingaleProfit := 0.0
	martingaleMultiplier := 1
	martingaleMaxMultiplier := 1

	for _, tick := range ticks {
		res := ProcessTick(state, tick, config)

		if res.TradeSignal == "BUY" {
			entryPrice = ApplyCost("BUY", tick.Price, config)
		} else if res.TradeSignal == "SELL" {
			exitPrice := ApplyCost("SELL", tick.Price, config)
			cycleProfit := exitPrice - entryPrice

			// 일반 전략: 1배 고정
			totalProfit += cycleProfit
			cycleCount++

			// 마틴게일 전략: 배율 적용
			martingaleProfit += cycleProfit * float64(martingaleMultiplier)

			// 수익/손실 분류 및 마틴게일 배율 조정
			if cycleProfit > 0 {
				winCount++
				totalWin += cycleProfit
				// 마틴게일: 수익 시 배율 초기화
				martingaleMultiplier = 1
			} else {
				lossCount++
				totalLoss += cycleProfit // 음수값
				// 마틴게일: 손실 시 배율 2배
				martingaleMultiplier *= 2
				// 최대 배율 기록 (리스크 지표)
				if martingaleMultiplier > martingaleMaxMultiplier {
					martingaleMaxMultiplier = martingaleMultiplier
				}
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
		// 일반 전략
		Profit:    totalProfit,
		WinCount:  winCount,
		LossCount: lossCount,
		WinRate:   winRate,
		AvgWin:    avgWin,
		AvgLoss:   avgLoss,
		// 공통
		CycleCount: cycleCount,
		// 마틴게일 전략
		MartingaleProfit:        martingaleProfit,
		MartingaleMaxMultiplier: martingaleMaxMultiplier,
	}
}
