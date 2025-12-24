import { create } from 'zustand'

export interface Trade {
    ts: number
    price: number
    type: string
    execution_price: number
}

export interface OptimizationResult {
    interval_duration: number // NS
    profit: number
    trade_count: number
}

interface TrendState {
    currentPrice: number
    currentAverage: number
    currentSlope: number | null
    isHolding: boolean
    lastSignal: string

    trades: Trade[]
    optimizationResults: OptimizationResult[]

    updateTick: (data: any) => void
    addTrade: (signal: string, price: number, ts: number) => void // Simplified
    setOptimizationResults: (results: OptimizationResult[]) => void
}

export const useTrendStore = create<TrendState>((set) => ({
    currentPrice: 0,
    currentAverage: 0,
    currentSlope: null,
    isHolding: false,
    lastSignal: "",
    trades: [],
    optimizationResults: [],

    updateTick: (data: any) => set((state) => ({
        currentPrice: data.new_state.interval_buffer.length > 0 ? data.new_state.interval_buffer[data.new_state.interval_buffer.length - 1].price : state.currentPrice,
        // Wait, ProcessTick returns current_average only when closed?
        // Let's check logic.go. 
        // Logic.go returns CurrentAverage when closed. What about open?
        // Logic.go returns CurrentAverage ONLY when closed.
        // So we should update it only if interval_closed is true, OR use previous average.
        // But for "Realtime" chart, we might want real-time average?
        // Logic.go calculates average only on close.
        // Okay, we visualize LAST CLOSED average.
        currentAverage: data.interval_closed ? data.current_average : state.currentAverage,
        currentSlope: data.interval_closed ? data.current_slope : state.currentSlope,
        isHolding: data.new_state.is_holding,
        lastSignal: data.trade_signal || state.lastSignal
    })),

    addTrade: (signal, price, ts) => set((state) => ({
        trades: [...state.trades, { type: signal, price, ts, execution_price: price }]
    })),

    setOptimizationResults: (results) => set({ optimizationResults: results })
}))
