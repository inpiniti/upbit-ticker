package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"upbit-ticker/types"
	"upbit-ticker/websocket"
)

func main() {
	log.Println("🚀 업비트 실시간 티커 시작")

	// 웹소켓 클라이언트 생성 (KRW-BTC 구독)
	client := websocket.NewClient([]string{"KRW-BTC"})

	// onTick 이벤트 핸들러 등록
	client.OnTick(onTick)

	// 웹소켓 연결
	if err := client.Connect(); err != nil {
		log.Fatalf("연결 실패: %v", err)
	}

	// 티커 구독
	if err := client.Subscribe(); err != nil {
		log.Fatalf("구독 실패: %v", err)
	}

	// 메시지 수신 시작
	client.Start()

	// 종료 시그널 대기
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("👋 프로그램 종료 중...")
	client.Stop()
}

// onTick 틱 데이터 수신 시 호출되는 핸들러
func onTick(tick types.Ticker) {
	// 변동 표시
	changeSymbol := "━"
	changeColor := "\033[0m" // 기본
	
	switch tick.Change {
	case "RISE":
		changeSymbol = "▲"
		changeColor = "\033[31m" // 빨간색 (상승)
	case "FALL":
		changeSymbol = "▼"
		changeColor = "\033[34m" // 파란색 (하락)
	}

	// 시간 포맷
	timestamp := time.UnixMilli(tick.Timestamp)
	timeStr := timestamp.Format("15:04:05")

	// 현재가 출력
	fmt.Printf(
		"[%s] %s %s현재가: %,.0f원 %s %s%.2f%% (%+,.0f원)\033[0m\n",
		timeStr,
		tick.Code,
		changeColor,
		tick.TradePrice,
		changeSymbol,
		changeColor,
		tick.SignedChangeRate*100,
		tick.SignedChangePrice,
	)
}
