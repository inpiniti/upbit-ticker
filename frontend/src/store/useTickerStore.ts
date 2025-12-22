import { create } from 'zustand'

export interface Ticker {
    code: string
    trade_price: number
    signed_change_price: number
    signed_change_rate: number
    change: string
    timestamp: number
}

interface TickerState {
    tickers: Ticker[]
    addTicker: (ticker: Ticker) => void
    setTickers: (tickers: Ticker[]) => void
}

export const useTickerStore = create<TickerState>((set) => ({
    tickers: [],
    addTicker: (ticker) => set((state) => {
        // Keep only the latest 100 entries for the list
        const newTickers = [ticker, ...state.tickers];
        if (newTickers.length > 100) {
            return { tickers: newTickers.slice(0, 100) };
        }
        return { tickers: newTickers };
    }),
    setTickers: (tickers) => set({ tickers }),
}))
