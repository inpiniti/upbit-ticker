package analysis

import (
	"time"
	"upbit-ticker/types"
)

// ProcessTick 상태와 틱을 받아 새로운 상태와 이벤트를 반환하는 순수 함수
func ProcessTick(state types.MarketState, tick types.RawTick, config types.BotConfiguration) types.ProcessResult {
	// 1. 버퍼 업데이트 (불변성 유지를 위해 새로운 슬라이스 생성)
	nextBuffer := make([]types.RawTick, len(state.IntervalBuffer)+1)
	copy(nextBuffer, state.IntervalBuffer)
	nextBuffer[len(state.IntervalBuffer)] = tick

	// 시작 시간 설정
	startTime := state.IntervalStartTime
	if len(state.IntervalBuffer) == 0 {
		startTime = tick.Timestamp
	}

	// 2. 구간 종료 확인 (현재 Tick 시간 - 시작 시간 >= 설정 시간)
	// 시간 단위는 ms를 가정 (JS 코드 기준). Go time.Duration은 ns 단위이므로 변환 주의.
	// tick.Timestamp가 Unix Milliseconds라고 가정.
	elapsed := time.Duration(tick.Timestamp-startTime) * time.Millisecond
	isIntervalFinished := elapsed >= config.IntervalDuration

	if !isIntervalFinished {
		return types.ProcessResult{
			NewState: types.MarketState{
				IntervalBuffer:    nextBuffer,
				IntervalStartTime: startTime,
				PrevAverage:       state.PrevAverage,
				PrevSlope:         state.PrevSlope,
				IsHolding:         state.IsHolding,
			},
			TradeSignal:    "",
			IntervalClosed: false,
		}
	}

	// --- 구간 완성 시 로직 ---
	currentAverage := CalculateAverage(nextBuffer)
	currentSlope := CalculateSlope(currentAverage, state.PrevAverage)
	signal := EvaluateSignal(state.PrevSlope, currentSlope)

	tradeSignal := ""
	nextIsHolding := state.IsHolding

	if signal == "BUY" && !state.IsHolding {
		tradeSignal = "BUY"
		nextIsHolding = true
	} else if signal == "SELL" && state.IsHolding {
		tradeSignal = "SELL"
		nextIsHolding = false
	}

	return types.ProcessResult{
		NewState: types.MarketState{
			IntervalBuffer:    []types.RawTick{}, // 버퍼 리셋
			IntervalStartTime: 0,                 // 초기화 (다음 틱에서 설정됨)
			PrevAverage:       &currentAverage,
			PrevSlope:         currentSlope,
			IsHolding:         nextIsHolding,
		},
		TradeSignal:    tradeSignal,
		IntervalClosed: true,
		CurrentAverage: currentAverage,
		CurrentSlope:   currentSlope,
	}
}

// CalculateAverage 평균 계산
func CalculateAverage(ticks []types.RawTick) float64 {
	if len(ticks) == 0 {
		return 0
	}
	sum := 0.0
	for _, t := range ticks {
		sum += t.Price
	}
	return sum / float64(len(ticks))
}

// CalculateSlope 기울기 계산 (현재 - 이전)
func CalculateSlope(curr float64, prev *float64) *float64 {
	if prev == nil {
		return nil
	}
	slope := curr - *prev
	return &slope
}

// EvaluateSignal 매매 신호 평가
func EvaluateSignal(prevSlope *float64, currSlope *float64) string {
	if prevSlope == nil || currSlope == nil {
		return "HOLD"
	}

	// V자 반등: 이전 기울기 음수 -> 현재 기울기 양수
	if *prevSlope < 0 && *currSlope > 0 {
		return "BUY"
	}
	// 역V자 하락: 이전 기울기 양수 -> 현재 기울기 음수
	if *prevSlope > 0 && *currSlope < 0 {
		return "SELL"
	}

	return "HOLD"
}

// ApplyCost 비용 적용 (수익률 계산용)
func ApplyCost(tradeType string, price float64, config types.BotConfiguration) float64 {
	if tradeType == "BUY" {
		return price * (1 + config.SlippageRate) * (1 + config.FeeRate)
	}
	// SELL
	return price * (1 - config.SlippageRate) * (1 - config.FeeRate)
}
