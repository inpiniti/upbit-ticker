# 업비트 실시간 티커 (Upbit Ticker)

Go 언어로 작성된 업비트 웹소켓 실시간 현재가 수신 프로젝트입니다.

## 📁 프로젝트 구조

```
upbit-ticker/
├── go.mod              # Go 모듈 파일
├── go.sum              # 의존성 체크섬
├── main.go             # 메인 진입점
├── types/
│   └── ticker.go       # Ticker 타입 정의
└── websocket/
    └── client.go       # 웹소켓 클라이언트
```

## 🚀 실행 방법

### 1. Go 설치 확인

먼저 Go가 설치되어 있는지 확인하세요:

```bash
go version
```

Go가 설치되어 있지 않다면 [Go 공식 사이트](https://go.dev/dl/)에서 다운로드하세요.

### 2. 의존성 다운로드

```bash
cd upbit-ticker
go mod tidy
```

### 3. 실행

```bash
go run .
```

또는 빌드 후 실행:

```bash
go build -o upbit-ticker.exe
./upbit-ticker.exe
```

## 📡 기능

- **웹소켓 연결**: 업비트 공개 웹소켓 API 연결
- **실시간 티커 구독**: KRW-BTC 현재가 실시간 수신
- **onTick 이벤트**: 틱 데이터 수신 시 콜백 함수 호출
- **컬러 출력**: 상승(빨간색)/하락(파란색) 표시

## 📋 출력 예시

```
🚀 업비트 실시간 티커 시작
✅ 업비트 웹소켓 연결 성공
📡 구독 요청 완료: [KRW-BTC]
[15:30:45] KRW-BTC 현재가: 145,230,000원 ▲ +2.35% (+3,330,000원)
[15:30:46] KRW-BTC 현재가: 145,225,000원 ▼ +2.35% (+3,325,000원)
```

## 🔧 커스터마이징

### 다른 코인 구독하기

`main.go`에서 구독할 코인을 변경할 수 있습니다:

```go
// 여러 코인 구독
client := websocket.NewClient([]string{"KRW-BTC", "KRW-ETH", "KRW-XRP"})
```

### onTick 핸들러 수정하기

`main.go`의 `onTick` 함수를 수정하여 원하는 로직을 추가하세요:

```go
func onTick(tick types.Ticker) {
    // 여기에 원하는 로직 추가
    // 예: 데이터베이스 저장, 매매 신호 분석 등
}
```

## 📚 API 참고

- [업비트 웹소켓 API 문서](https://docs.upbit.com/docs/upbit-quotation-websocket)

## 아래 GPT와 대화하고 만든 코드임 (추후 프로젝트에도 반영하면 좋을듯)

```
tick 구조
{
  ts,
  price,
  ma20,
  ma20Slope,
  ma20Accel
}

ticks : tick 이 들어올때마다 쌓음

// 0.02% (BTC 기준 현실적)
const SLIPPAGE_RATE = 0.0002
const FEE_RATE = 0.0005       // 0.05%

// 상태
let prevSignal = 'HOLD'

// 백테스트
historicalTicks.forEach(tick => onTick(tick))

// 웹소켓 수신
onTick(tick) {
  ticks = updateWindow(ticks, tick)
  ticks = indicators.calculate(ticks)
  
  const last = getLastTick(ticks)
  const currentSignal = evaluateSignal(last)

  if (isSignalEdge(prevSignal, currentSignal)) {
    trading(currentSignal, last)
  }

  prevSignal = currentSignal
}

// 매매
trading(signal, last) {
  if (signal === 'BUY') onBuy(last)
  if (signal === 'SELL') onSell(last)
}

// 매수
onBuy(tick) {
  // const executionPrice = applySlippage('BUY', tick.price)
  const executionPrice = applyBuyCost(tick.price)
}

// 매도
onSell(tick) {
  // const executionPrice = applySlippage('SELL', tick.price)
  const executionPrice = applySellProceeds(tick.price)
}

// 체결가 (슬리피지)
applySlippage(side, price) {
  if (side === 'BUY') {
    return price * (1 + SLIPPAGE_RATE)
  }
  if (side === 'SELL') {
    return price * (1 - SLIPPAGE_RATE)
  }
  return price
}

applyBuyCost(price) {
  const withSlippage = price * (1 + SLIPPAGE_RATE)
  const withFee = withSlippage * (1 + FEE_RATE)
  return withFee
}

applySellProceeds(price) {
  const withSlippage = price * (1 - SLIPPAGE_RATE)
  const withFee = withSlippage * (1 - FEE_RATE)
  return withFee
}

// edge 판단 함수
isSignalEdge(prev, curr) {
  if (prev !== 'BUY' && curr === 'BUY') return true
  if (prev !== 'SELL' && curr === 'SELL') return true
  return false
}

// 수신한 데이터 추가
indicators.calculate = (ticks) =>
  pipe(
    addMa20,
    addMa20Slope,
    addMa20Accel
  )(ticks)

// 매매 시그널
evaluateSignal(lastTick) {
  if (lastTick.ma20Accel > 0.1) return 'BUY'
  if (lastTick.ma20Accel < -0.1) return 'SELL'
  return 'HOLD'
}
```