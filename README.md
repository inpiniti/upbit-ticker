# Upbit Ticker (Wails + React)

Go(Wails)와 React를 활용하여 업비트 실시간 시세를 조회하고 SQLite에 저장하는 데스크탑 애플리케이션입니다.

## � 기술 스택

- **Backend**: Go (Wails Framework)
- **Frontend**: React, TypeScript, TailwindCSS, Zustand
- **Database**: SQLite (Gorm)
- **API**: Upbit WebSocket API

## �📁 프로젝트 구조

```
upbit-ticker/
├── apps.go             # Wails 애플리케이션 로직 (DB, WS 연동)
├── main.go             # 메인 진입점 (Wails 설정)
├── wails.json          # Wails 프로젝트 설정
├── frontend/           # React 프론트엔드
│   ├── src/
│   │   ├── store/      # Zustand 상태 관리
│   │   └── App.tsx     # UI 컴포넌트
│   └── wailsjs/        # Wails 자동 생성 (빌드 시 생성됨)
├── types/              # 공용 데이터 타입
└── websocket/          # WebSocket 클라이언트 패키지
```

## 🚀 실행 방법

### 1. 필수 요구사항
- [Go](https://go.dev/dl/) 1.18+
- [Node.js](https://nodejs.org/) 16+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 2. 개발 모드 실행
프론트엔드와 백엔드를 동시에 실행하며 변경 사항을 실시간으로 반영합니다.

```bash
wails dev
```
최초 실행 시 `frontend` 의존성을 자동으로 설치하므로 시간이 소요될 수 있습니다.

### 3. 프로덕션 빌드

```bash
wails build
```
`build/bin` 폴더에 실행 파일이 생성됩니다.

## � 데이터베이스
앱 실행 시 `upbit_ticker.db` 파일이 자동 생성되며 실시간 수신된 티커 데이터가 저장됩니다.

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