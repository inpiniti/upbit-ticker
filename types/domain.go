package types

import "time"

// RawTick 분석에 필요한 최소 데이터
type RawTick struct {
	Timestamp int64   `json:"ts"`
	Price     float64 `json:"price"`
}

// BotConfiguration 트레이딩 봇 설정
type BotConfiguration struct {
	IntervalDuration time.Duration `json:"interval_duration"`
	SlippageRate     float64       `json:"slippage_rate"`
	FeeRate          float64       `json:"fee_rate"`
}

// MarketState 트레이딩 시스템의 현재 상태 (Immutable 지향)
type MarketState struct {
	IntervalBuffer    []RawTick `json:"interval_buffer"`
	IntervalStartTime int64     `json:"interval_start_time"`
	PrevAverage       *float64  `json:"prev_average"`
	PrevSlope         *float64  `json:"prev_slope"`
	IsHolding         bool      `json:"is_holding"`
}

// ProcessResult Tick 처리 결과
type ProcessResult struct {
	NewState       MarketState `json:"new_state"`
	TradeSignal    string      `json:"trade_signal"`    // "BUY", "SELL", or ""
	IntervalClosed bool        `json:"interval_closed"` // 구간 종료 여부
	CurrentAverage float64     `json:"current_average"`
	CurrentSlope   *float64    `json:"current_slope"`
}

// TradeRecord 매매 기록
type TradeRecord struct {
	Timestamp      int64   `json:"ts"`
	Price          float64 `json:"price"`
	Type           string  `json:"type"`            // "BUY" or "SELL"
	ExecutionPrice float64 `json:"execution_price"` // 수수료/슬리피지 적용가
}

// OptimizationResult 최적화 결과 (사이클 통계 포함)
type OptimizationResult struct {
	IntervalDuration time.Duration `json:"interval_duration"`
	Profit           float64       `json:"profit"`      // 순수익
	CycleCount       int           `json:"cycle_count"` // 총 사이클 수 (BUY-SELL 쌍)
	WinCount         int           `json:"win_count"`   // 수익 사이클 수
	LossCount        int           `json:"loss_count"`  // 손실 사이클 수
	WinRate          float64       `json:"win_rate"`    // 승률 (0.0 ~ 1.0)
	AvgWin           float64       `json:"avg_win"`     // 평균 수익 (수익 사이클)
	AvgLoss          float64       `json:"avg_loss"`    // 평균 손실 (손실 사이클)
}
