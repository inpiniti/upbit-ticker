# Wails + React + SQLite + Go Development Guidelines

이 문서는 Wails 프레임워크를 기반으로 React 프론트엔드와 Go 백엔드, 그리고 SQLite 데이터베이스를 사용하는 프로젝트의 개발 가이드라인입니다.

## 1. 아키텍처 개요 (Architecture Overview)

-   **Frontend**: React (UI), Zustand (State), TailwindCSS (Styling)
-   **Backend**: Go (Business Logic, Data Processing, SQLite Interface)
-   **Bridge**: Wails (Frontend-Backend Communication via `window.go.main.App`)
-   **Database**: SQLite (Local Embedded DB)

---

## 2. Go (Backend) 개발 가이드라인

### 2.1 코드 구조 및 네이밍
-   **패키지 구조**: 핵심 로직은 `internal` 패키지에, 재사용 가능한 라이브러리는 `pkg`에 배치합니다.
-   **네이밍**: Go 표준 컨벤션(`camelCase` for internal, `PascalCase` for exported)을 따릅니다.
-   **모듈화**: 기능별로 패키지를 분리하여 의존성을 관리합니다 (예: `database`, `services`, `models`).

### 2.2 에러 처리 및 로깅
-   **에러 래핑**: 에러를 반환할 때는 `fmt.Errorf("context: %w", err)`를 사용하여 컨텍스트를 유지합니다.
-   **패닉 지양**: `panic`은 복구 불가능한 치명적인 오류 외에는 사용하지 않으며, 우아한 에러 처리를 우선합니다.
-   **로깅**: 구조화된 로깅을 사용하여 디버깅과 모니터링을 용이하게 합니다 (Wails의 `runtime.Log` 활용 권장).

### 2.3 동시성 (Concurrency)
-   **Goroutine 관리**: Goroutine 누수를 방지하기 위해 `context.Context`나 `quit` 채널을 활용하여 생명주기를 관리합니다.
-   **데이터 레이스 방지**: 공유 자원 접근 시에는 `sync.Mutex`나 `channels`를 사용하여 동기화합니다.

### 2.4 SQLite & 데이터베이스
-   **Prepared Statements**: SQL 인젝션 방지와 성능 향상을 위해 항상 파라미터화된 쿼리 또는 Prepared Statement를 사용합니다.
-   **트랜잭션**: 데이터 일관성이 필요한 쓰기 작업은 반드시 트랜잭션(`tx`) 내에서 수행합니다.
-   **리소스 해제**: `rows.Close()` 등 데이터베이스 리소스는 `defer`를 사용하여 확실하게 해제합니다.
-   **배치 처리**: 대량의 데이터 삽입 시에는 트랜잭션을 묶거나 배치 처리를 통해 성능을 최적화합니다.

---

## 3. Wails (Bridge) 개발 가이드라인

### 3.1 메서드 노출
-   **App 구조체**: 프론트엔드에 노출할 메서드는 `App` 구조체의 메서드로 정의합니다.
-   **데이터 타입**: 프론트엔드와 주고받는 데이터는 Go의 `struct`에 JSON 태그(`json:"fieldName"`)를 명시하여 직렬화 문제를 방지합니다.
-   **Context**: `App`의 메서드에서 `context`가 필요한 경우 `startup` 시 저장해둔 `ctx`를 사용합니다.

### 3.2 이벤트 기반 통신
-   **단방향 데이터 흐름**: 백엔드에서 실시간 데이터(예: 티커, 로그) 전송 시 `runtime.EventsEmit`을 사용하고, 프론트엔드에서 `EventsOn`으로 수신합니다.

---

## 4. React (Frontend) 개발 가이드라인

### 4.1 컴포넌트 설계 및 구조화
-   **컴포넌트 이름**: 항상 `PascalCase` 사용.
-   **단일 책임 원칙 (SRP)**: 컴포넌트는 하나의 역할만 수행하도록 분리.
-   **함수형 컴포넌트 & Hook**: 클래스형 대신 함수형 컴포넌트 권장.
-   **파일 구조**: 예: `src/components/MyComponent.tsx`, `src/features/Trading/TradingChart.tsx`.

### 4.2 상태 관리 (Zustand)
-   **Store 분리**: 도메인별로 Store를 분리하여 관리합니다 (예: `useTickerStore`, `useUserStore`).
-   **불변성**: 상태 업데이트 시 불변성을 유지합니다.
-   **선택적 구독**: 렌더링 최적화를 위해 필요한 상태만 선택하여 구독(`useStore(state => state.value)`)합니다.

### 4.3 성능 최적화
-   **메모이제이션**: 불필요한 리렌더링 방지를 위해 `React.memo`, `useMemo`, `useCallback`을 적절히 사용합니다.
-   **가상화**: 대량의 데이터(예: 로그 리스트, 호가창) 렌더링 시 `react-window` 등의 가상화 라이브러리 사용을 고려합니다.

---

## 5. 코드 품질 및 협업

### 5.1 Linting & Formatting
-   **Go**: `go vet`, `staticcheck` 사용. `gofmt`로 포맷팅.
-   **JS/TS**: `ESLint`, `Prettier` 규칙 준수.

### 5.2 버전 관리
-   **커밋 메시지**: 명확하고 간결하게 작성하며, 변경 사항의 '이유'를 포함합니다.
-   **브랜치 전략**: 기능 개발은 별도 브랜치에서 진행 후 병합합니다.

---

이 문서는 프로젝트 진행 상황에 따라 지속적으로 업데이트됩니다.
