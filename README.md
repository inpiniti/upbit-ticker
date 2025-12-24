# Upbit Ticker (Wails + React)

Go(Wails)와 React를 활용하여 업비트 실시간 시세를 조회하고 SQLite에 저장하는 데스크탑 애플리케이션입니다.

## 🛠 기술 스택

- **Backend**: Go (Wails Framework)
- **Frontend**: React, TypeScript, TailwindCSS, Zustand
- **Database**: SQLite (Gorm)
- **API**: Upbit WebSocket API

## 📁 프로젝트 구조

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

## 📊 데이터베이스
앱 실행 시 `upbit_ticker.db` 파일이 자동 생성되며 실시간 수신된 티커 데이터가 저장됩니다.

## 📊 아래 GPT와 대화하고 만든 js 코드임 (추후 프로젝트에도 반영하면 좋을듯)

```javascript
// --- 설정 (Configuration) ---
// 화면에서 변경 가능한 옵션값들
const CONFIG = {
  INTERVAL_MS: 60 * 1000, 
  SLIPPAGE_RATE: 0.0002,  
  FEE_RATE: 0.0005        
}

// --- 전역 상태 (Global State) ---
let appState = {
  intervalBuffer: [],
  intervalStartTime: 0,
  prevAverage: null,
  prevSlope: null,
  isHolding: false
}

// --- 메인 이벤트 핸들러 (Entry Point) ---
function onTick(rawTick) {
  // 1. [I/O] 저장
  db.saveRawTick({ ts: rawTick.ts, price: rawTick.price })

  // 2. [Logic] 실행 
  // processTick은 순수 상태 변경과 '발생한 이벤트(Signal)'를 반환함
  const result = processTick(appState, rawTick, CONFIG)
  
  // 3. 상태 업데이트
  appState = result.newState
  
  // 4. [Side Effect] 매매 기록
  if (result.tradeEvent) {
    recordTrade(result.tradeEvent, rawTick, CONFIG)
  }
}

// --- [Optimzer] 백테스트 최적화 함수 ---
// 1초 ~ 24시간까지 모든 구간을 시뮬레이션하여 최적의 수익률 구간을 찾음
function findSweetSpot(allTicks) {
  const results = []
  
  // 탐색 범위 생성 (1초, 2초... 1분... 24시간)
  const testIntervals = generateTestIntervals()
  
  // 각 구간별 시뮬레이션 실행 (Go에서는 Goroutine 병렬 처리 권장)
  testIntervals.forEach(intervalMs => {
    const { profit, tradeCount } = runSimulation(allTicks, intervalMs)
    
    results.push({
      intervalMs,
      profit,
      tradeCount
    })
  })
  
  // 수익률 높은 순 정렬
  results.sort((a, b) => b.profit - a.profit)
  
  // 최적 결과 반환 (Top 1)
  console.log(`Best Interval: ${results[0].intervalMs / 1000}s, Profit: ${results[0].profit}`)
  return results[0]
}

// 시뮬레이션 실행기 (In-Memory Backtest)
function runSimulation(ticks, intervalMs) {
  // 시뮬레이션용 격리된 상태 (매번 초기화)
  let simState = {
    intervalBuffer: [], intervalStartTime: 0,
    prevAverage: null, prevSlope: null, isHolding: false
  }
  let totalProfit = 0
  let tradeCount = 0
  let entryPrice = 0
  
  // 테스트용 설정 (Interval만 변경)
  const simConfig = { ...CONFIG, INTERVAL_MS: intervalMs }

  ticks.forEach(tick => {
    // 순수 로직 processTick 재사용
    const { newState, tradeEvent } = processTick(simState, tick, simConfig)
    
    // 매매 손익 계산 (Profit Calculation)
    if (tradeEvent === 'BUY') {
      entryPrice = applyCost('BUY', tick.price, simConfig)
    } else if (tradeEvent === 'SELL') {
      const exitPrice = applyCost('SELL', tick.price, simConfig)
      totalProfit += (exitPrice - entryPrice)
      tradeCount++
    }
    
    simState = newState
  })

  // 마지막에 보유 중이면 현재가 청산 가정 (선택사항, 보통은 청산 후 수익 확정)
  // if (simState.isHolding) { ... }

  return { profit: totalProfit, tradeCount }
}

// 테스트 구간 생성 헬퍼
function generateTestIntervals() {
  const list = []
  // 1초 ~ 59초
  for (let s = 1; s < 60; s++) list.push(s * 1000)
  // 1분 ~ 24시간 (분 단위)
  for (let m = 1; m <= 60 * 24; m++) list.push(m * 60 * 1000)
  return list
}


// --- 1. 순수 로직 (Pure Functions) ---

// 핵심 로직: 상태(State) + 입력(Tick) -> 새로운 상태(NewState) + 이벤트(Event)
function processTick(state, tick, config) {
  const nextBuffer = [...state.intervalBuffer, tick]
  const startTime = state.intervalBuffer.length === 0 ? tick.ts : state.intervalStartTime
  
  // 구간 종료 확인
  const isIntervalFinished = (tick.ts - startTime) >= config.INTERVAL_MS
  
  if (!isIntervalFinished) {
    return {
      newState: {
        ...state,
        intervalBuffer: nextBuffer,
        intervalStartTime: startTime
      },
      tradeEvent: null
    }
  }

  // --- 구간 완성 시 로직 ---
  const currentAverage = calculateAverage(nextBuffer)
  const currentSlope = calculateSlope(currentAverage, state.prevAverage)
  const signal = evaluateSignal(state.prevSlope, currentSlope)
  
  let tradeEvent = null
  let nextIsHolding = state.isHolding

  // 매매 신호 처리
  if (signal === 'BUY' && !state.isHolding) {
    tradeEvent = 'BUY'
    nextIsHolding = true
  } else if (signal === 'SELL' && state.isHolding) {
    tradeEvent = 'SELL'
    nextIsHolding = false
  }

  return {
    newState: {
      ...state,
      intervalBuffer: [],
      intervalStartTime: 0,
      prevAverage: currentAverage,
      prevSlope: currentSlope,
      isHolding: nextIsHolding
    },
    tradeEvent: tradeEvent
  }
}

// 평균 계산
function calculateAverage(ticks) {
  if (ticks.length === 0) return 0
  return ticks.reduce((acc, t) => acc + t.price, 0) / ticks.length
}

// 기울기 계산
function calculateSlope(curr, prev) {
  if (prev === null) return null
  return curr - prev
}

// 신호 평가
function evaluateSignal(prevSlope, currSlope) {
  if (prevSlope === null || currSlope === null) return 'HOLD'
  if (prevSlope > 0 && currSlope < 0) return 'BUY' // **FIX: V자 반등은 (음수 -> 양수)**
  if (prevSlope < 0 && currSlope > 0) return 'SELL' // **Wait, original V-shape logic was neg->pos=BUY**
  // Let's re-verify the logic requested:
  // "이전평균보다 현재평균이 낮으면 음수, 높으면 양수" (Slope = Curr - Prev)
  // "이전 slope 음수 -> 현재 slope 양수 : 매수 신호 (V자 반등)" (Correct)
  // "이전 slope 양수 -> 현재 slope 음수 : 매도 신호 (역V자)" (Correct)
  
  if (prevSlope < 0 && currSlope > 0) return 'BUY'
  if (prevSlope > 0 && currSlope < 0) return 'SELL'
  
  return 'HOLD'
}

// 비용 적용 (가격 보정)
function applyCost(type, price, config) {
  if (type === 'BUY') return price * (1 + config.SLIPPAGE_RATE) * (1 + config.FEE_RATE)
  return price * (1 - config.SLIPPAGE_RATE) * (1 - config.FEE_RATE)
}

// --- 2. Side Effect (DB) ---
function recordTrade(type, tick, config) {
  db.insertTrade({
    ts: tick.ts,
    price: tick.price,
    saleflag: type,
    executionPrice: applyCost(type, tick.price, config) // 수익률 계산용
  })
}
```

### 📋 Q&A 반영 사항

**Q9. 구간 옵션을 화면단에서 1초, 1분, 1시간 등으로 변경하면 차트 변경이 될까?**
- **가능합니다.**
- 원본 데이터(`rawTick` - ts, price)를 모두 DB에 저장하고 있기 때문에, 옵션(`CONFIG.INTERVAL_MS`)만 변경하고 `onTick` 로직을 저장된 데이터에 대해 처음부터 다시 돌리면(Re-calculation), 해당 구간 기준의 새로운 `Average`, `Slope` 그래프와 매매 타점을 즉시 다시 그려낼 수 있습니다.

**Q10. 차트에 매매기록(빨간점, 파란점) 표시가 될까?**
- **가능합니다.**
- 차트 라이브러리(Recharts 등)에서 Scatter Chart(산점도)를 Line Chart 위에 중첩(ComposedChart)시킬 수 있습니다.
- `Trade` 테이블의 데이터를 읽어서 매수(`BUY`)는 빨간색, 매도(`SELL`)는 파란색 점으로 좌표(`ts`, `price`)에 찍어주면 됩니다.

**Q11. 차트는 tick 선, average 점선, 매매기록 점으로 표현 가능할까?**
- **가능합니다.**
- **Tick (선)**: 전체 Raw Tick 데이터를 얇은 실선으로 그립니다.
- **Average (점선)**: 계산된 구간별 Average 값을 점선(strokeDasharray)으로 Tick 위에 겹쳐서 그립니다.
- **매매기록 (점)**: 위에서 언급한 대로 Scatter 그래프를 가장 상위 레이어에 그리면 됩니다.
- 이렇게 하면 한눈에 시세 흐름, 추세선(Average), 그리고 매매 타점을 파악할 수 있는 훌륭한 백테스팅 차트가 됩니다.