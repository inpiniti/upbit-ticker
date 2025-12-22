package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"upbit-ticker/types"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// UpbitWebsocketURL 업비트 웹소켓 엔드포인트
	UpbitWebsocketURL = "wss://api.upbit.com/websocket/v1"
)

// TickHandler onTick 이벤트 핸들러 타입
type TickHandler func(tick types.Ticker)

// Client 업비트 웹소켓 클라이언트
type Client struct {
	conn      *websocket.Conn
	onTick    TickHandler
	codes     []string
	isRunning bool
	stopChan  chan struct{}
}

// NewClient 새로운 웹소켓 클라이언트 생성
func NewClient(codes []string) *Client {
	return &Client{
		codes:    codes,
		stopChan: make(chan struct{}),
	}
}

// OnTick onTick 이벤트 핸들러 등록
func (c *Client) OnTick(handler TickHandler) {
	c.onTick = handler
}

// Connect 웹소켓 연결
func (c *Client) Connect() error {
	// 업비트 API 연결에 필요한 HTTP 헤더 설정
	header := http.Header{}
	header.Add("Origin", "https://upbit.com")
	header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	header.Add("Accept-Language", "ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7")

	conn, resp, err := websocket.DefaultDialer.Dial(UpbitWebsocketURL, header)
	if err != nil {
		if resp != nil {
			log.Printf("HTTP 응답 상태: %s", resp.Status)
			// 응답 본문 읽기
			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			if n > 0 {
				log.Printf("응답 본문: %s", string(body[:n]))
			}
		}
		return fmt.Errorf("웹소켓 연결 실패: %w", err)
	}
	c.conn = conn
	log.Println("✅ 업비트 웹소켓 연결 성공")
	return nil
}

// Subscribe 티커 구독 요청
func (c *Client) Subscribe() error {
	// 구독 메시지 생성
	subscribeMsg := []map[string]interface{}{
		{
			"ticket": uuid.New().String(),
		},
		{
			"type":  "ticker",
			"codes": c.codes,
		},
	}

	msgBytes, err := json.Marshal(subscribeMsg)
	if err != nil {
		return fmt.Errorf("구독 메시지 생성 실패: %w", err)
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		return fmt.Errorf("구독 요청 실패: %w", err)
	}

	log.Printf("📡 구독 요청 완료: %v\n", c.codes)
	return nil
}

// Start 메시지 수신 시작
func (c *Client) Start() {
	c.isRunning = true

	go func() {
		for c.isRunning {
			select {
			case <-c.stopChan:
				return
			default:
				_, message, err := c.conn.ReadMessage()
				if err != nil {
					if c.isRunning {
						log.Printf("메시지 수신 오류: %v\n", err)
					}
					return
				}

				// Ticker 파싱
				var tick types.Ticker
				if err := json.Unmarshal(message, &tick); err != nil {
					log.Printf("메시지 파싱 오류: %v\n", err)
					continue
				}

				// onTick 이벤트 호출
				if c.onTick != nil {
					c.onTick(tick)
				}
			}
		}
	}()
}

// Stop 웹소켓 연결 종료
func (c *Client) Stop() {
	c.isRunning = false
	close(c.stopChan)
	if c.conn != nil {
		c.conn.Close()
		log.Println("🔌 웹소켓 연결 종료")
	}
}
