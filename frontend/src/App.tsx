import { useEffect } from 'react'
import { useTickerStore } from './store/useTickerStore'

// We need to define the type for the custom event if we want to be type-safe without the Wails generated files yet.
// In a real setup, you would import { EventsOn } from '../wailsjs/runtime'

declare global {
    interface Window {
        runtime: {
            EventsOn: (eventName: string, callback: (data: any) => void) => void
        }
    }
}

function App() {
    const { tickers, addTicker } = useTickerStore()

    useEffect(() => {
        // Connect to Wails events
        // When the backend sends "tick", we update the store
        if (window.runtime && window.runtime.EventsOn) {
            window.runtime.EventsOn("tick", (data: any) => {
                console.log("Tick received:", data);
                addTicker(data);
            });
        } else {
            console.log("Wails runtime not found. Are you running in a browser?");
        }
    }, [addTicker])

    return (
        <div className="min-h-screen bg-neutral-900 text-white p-8 font-sans">
            <div className="max-w-4xl mx-auto">
                <header className="flex justify-between items-center mb-8">
                    <h1 className="text-3xl font-bold bg-gradient-to-r from-yellow-400 to-orange-500 bg-clip-text text-transparent">
                        Upbit Realtime Ticker
                    </h1>
                    <div className="text-sm text-neutral-500">
                        Powered by Wails + React + SQLite
                    </div>
                </header>

                <div className="bg-neutral-800/50 backdrop-blur border border-neutral-700 rounded-xl overflow-hidden shadow-2xl">
                    <div className="overflow-x-auto">
                        <table className="w-full text-left text-sm">
                            <thead className="bg-neutral-800 text-neutral-400 uppercase tracking-wider text-xs font-semibold">
                                <tr>
                                    <th className="px-6 py-4">Time</th>
                                    <th className="px-6 py-4">Code</th>
                                    <th className="px-6 py-4 text-right">Price (KRW)</th>
                                    <th className="px-6 py-4 text-right">Change</th>
                                    <th className="px-6 py-4 text-right">Amount</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-neutral-700">
                                {tickers.map((t, i) => (
                                    <tr key={i} className="hover:bg-white/5 transition-colors duration-150 group">
                                        <td className="px-6 py-4 text-neutral-400 font-mono text-xs">
                                            {new Date(t.timestamp).toLocaleTimeString()}
                                        </td>
                                        <td className="px-6 py-4 font-bold text-white group-hover:text-yellow-400 transition-colors">
                                            {t.code}
                                        </td>
                                        <td className="px-6 py-4 text-right font-mono text-neutral-200">
                                            {t.trade_price.toLocaleString()}
                                        </td>
                                        <td className={`px-6 py-4 text-right font-medium ${t.change === 'RISE' ? 'text-red-400' :
                                                t.change === 'FALL' ? 'text-blue-400' : 'text-neutral-400'
                                            }`}>
                                            <span className="inline-block w-16 px-2 py-1 rounded bg-opacity-10 bg-current text-xs">
                                                {t.signed_change_rate > 0 ? '+' : ''}
                                                {(t.signed_change_rate * 100).toFixed(2)}%
                                            </span>

                                        </td>
                                        <td className="px-6 py-4 text-right text-neutral-500 text-xs">
                                            {t.signed_change_price.toLocaleString()}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                    {tickers.length === 0 && (
                        <div className="p-12 text-center text-neutral-500 flex flex-col items-center">
                            <div className="animate-pulse mb-4 text-yellow-500">Waiting for market data...</div>
                            <div className="text-xs text-neutral-600">Check if your Upbit connection is active</div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}

export default App
