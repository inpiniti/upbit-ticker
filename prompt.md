# Agent Prompt: 이동평균 가속도 기반 트레이딩 봇 구현 (함수형 패러다임)

당신은 Go(Wails)와 React를 사용하는 전문 풀스택 개발자입니다.
현재 프로젝트(`upbit-ticker`)에는 기본적인 Wails 설정, React 프론트엔드, SQLite 연결, 그리고 Upbit WebSocket 연동(`websocket` 패키지)이 구현되어 있습니다.

이제 `README.md`에 있는 **자바스크립트 프로토타입 코드(FP 스타일)**를 기반으로, **이동평균(MA20) 가속도 기반 트레이딩 알고리즘**을 구현해야 합니다.

## ⚠️ 핵심 코딩 원칙: 함수형 프로그래밍 (Functional Programming)
**모든 핵심 로직은 반드시 "순수 함수(Pure Function)"로 구현해야 합니다.**
1.  **순수 함수 (Pure Function)**:
    *   함수 내부에서 전역 변수나 외부 상태를 절대 변경하지 마세요.
    *   동일한 입력에 대해 항상 동일한 출력을 반환해야 합니다.
    *   모든 필요한 데이터는 인자(Argument)로 전달받아야 합니다.
2.  **불변성 (Immutability)**:
    *   입력받은 구조체나 슬라이스를 직접 수정(Mutation)하지 마세요.
    *   변경된 상태를 가진 **새로운 객체(또는 복사본)**를 반환하세요.
    *   예: `CalculateMA20(ticks []Tick)` -> `Tick` (Update된 틱이 아닌 계산 결과만 반환하거나, 새 구조체 반환)
3.  **상태 관리 (State Management)**:
    *   상태(`State` 구조체)는 로직 함수들이 주고받는 데이터일 뿐입니다.
    *   상태의 영속성(Persistence)과 I/O(DB 저장, 로그)는 비즈니스 로직(순수 함수) 밖의 최상위 계층(`Effect` 영역, 예: `onTick` 핸들러)에서만 처리하세요.

---

## 구현 단계

### 1. 데이터 구조 정의 (Immutable Data Structures)
`types` 또는 `domain` 패키지에 상태를 표현하는 불변 데이터 구조를 정의하세요.
- **AnalyzedTick**: `Tick` 데이터와 분석 결과(`MA20`, `Slope`, `Accel`)를 포함하는 구조체.
- **MarketState**: 트레이딩 시스템의 전체 상태를 담는 구조체 (예: `Window` (최근 Tick들의 리스트), `Positions` (현재 매수/매도 상태), `TradeHistory`).

### 2. 순수 함수 로직 구현 (`core` 또는 `logic` 패키지)
다음 함수들을 **순수 함수**로 구현하세요. (메서드 리시버보다는 일반 함수 `func Function(state, input) newState` 형태 권장)

#### A. 데이터 가공 및 윈도우 관리
- `UpdateWindow(window []Tick, newTick Tick) []Tick`:
    - 기존 윈도우에 새 틱을 추가하고, 최대 크기(예: 20+a)를 유지한 **새로운 슬라이스**를 반환합니다.

#### B. 지표 계산 (Indicator Calculation)
- `CalculateIndicators(window []Tick) AnalyzedTick`:
    - 윈도우 데이터를 기반으로 최신 틱의 `MA20`, `Slope`, `Accel`을 계산하여 반환합니다.
    - JS의 `pipe(addMa20, addSlope, ...)` 패턴처럼 작은 계산 함수들을 조합해도 좋습니다.

#### C. 전략 평가 (Strategy Evaluation)
- `EvaluateSignal(tick AnalyzedTick) Signal`:
    - `BUY` (`Accel > 0.1`) / `SELL` (`Accel < -0.1`) / `HOLD`여부를 판단하여 반환합니다.
- `IsSignalEdge(prevSignal Signal, currSignal Signal) bool`:
    - 신호가 변하는 시점인지, 상태 변화가 없는지 판단합니다.

#### D. 매매 시뮬레이션 (Virtual Execution)
- `ExecuteTrade(prevState MarketState, signal Signal, price float64) MarketState`:
    - 현재 상태와 신호에 따라 가상 매매를 수행하고, **업데이트된 새로운 MarketState**를 반환합니다.
    - 비용(Slippage 0.02%, Fee 0.05%)을 적용한 수익률 계산 로직을 포함합니다.

### 3. I/O 및 효과 처리 (Effect Handling)
**이 부분은 순수 함수가 아니며, 실제 상태 변경과 DB 저장을 담당합니다.** (`main.go` 또는 `app.go`의 `onTick` 핸들러)

- **Workflow**:
  1. **Read**: 현재 상태(`currentMarketState`)를 읽어옵니다.
  2. **Calculate (Pure)**:
     - `nextWindow = UpdateWindow(current.Window, incomingTick)`
     - `analyzedTick = CalculateIndicators(nextWindow)`
     - `signal = EvaluateSignal(analyzedTick)`
     - `nextMarketState = ExecuteTrade(current, signal, newPrice)`
  3. **Write (Side-Effect)**:
     - 계산된 `nextMarketState`로 애플리케이션 상태 포인터 업데이트.
     - `analyzedTick`을 버퍼에 추가 (Batch Insert용).
     - 매매 발생 시 `trades` 테이블에 Insert/Update.
     - Wails Event Emit (Frontend로 전송).

### 4. 데이터베이스 및 배치 처리
- **SQLite & Gorm**: `ticks`, `trades` 테이블 모델링.
- **Batcher**: Tick 데이터를 100개씩 모아서 DB에 저장하는 로직 (이 녀석은 본질적으로 Impure하므로 별도 관리).

### 5. 프론트엔드 (React + Zustand)
- 백엔드에서 날아온 불변 데이터를 받아 그대로 렌더링하는데 집중하세요.
- `useTrendStore` 등을 사용하여 상태 흐름을 단방향으로 유지하세요.

---

**요약**: "도메인 로직(계산)"과 "실행 컨텍스트(I/O, 상태저장)"를 철저히 분리하여, JS 코드의 함수형 스타일을 Go 언어로 우아하게 번역해주세요.
