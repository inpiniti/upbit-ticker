package main

import (
	"context"
	"log"

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
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 1. Initialize SQLite
	// This will create 'upbit_ticker.db' in the same directory
	var err error
	a.db, err = gorm.Open(sqlite.Open("upbit_ticker.db"), &gorm.Config{})
	if err != nil {
		log.Printf("failed to connect database: %v", err)
		// Check if we can continue or should fail
	} else {
		// Auto Migrate
		a.db.AutoMigrate(&TickerModel{})
	}

	// 2. Start WebSocket
	a.wsClient = websocket.NewClient([]string{"KRW-BTC"})

	// Setup OnTick
	a.wsClient.OnTick(func(data types.Ticker) {
		// Save to DB
		if a.db != nil {
			model := TickerModel{
				Code:              data.Code,
				TradePrice:        data.TradePrice,
				Timestamp:         data.Timestamp,
				Change:            data.Change,
				SignedChangeRate:  data.SignedChangeRate,
				SignedChangePrice: data.SignedChangePrice,
			}
			// Use Create which is async-ish in nature if we don't wait for result?
			// Actually GORM is sync. This might block the websocket reader slightly.
			// For production, use a buffered channel for DB writes.
			// For this task, direct write is fine.
			a.db.Create(&model)
		}

		// Emit to Frontend
		runtime.EventsEmit(a.ctx, "tick", data)
	})

	if err := a.wsClient.Connect(); err != nil {
		log.Printf("WS Connect Error: %v", err)
		return
	}
	if err := a.wsClient.Subscribe(); err != nil {
		log.Printf("WS Subscribe Error: %v", err)
		return
	}
	// Start reading in a goroutine
	go a.wsClient.Start()
}

func (a *App) shutdown(ctx context.Context) {
	if a.wsClient != nil {
		a.wsClient.Stop()
	}
}

// TickerModel for Database
type TickerModel struct {
	gorm.Model
	Code              string
	TradePrice        float64
	Timestamp         int64
	Change            string
	SignedChangeRate  float64
	SignedChangePrice float64
}

// GetRecentTickers returns the last 50 tickers from DB
// Exposed to Wails frontend
func (a *App) GetRecentTickers() []types.Ticker {
	if a.db == nil {
		return []types.Ticker{}
	}

	var models []TickerModel
	a.db.Order("timestamp desc").Limit(50).Find(&models)

	var tickers []types.Ticker
	for _, m := range models {
		tickers = append(tickers, types.Ticker{
			Code:              m.Code,
			TradePrice:        m.TradePrice,
			Timestamp:         m.Timestamp,
			Change:            m.Change,
			SignedChangeRate:  m.SignedChangeRate,
			SignedChangePrice: m.SignedChangePrice,
			// Fill other fields if needed, or modify types.Ticker to match better
		})
	}
	return tickers
}
