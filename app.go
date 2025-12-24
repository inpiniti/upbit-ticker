package main

import (
	"context"
	"log"
	"time"

	"upbit-ticker/internal/analysis"
	"upbit-ticker/types"
	"upbit-ticker/websocket"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// App struct
type App struct {
	ctx      context.Context
	db       *gorm.DB
	wsClient *websocket.Client

	// Trading State
	marketState types.MarketState
	botConfig   types.BotConfiguration

	// Batcher
	tickChan chan types.RawTick
}

// Models for SQLite
type RawTickModel struct {
	Timestamp int64 `gorm:"primaryKey"` // Millisecond
	Price     float64
}

type TradeModel struct {
	gorm.Model
	Timestamp      int64
	Price          float64
	Type           string // BUY, SELL
	ExecutionPrice float64
	Profit         float64 // For SELL only
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		tickChan: make(chan types.RawTick, 1000),
		marketState: types.MarketState{
			IntervalBuffer: make([]types.RawTick, 0),
		},
		botConfig: types.BotConfiguration{
			IntervalDuration: 60 * time.Second, // Default 1 min
			SlippageRate:     0.0002,
			FeeRate:          0.0005,
		},
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 1. Initialize SQLite
	var err error
	a.db, err = gorm.Open(sqlite.Open("upbit_ticker.db"), &gorm.Config{})
	if err != nil {
		log.Printf("failed to connect database: %v", err)
	} else {
		// Auto Migrate
		a.db.AutoMigrate(&RawTickModel{}, &TradeModel{})
	}

	// 2. Start Tick Batcher
	go a.runTickBatcher()

	// 3. Start WebSocket
	a.wsClient = websocket.NewClient([]string{"KRW-BTC"})

	// Setup OnTick
	a.wsClient.OnTick(func(data types.Ticker) {
		rawTick := types.RawTick{
			Timestamp: data.Timestamp, // Milliseconds
			Price:     data.TradePrice,
		}

		// 1. Enqueue for DB Save
		select {
		case a.tickChan <- rawTick:
		default:
			// Buffer full
		}

		// 2. Process Logic (Pure)
		result := analysis.ProcessTick(a.marketState, rawTick, a.botConfig)
		a.marketState = result.NewState

		// 3. Emit Process Result to Frontend
		runtime.EventsEmit(a.ctx, "tick_processed", result)

		// 4. Handle Trade Signal
		if result.TradeSignal != "" {
			executionPrice := analysis.ApplyCost(result.TradeSignal, rawTick.Price, a.botConfig)
			log.Printf("TRADE SIGNAL: %s @ %f", result.TradeSignal, executionPrice)

			// Save Trade
			a.db.Create(&TradeModel{
				Timestamp:      rawTick.Timestamp,
				Price:          rawTick.Price,
				Type:           result.TradeSignal,
				ExecutionPrice: executionPrice,
			})

			// Emit Trade Event
			runtime.EventsEmit(a.ctx, "trade_event", result.TradeSignal)
		}
	})

	if err := a.wsClient.Connect(); err != nil {
		log.Printf("WS Connect Error: %v", err)
		return
	}
	if err := a.wsClient.Subscribe(); err != nil {
		log.Printf("WS Subscribe Error: %v", err)
		return
	}
	go a.wsClient.Start()
}

func (a *App) runTickBatcher() {
	var buffer []RawTickModel
	ticker := time.NewTicker(2 * time.Second) // Batch every 2 sec
	defer ticker.Stop()

	for {
		select {
		case tick := <-a.tickChan:
			buffer = append(buffer, RawTickModel{
				Timestamp: tick.Timestamp,
				Price:     tick.Price,
			})
			if len(buffer) >= 100 {
				a.flushTicks(buffer)
				buffer = nil
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				a.flushTicks(buffer)
				buffer = nil
			}
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *App) flushTicks(ticks []RawTickModel) {
	if a.db == nil || len(ticks) == 0 {
		return
	}
	// Batch Insert
	if err := a.db.CreateInBatches(ticks, 100).Error; err != nil {
		log.Printf("Failed to batch insert ticks: %v", err)
	}
}

func (a *App) shutdown(ctx context.Context) {
	if a.wsClient != nil {
		a.wsClient.Stop()
	}
}

// --- Exposed Methods ---

// UpdateConfig updates the bot configuration dynamically
func (a *App) UpdateConfig(intervalSeconds int) {
	a.botConfig.IntervalDuration = time.Duration(intervalSeconds) * time.Second
	// Reset State
	a.marketState = types.MarketState{
		IntervalBuffer:    make([]types.RawTick, 0),
		IntervalStartTime: 0,
		PrevAverage:       nil,
		PrevSlope:         nil,
		IsHolding:         false,
	}
	log.Printf("Config Updated: Interval %v", a.botConfig.IntervalDuration)
}

// RunOptimizer runs the backtest optimization
func (a *App) RunOptimizer() []types.OptimizationResult {
	log.Println("Starting Optimization...")

	// 1. Load All Ticks
	var models []RawTickModel
	if err := a.db.Order("timestamp asc").Find(&models).Error; err != nil {
		log.Printf("Error loading ticks: %v", err)
		return []types.OptimizationResult{}
	}

	// 2. Convert to Domain Type
	rawTicks := make([]types.RawTick, len(models))
	for i, m := range models {
		rawTicks[i] = types.RawTick{
			Timestamp: m.Timestamp,
			Price:     m.Price,
		}
	}

	// 3. Run Logic
	results := analysis.FindSweetSpot(rawTicks, a.botConfig)

	return results
}

// GetTradeHistory returns all trades
func (a *App) GetTradeHistory() []TradeModel {
	var trades []TradeModel
	a.db.Order("timestamp desc").Find(&trades)
	return trades
}
