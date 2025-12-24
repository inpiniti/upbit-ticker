import { useEffect, useState } from 'react'
import { useTickerStore } from './store/useTickerStore'
import { useTrendStore } from './store/useTrendStore'

declare global {
    interface Window {
        runtime: {
            EventsOn: (eventName: string, callback: (data: any) => void) => void
        }
        go: {
            main: {
                App: {
                    UpdateConfig: (intervalSec: number) => Promise<void>
                    RunOptimizer: () => Promise<any[]>
                }
            }
        }
    }
}

function App() {
    const { tickers, addTicker } = useTickerStore()
    const {
        currentPrice, currentAverage, currentSlope, isHolding, lastSignal,
        optimizationResults, updateTick, setOptimizationResults
    } = useTrendStore()

    const [intervalSec, setIntervalSec] = useState(60)
    const [isOptimizing, setIsOptimizing] = useState(false)

    useEffect(() => {
        if (window.runtime && window.runtime.EventsOn) {
            // Raw Ticker Event
            window.runtime.EventsOn("tick", (data: any) => {
                addTicker(data)
            })

            // Processed Logic Event
            window.runtime.EventsOn("tick_processed", (data: any) => {
                updateTick(data)
            })
        }
    }, [])

    const handleConfigChange = async (sec: number) => {
        setIntervalSec(sec)
        if (window.go?.main?.App?.UpdateConfig) {
            await window.go.main.App.UpdateConfig(sec)
        }
    }

    const runOptimizer = async () => {
        if (window.go?.main?.App?.RunOptimizer) {
            setIsOptimizing(true)
            try {
                const results = await window.go.main.App.RunOptimizer()
                console.log("Opt Results:", results)
                setOptimizationResults(results)
            } finally {
                setIsOptimizing(false)
            }
        }
    }

    // Format helpers
    const fmtPrice = (p: number) => p ? p.toLocaleString() : '-'
    const fmtSlope = (s: number | null) => s !== null ? s.toFixed(2) : '-'

    return (
        <div className="min-h-screen bg-[#0f1115] text-white font-sans selection:bg-orange-500/30">
            {/* Background Gradients */}
            <div className="fixed top-0 left-0 w-full h-full overflow-hidden pointer-events-none z-0">
                <div className="absolute top-[-10%] right-[-5%] w-[500px] h-[500px] bg-purple-600/20 rounded-full blur-[120px]" />
                <div className="absolute bottom-[-10%] left-[-10%] w-[600px] h-[600px] bg-blue-600/10 rounded-full blur-[100px]" />
            </div>

            <div className="relative z-10 max-w-6xl mx-auto p-8">
                {/* Header */}
                <header className="flex justify-between items-end mb-12">
                    <div>
                        <h1 className="text-4xl font-black tracking-tight bg-gradient-to-r from-white via-neutral-200 to-neutral-500 bg-clip-text text-transparent mb-2">
                            QUANT<span className="text-orange-500">.AI</span>
                        </h1>
                        <p className="text-neutral-500 font-medium">Algorithmic Trading System</p>
                    </div>
                    <div className="flex items-center gap-4">
                        <div className={`px-3 py-1 rounded-full text-xs font-bold border ${isHolding ? 'border-green-500/50 bg-green-500/10 text-green-400' : 'border-neutral-700 bg-neutral-800 text-neutral-400'}`}>
                            {isHolding ? 'HOLDING LONG' : 'NO POSITION'}
                        </div>
                    </div>
                </header>

                {/* Main Dashboard Grid */}
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 mb-12">

                    {/* Active Stat Card */}
                    <div className="lg:col-span-2 grid grid-cols-2 gap-4">
                        <div className="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-6">
                            <div className="text-neutral-400 text-sm font-semibold uppercase tracking-wider mb-1">Current Price</div>
                            <div className="text-3xl font-mono font-bold text-white tracking-tight">{fmtPrice(currentPrice)} <span className="text-base text-neutral-500">KRW</span></div>
                        </div>
                        <div className="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-6">
                            <div className="text-neutral-400 text-sm font-semibold uppercase tracking-wider mb-1">Interval Average</div>
                            <div className="text-3xl font-mono font-bold text-blue-400 tracking-tight">{fmtPrice(currentAverage)}</div>
                        </div>
                        <div className="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-6">
                            <div className="text-neutral-400 text-sm font-semibold uppercase tracking-wider mb-1">Slope (Momentum)</div>
                            <div className={`text-3xl font-mono font-bold tracking-tight ${(currentSlope || 0) > 0 ? 'text-green-400' : 'text-red-400'}`}>
                                {fmtSlope(currentSlope)}
                            </div>
                        </div>
                        <div className="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-6 relative overflow-hidden group">
                            <div className="absolute inset-0 bg-gradient-to-br from-orange-500/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
                            <div className="text-neutral-400 text-sm font-semibold uppercase tracking-wider mb-1">Last Signal</div>
                            <div className="text-3xl font-black text-white tracking-tight">{lastSignal || "WAITING"}</div>
                        </div>
                    </div>

                    {/* Control Panel */}
                    <div className="bg-neutral-900/50 backdrop-blur border border-white/5 rounded-2xl p-6 flex flex-col gap-6">
                        <div>
                            <label className="block text-neutral-400 text-xs font-bold uppercase mb-3">Time Interval</label>
                            <div className="grid grid-cols-3 gap-2">
                                {[10, 60, 300, 1800, 3600].map(sec => (
                                    <button
                                        key={sec}
                                        onClick={() => handleConfigChange(sec)}
                                        className={`px-3 py-2 rounded-lg text-xs font-bold transition-all ${intervalSec === sec
                                                ? 'bg-orange-600 text-white shadow-lg shadow-orange-900/50'
                                                : 'bg-neutral-800 text-neutral-400 hover:bg-neutral-700'
                                            }`}
                                    >
                                        {sec < 60 ? `${sec}s` : `${sec / 60}m`}
                                    </button>
                                ))}
                            </div>
                        </div>

                        <div className="h-px bg-white/10" />

                        <div>
                            <label className="block text-neutral-400 text-xs font-bold uppercase mb-3">Backtest Optimizer</label>
                            <button
                                onClick={runOptimizer}
                                disabled={isOptimizing}
                                className="w-full py-4 rounded-xl bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-500 hover:to-cyan-500 text-white font-bold shadow-xl shadow-blue-900/30 disabled:opacity-50 disabled:cursor-not-allowed transition-all relative overflow-hidden"
                            >
                                {isOptimizing ? (
                                    <span className="flex items-center justify-center gap-2">
                                        <svg className="animate-spin h-4 w-4 text-white" viewBox="0 0 24 24">
                                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                        </svg>
                                        Finding Sweet Spot...
                                    </span>
                                ) : 'RUN AI OPTIMIZER'}
                            </button>
                        </div>
                    </div>
                </div>

                {/* Optimization Results */}
                {optimizationResults.length > 0 && (
                    <div className="mb-12 animate-fade-in-up">
                        <h3 className="text-xl font-bold text-white mb-6 flex items-center gap-2">
                            <span className="w-2 h-8 bg-orange-500 rounded-full" />
                            Optimization Results (Top 5)
                        </h3>
                        <div className="bg-neutral-800/50 border border-white/5 rounded-xl overflow-hidden">
                            <table className="w-full text-left">
                                <thead className="bg-black/20 text-neutral-500 uppercase text-xs font-bold">
                                    <tr>
                                        <th className="px-6 py-4">Rank</th>
                                        <th className="px-6 py-4">Interval</th>
                                        <th className="px-6 py-4 text-right">Profit</th>
                                        <th className="px-6 py-4 text-right">Trades</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-white/5">
                                    {optimizationResults.map((res, i) => (
                                        <tr key={i} className="hover:bg-white/5 transition-colors">
                                            <td className="px-6 py-4 font-mono text-neutral-400">#{i + 1}</td>
                                            <td className="px-6 py-4 font-bold text-white">{(res.interval_duration / 1e9).toFixed(0)}s</td>
                                            <td className={`px-6 py-4 text-right font-mono font-bold ${res.profit >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                                                {res.profit > 0 ? '+' : ''}{res.profit.toLocaleString()}
                                            </td>
                                            <td className="px-6 py-4 text-right text-neutral-400">{res.trade_count}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                )}
            </div>
        </div>
    )
}

export default App
