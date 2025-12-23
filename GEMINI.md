# 프로젝트 구현 계획 (Based on README.md)

README.md에 포함된 "GPT와 작성한 알고리즘(이동평균 가속도 기반 트레이딩)"을 현재 Wails + React 프로젝트에 통합하기 위한 단계별 계획입니다.

## 1. 개요 및 목표
- **핵심 로직**: 이동평균선(MA20)의 가속도(`Accel`)를 기반으로 매수/매도 시그널을 생성.
- **아키텍처**: 
  - **Go (Backend)**: 실시간 데이터 수집, 지표 계산, 신호 발생, 가상 매매 로직 수행.
  - **React (Frontend)**: 계산된 지표 및 매매 신호 시각화, 수익률 대시보드.

---

## 2. 상세 구현 계획

### Phase 1: 백엔드 지표 계산 로직 (Go)
데이터의 수집과 가공을 담당하는 핵심 모듈을 개발합니다.

1.  **데이터 구조 정의 (`types/ticker.go` 등)**
    *   `Tick` 구조체에 분석용 필드 추가 (DB 저장용과 인메모리 분석용 구분 필요 가능성).
    *   필드: `Price`, `MA20`, `Slope` (기울기), `Accel` (가속도), `Timestamp`.


2.  **데이터 윈도우 관리 및 저장 최적화**
    *   **Sliding Window**: 이동평균 계산을 위해 일정 개수(예: 20개 이상)의 과거 Tick을 메모리에 유지.
    *   **Batch Insert (DB 최적화)**: 매 Tick마다 DB에 접근하지 않고, 버퍼(Buffer)에 모았다가 일정 개수(`BatchSize`)나 일정 시간(`Interval`)이 차면 한 번에 저장.
    *   예: `tickBuffer`에 쌓고, 100개가 되거나 1초가 지나면 SQLite에 Bulk Insert.

3.  **지표 계산 함수 (`indicators` 패키지)**
    *   `CalculateMA20(ticks)`: 단순 이동평균 계산.
    *   `CalculateSlope(prevMA, currMA)`: 변화량(기울기) 계산.
    *   `CalculateAccel(prevSlope, currSlope)`: 변화량의 변화량(가속도) 계산.

### Phase 2: 매매 전략 및 시뮬레이션 엔진 (Go)
정의된 규칙에 따라 매매 신호를 생성하고 가상 매매를 수행합니다.

1.  **전략 로직 (`strategy` 패키지)**
    *   **Signal Evaluation**:
        *   `BUY`: `Accel > 0.1` (가속도가 양수 임계값 초과)
        *   `SELL`: `Accel < -0.1` (가속도가 음수 임계값 미만)
        *   `HOLD`: 그 외.
    *   **Edge Detection**: 상태가 변하는 순간(`prev != curr`)만 감지하여 이벤트 트리거.

2.  **가상 매매 (Paper Trading)**
    *   상수 정의:
        *   `SLIPPAGE_RATE = 0.0002` (0.02%)
        *   `FEE_RATE = 0.0005` (0.05%)

    *   **비용 계산**:
        *   매수 시: `Price * (1 + Slippage) * (1 + Fee)`
        *   매도 시: `Price * (1 - Slippage) * (1 - Fee)`
    *   **포트폴리오 관리**: 가상 잔고 및 보유 수량, 평단가 추적.

3.  **매매 이력 관리 (Trade History)**
    *   **데이터 구조 (`Trade`)**:
        *   `ID`, `BuyPrice`, `BuyTime`, `SellPrice`, `SellTime`, `Profit` (수익금/률).
    *   **로직 (Stack/LIFO 방식)**:
        *   **Buy**: 새로운 `Trade` 레코드 생성 (Insert). `SellPrice/Time`은 `NULL`.
        *   **Sell**: `SellTime`이 `NULL`인 가장 최근 레코드를 조회하여 `SellPrice/Time` 업데이트 (Update).

### Phase 3: 프론트엔드 시각화 (React)
백엔드에서 처리된 정보를 사용자에게 보여줍니다.

1.  **데이터 연동**
    *   Wails 이벤트(`Runtime.EventsEmit`)를 통해 매 Tick마다 계산된 지표 및 신호 수신.
    *   Zustand Store에 실시간 데이터 업데이트.

2.  **UI 컴포넌트 개발**
    *   **대시보드**: 현재가, MA20, 기울기, 가속도 수치 표시.
    *   **차트**: 캔들/라인 차트 위에 매수/매도 시점 마킹.
    *   **로그 창**: 발생한 시그널 및 체결 내역(가상) 리스트 출력.

---

## 3. 원본 알고리즘 참조 (JS -> Go 변환 가이드)

| 항목 | JS 원본 | Go 구현 방향 |
| :--- | :--- | :--- |
| **변수 관리** | `ticks` (Array), `prevSignal` (Let) | `struct` 내 슬라이스 및 필드로 상태 관리 |
| **파이프라인** | `pipe(addMa20, ...)` | 함수 체이닝 또는 절차적 호출로 구현 |
| **신호 감지** | `isSignalEdge` | `if prevSignal != currentSignal` 로직 적용 |
| **비용 적용** | `applyBuyCost`, `applySellProceeds` | 별도 유틸리티 함수 또는 메서드로 분리 |

---

## 4. Next Steps
우선순위에 따른 작업 순서입니다.


1.  [ ] Go: `types` 패키지에 분석용 구조체 정의.
2.  [ ] Go: Tick 버퍼링 및 배치 저장(Batch Insert) 로직 구현.
3.  [ ] Go: 메모리 내 MA/Slope/Accel 계산 로직 작성.
4.  [ ] Go: 매매 이력 저장(Insert/Update) 및 관리 로직 구현.
5.  [ ] Go: WebSocket 수신부(`onTick`)에 계산 및 저장 로직 통합.
