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

                        {/* Best Result Summary */}
                        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
                            <div className="bg-gradient-to-br from-orange-600/20 to-orange-800/10 border border-orange-500/30 rounded-xl p-4">
                                <div className="text-orange-400 text-xs font-bold uppercase mb-1">🏆 Best Interval</div>
                                <div className="text-2xl font-bold text-white">
                                    {Math.floor(optimizationResults[0].interval_duration / 1e9 / 60)}분
                                </div>
                            </div>
                            <div className="bg-white/5 border border-white/10 rounded-xl p-4">
                                <div className="text-neutral-400 text-xs font-bold uppercase mb-1">💰 Net Profit</div>
                                <div className={`text-2xl font-bold ${optimizationResults[0].profit >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                                    {optimizationResults[0].profit > 0 ? '+' : ''}{optimizationResults[0].profit.toLocaleString()}
                                </div>
                            </div>
                            <div className="bg-white/5 border border-white/10 rounded-xl p-4">
                                <div className="text-neutral-400 text-xs font-bold uppercase mb-1">🔄 Total Cycles</div>
                                <div className="text-2xl font-bold text-white">
                                    {optimizationResults[0].cycle_count}회
                                </div>
                            </div>
                            <div className="bg-white/5 border border-white/10 rounded-xl p-4">
                                <div className="text-neutral-400 text-xs font-bold uppercase mb-1">📈 Win Rate</div>
                                <div className={`text-2xl font-bold ${optimizationResults[0].win_rate >= 0.5 ? 'text-green-400' : 'text-red-400'}`}>
                                    {(optimizationResults[0].win_rate * 100).toFixed(1)}%
                                </div>
                            </div>
                        </div>

                        {/* Cycle Statistics */}
                        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
                            <div className="bg-green-500/10 border border-green-500/20 rounded-xl p-4">
                                <div className="text-green-400 text-xs font-bold uppercase mb-1">✅ Win Cycles</div>
                                <div className="text-xl font-bold text-green-400">
                                    {optimizationResults[0].win_count}회 ({((optimizationResults[0].win_count / optimizationResults[0].cycle_count) * 100 || 0).toFixed(1)}%)
                                </div>
                            </div>
                            <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-4">
                                <div className="text-red-400 text-xs font-bold uppercase mb-1">❌ Loss Cycles</div>
                                <div className="text-xl font-bold text-red-400">
                                    {optimizationResults[0].loss_count}회 ({((optimizationResults[0].loss_count / optimizationResults[0].cycle_count) * 100 || 0).toFixed(1)}%)
                                </div>
                            </div>
                            <div className="bg-white/5 border border-white/10 rounded-xl p-4">
                                <div className="text-neutral-400 text-xs font-bold uppercase mb-1">📈 Avg Win</div>
                                <div className="text-xl font-bold text-green-400">
                                    +{optimizationResults[0].avg_win.toLocaleString()}
                                </div>
                            </div>
                            <div className="bg-white/5 border border-white/10 rounded-xl p-4">
                                <div className="text-neutral-400 text-xs font-bold uppercase mb-1">📉 Avg Loss</div>
                                <div className="text-xl font-bold text-red-400">
                                    {optimizationResults[0].avg_loss.toLocaleString()}
                                </div>
                            </div>
                        </div>

                        {/* Martingale vs Normal Strategy Comparison */}
                        <div className="bg-gradient-to-br from-purple-900/20 to-blue-900/20 border border-purple-500/30 rounded-xl p-6 mb-6">
                            <h4 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
                                <span className="text-purple-400">🎲</span> 전략 비교 (Normal vs Martingale)
                            </h4>
                            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                                {/* Normal Strategy */}
                                <div className="bg-white/5 rounded-xl p-4">
                                    <div className="text-neutral-400 text-sm font-bold uppercase mb-3">📊 일반 전략 (1배 고정)</div>
                                    <div className={`text-3xl font-bold mb-2 ${optimizationResults[0].profit >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                                        {optimizationResults[0].profit > 0 ? '+' : ''}{optimizationResults[0].profit.toLocaleString()} KRW
                                    </div>
                                    <div className="text-neutral-500 text-sm">
                                        매 사이클 동일 수량으로 거래
                                    </div>
                                </div>

                                {/* Martingale Strategy */}
                                <div className="bg-purple-500/10 border border-purple-500/20 rounded-xl p-4">
                                    <div className="text-purple-400 text-sm font-bold uppercase mb-3">🎰 마틴게일 전략 (손실시 2배)</div>
                                    <div className={`text-3xl font-bold mb-2 ${optimizationResults[0].martingale_profit >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                                        {optimizationResults[0].martingale_profit > 0 ? '+' : ''}{optimizationResults[0].martingale_profit.toLocaleString()} KRW
                                    </div>
                                    <div className="flex items-center gap-2 text-sm">
                                        <span className="text-neutral-500">최대 배율:</span>
                                        <span className={`font-bold ${optimizationResults[0].martingale_max_multiplier >= 8 ? 'text-red-400' : optimizationResults[0].martingale_max_multiplier >= 4 ? 'text-yellow-400' : 'text-green-400'}`}>
                                            {optimizationResults[0].martingale_max_multiplier}배
                                        </span>
                                        {optimizationResults[0].martingale_max_multiplier >= 8 && (
                                            <span className="text-red-400 text-xs">⚠️ 고위험</span>
                                        )}
                                    </div>
                                </div>
                            </div>

                            {/* Comparison Insight */}
                            <div className="mt-4 p-3 bg-black/20 rounded-lg">
                                <div className="text-sm">
                                    {optimizationResults[0].martingale_profit > optimizationResults[0].profit ? (
                                        <span className="text-purple-400">
                                            💡 마틴게일 전략이 <span className="font-bold text-green-400">+{(optimizationResults[0].martingale_profit - optimizationResults[0].profit).toLocaleString()}</span> 더 수익
                                            {optimizationResults[0].martingale_max_multiplier >= 8 && " (단, 고위험 주의)"}
                                        </span>
                                    ) : (
                                        <span className="text-neutral-400">
                                            💡 일반 전략이 <span className="font-bold text-green-400">+{(optimizationResults[0].profit - optimizationResults[0].martingale_profit).toLocaleString()}</span> 더 안정적
                                        </span>
                                    )}
                                </div>
                            </div>
                        </div>

                        {/* Results Table */}
                        <div className="bg-neutral-800/50 border border-white/5 rounded-xl overflow-hidden">
                            <table className="w-full text-left">
                                <thead className="bg-black/20 text-neutral-500 uppercase text-xs font-bold">
                                    <tr>
                                        <th className="px-4 py-4">Rank</th>
                                        <th className="px-4 py-4">Interval</th>
                                        <th className="px-4 py-4 text-right">Profit</th>
                                        <th className="px-4 py-4 text-right">Cycles</th>
                                        <th className="px-4 py-4 text-right">Win</th>
                                        <th className="px-4 py-4 text-right">Loss</th>
                                        <th className="px-4 py-4 text-right">Win Rate</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-white/5">
                                    {optimizationResults.map((res, i) => (
                                        <tr key={i} className={`hover:bg-white/5 transition-colors ${i === 0 ? 'bg-orange-500/5' : ''}`}>
                                            <td className="px-4 py-4 font-mono text-neutral-400">#{i + 1}</td>
                                            <td className="px-4 py-4 font-bold text-white">
                                                {res.interval_duration / 1e9 >= 60
                                                    ? `${Math.floor(res.interval_duration / 1e9 / 60)}분`
                                                    : `${res.interval_duration / 1e9}초`}
                                            </td>
                                            <td className={`px-4 py-4 text-right font-mono font-bold ${res.profit >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                                                {res.profit > 0 ? '+' : ''}{res.profit.toLocaleString()}
                                            </td>
                                            <td className="px-4 py-4 text-right text-neutral-400">{res.cycle_count}</td>
                                            <td className="px-4 py-4 text-right text-green-400">{res.win_count}</td>
                                            <td className="px-4 py-4 text-right text-red-400">{res.loss_count}</td>
                                            <td className={`px-4 py-4 text-right font-bold ${res.win_rate >= 0.5 ? 'text-green-400' : 'text-red-400'}`}>
                                                {(res.win_rate * 100).toFixed(1)}%
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>

                        {/* Apply Best Result Button */}
                        <div className="mt-4 flex justify-end">
                            <button
                                onClick={() => handleConfigChange(Math.floor(optimizationResults[0].interval_duration / 1e9))}
                                className="px-6 py-3 rounded-xl bg-gradient-to-r from-orange-600 to-amber-600 hover:from-orange-500 hover:to-amber-500 text-white font-bold shadow-lg shadow-orange-900/30 transition-all"
                            >
                                Apply Best Interval ({Math.floor(optimizationResults[0].interval_duration / 1e9 / 60)}분)
                            </button>
                        </div>
                    </div>
                )}
            </div>
        </div>
    )
}

export default App
