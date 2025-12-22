package types

// Ticker 업비트 현재가 데이터 구조체
type Ticker struct {
	Type               string  `json:"type"`                 // 타입 (ticker)
	Code               string  `json:"code"`                 // 마켓 코드 (ex. KRW-BTC)
	OpeningPrice       float64 `json:"opening_price"`        // 시가
	HighPrice          float64 `json:"high_price"`           // 고가
	LowPrice           float64 `json:"low_price"`            // 저가
	TradePrice         float64 `json:"trade_price"`          // 현재가
	PrevClosingPrice   float64 `json:"prev_closing_price"`   // 전일 종가
	Change             string  `json:"change"`               // 전일 대비 (RISE, EVEN, FALL)
	ChangePrice        float64 `json:"change_price"`         // 전일 대비 값
	SignedChangePrice  float64 `json:"signed_change_price"`  // 전일 대비 값 (부호 포함)
	ChangeRate         float64 `json:"change_rate"`          // 전일 대비 등락율
	SignedChangeRate   float64 `json:"signed_change_rate"`   // 전일 대비 등락율 (부호 포함)
	TradeVolume        float64 `json:"trade_volume"`         // 가장 최근 거래량
	AccTradeVolume     float64 `json:"acc_trade_volume"`     // 누적 거래량 (UTC 0시 기준)
	AccTradeVolume24h  float64 `json:"acc_trade_volume_24h"` // 24시간 누적 거래량
	AccTradePrice      float64 `json:"acc_trade_price"`      // 누적 거래대금 (UTC 0시 기준)
	AccTradePrice24h   float64 `json:"acc_trade_price_24h"`  // 24시간 누적 거래대금
	TradeDate          string  `json:"trade_date"`           // 최근 거래 일자 (UTC)
	TradeTime          string  `json:"trade_time"`           // 최근 거래 시각 (UTC)
	TradeTimestamp     int64   `json:"trade_timestamp"`      // 체결 타임스탬프
	AskBid             string  `json:"ask_bid"`              // 매수/매도 구분 (ASK, BID)
	AccAskVolume       float64 `json:"acc_ask_volume"`       // 누적 매도량
	AccBidVolume       float64 `json:"acc_bid_volume"`       // 누적 매수량
	Highest52WeekPrice float64 `json:"highest_52_week_price"` // 52주 최고가
	Highest52WeekDate  string  `json:"highest_52_week_date"`  // 52주 최고가 달성일
	Lowest52WeekPrice  float64 `json:"lowest_52_week_price"`  // 52주 최저가
	Lowest52WeekDate   string  `json:"lowest_52_week_date"`   // 52주 최저가 달성일
	MarketState        string  `json:"market_state"`          // 거래 상태
	Timestamp          int64   `json:"timestamp"`             // 타임스탬프
	StreamType         string  `json:"stream_type"`           // 스트림 타입 (SNAPSHOT, REALTIME)
}
