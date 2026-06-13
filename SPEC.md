# Personal Investment Portfolio Manager

## 1. Overview

### Objective

Develop a modern web application that consolidates and tracks personal investments across multiple stock exchanges, initially:

* B3 (Brazilian Stock Exchange)
* NASDAQ (United States)
* Crypto (BTC, ETH, SOL, ...)

The application must provide portfolio management, profit/loss calculations, average cost tracking, performance analysis, and historical investment records.

The system should support manual transaction registration while automatically fetching real-time market data from public APIs.

---

## 2. Goals

### Functional Goals

* Track all stock purchases and sales.
* Calculate average cost per asset.
* Calculate realized and unrealized profit/loss.
* Consolidate holdings from multiple markets.
* Display historical portfolio evolution.
* Generate performance and allocation charts.
* Allow CRUD operations on all transactions.
* Automatically fetch stock prices.
* Support multiple currencies (BRL and USD).
* Calculate total net worth.

### Non-Functional Goals

* Responsive modern UI.
* Containerized deployment.
* Simple local installation.
* Extensible architecture.
* Secure authentication.
* Fast dashboard loading.

---

# 3. High-Level Architecture

## Frontend

### Technology

* React
* TypeScript
* Vite
* TailwindCSS
* shadcn/ui
* TanStack Query
* Recharts

### Responsibilities

* Dashboard
* Portfolio views
* Charts
* Transaction management
* Authentication
* Settings

---

## Backend

### Technology

* Go (Gin/Fiber)
* REST API
* JWT Authentication

### Responsibilities

* Portfolio calculations
* Asset management
* Transaction processing
* Market data aggregation
* Currency conversion
* User management

---

## Database

### Technology

PostgreSQL

Reason:

* Reliable
* Excellent support for financial data
* Docker friendly
* Strong indexing

---

## Infrastructure

### Containers

Docker Compose

Services:

* frontend
* backend
* postgres
* redis (optional caching)
* market-data-worker

---

# 4. System Components

## Frontend

### Pages

#### Dashboard

Shows:

* Total portfolio value
* Daily gain/loss
* Total gain/loss
* Portfolio allocation
* Best performers
* Worst performers
* Portfolio evolution chart

---

#### Portfolio

Displays:

* Asset
* Exchange
* Quantity
* Average cost
* Current price
* Current value
* Gain/Loss
* Gain/Loss %

---

#### Asset Details

Displays:

* Asset information
* Historical purchases
* Historical sales
* Average cost evolution
* Asset performance chart

---

#### Transactions

Displays:

* Purchase history
* Sale history

Actions:

* Create transaction
* Edit transaction
* Delete transaction

---

#### Reports

Displays:

* Monthly returns
* Annual returns
* Dividend history
* Portfolio growth

---

#### Settings

Displays:

* Preferred currency
* Theme
* API provider settings

---

# 5. Authentication

## Features

* Registration
* Login
* Password reset
* JWT authentication

## Roles

### User

Can:

* Manage own portfolio
* Manage transactions
* View reports

---

# 6. Market Data Integration

## B3 Assets

Possible providers:

### BRAPI

https://brapi.dev

Provides:

* Current price
* Historical prices
* Company information

---

## NASDAQ Assets

Possible providers:

### Alpha Vantage

https://www.alphavantage.co

### Finnhub

https://finnhub.io

### Twelve Data

https://twelvedata.com

---
## Cryptocurrency Assets

Possible providers:
### CoinGecko (Recommended)

Website:
https://www.coingecko.com

API:
https://www.coingecko.com/en/api

Supported assets:

- Bitcoin (BTC)
- Ethereum (ETH)
- Solana (SOL)
- Stablecoins
- Any cryptocurrency supported by the provider

Use cases:

- Current crypto prices
- Historical charts
- Market data
- Asset metadata

---

## Currency Exchange

Required:

USD ↔ BRL conversion

Possible providers:

* ExchangeRate API
* Frankfurter API

---

# 7. Core Features

## Asset Tracking

Supported asset types:

### Initial Version

* Stocks
* REITs
* ETFs
* Crypto

### Future

* Mutual Funds
* Bonds

---

## Transaction Types

### BUY

Fields:

* Asset
* Quantity
* Price
* Fees
* Currency
* Date

---

### SELL

Fields:

* Asset
* Quantity
* Price
* Fees
* Currency
* Date

---

### DIVIDEND

Fields:

* Asset
* Amount
* Currency
* Date

---

# 8. Portfolio Calculations

## Average Cost

Formula:

Average Cost = Total Invested / Total Shares

---

## Unrealized P/L

Current Value - Total Cost

---

## Realized P/L

Calculated when shares are sold.

---

## Portfolio Value

Sum of all open positions.

---

## Daily Performance

Current Value - Previous Market Close

---

## Currency Conversion

Store:

* Original currency
* Converted currency

Always preserve original transaction values.

---

# 9. Dashboard Widgets

## Portfolio Summary Card

Displays:

* Total invested
* Current value
* Total profit/loss

---

## Allocation Chart

By:

* Sector
* Asset
* Country
* Currency

---

## Portfolio Evolution Chart

Time ranges:

* 1M
* 3M
* 6M
* 1Y
* 5Y
* ALL

---

## Top Movers

Displays:

* Top gainers
* Top losers

---

## Income Card

Displays:

* Dividends received
* Monthly income

---

# 10. Database Design

## users

| Column        | Type      |
| ------------- | --------- |
| id            | UUID      |
| email         | VARCHAR   |
| password_hash | VARCHAR   |
| created_at    | TIMESTAMP |

---

## assets

| Column     | Type      |
| ---------- | --------- |
| id         | UUID      |
| symbol     | VARCHAR   |
| name       | VARCHAR   |
| exchange   | VARCHAR   |
| currency   | VARCHAR   |
| sector     | VARCHAR   |
| created_at | TIMESTAMP |

---

## transactions

| Column           | Type      |
| ---------------- | --------- |
| id               | UUID      |
| user_id          | UUID      |
| asset_id         | UUID      |
| type             | VARCHAR   |
| quantity         | DECIMAL   |
| price            | DECIMAL   |
| fees             | DECIMAL   |
| currency         | VARCHAR   |
| transaction_date | TIMESTAMP |
| notes            | TEXT      |
| created_at       | TIMESTAMP |

---

## asset_prices

| Column    | Type      |
| --------- | --------- |
| id        | UUID      |
| asset_id  | UUID      |
| price     | DECIMAL   |
| currency  | VARCHAR   |
| timestamp | TIMESTAMP |

---

## dividends

| Column       | Type      |
| ------------ | --------- |
| id           | UUID      |
| user_id      | UUID      |
| asset_id     | UUID      |
| amount       | DECIMAL   |
| currency     | VARCHAR   |
| payment_date | TIMESTAMP |

---

# 11. REST API

## Authentication

POST /api/auth/register

POST /api/auth/login

POST /api/auth/logout

---

## Assets

GET /api/assets

GET /api/assets/:id

---

## Portfolio

GET /api/portfolio

GET /api/portfolio/summary

GET /api/portfolio/performance

---

## Transactions

GET /api/transactions

POST /api/transactions

PUT /api/transactions/:id

DELETE /api/transactions/:id

---

## Dividends

GET /api/dividends

POST /api/dividends

PUT /api/dividends/:id

DELETE /api/dividends/:id

---

# 12. Background Jobs

## Price Synchronization

Runs:

* Every 1 minute during market hours
* Every 15 minutes outside market hours

Updates:

* Latest asset prices
* Exchange rates

---

## Portfolio Recalculation

Triggered:

* New transaction
* Edit transaction
* Delete transaction

Updates:

* Average cost
* Holdings
* Profit/Loss

---

# 13. UI/UX Vision

Design inspiration:

* TradingView
* Yahoo Finance
* Wealthfront
* Robinhood
* Kinvo

Style:

* Dark mode by default
* Glassmorphism cards
* Smooth animations
* Responsive layout
* Professional finance dashboard

Features:

* Asset search
* Keyboard shortcuts
* Quick transaction entry
* Interactive charts

---

# 14. Docker Compose Structure

```yaml
services:
  frontend:
    build: ./frontend

  backend:
    build: ./backend

  postgres:
    image: postgres:17

  redis:
    image: redis:8

  market-data-worker:
    build: ./worker
```

---

# 15. Future Enhancements

* Brokerage import (CSV)
* Dividend forecasting
* Tax reports
* AI portfolio insights
* Portfolio rebalancing suggestions
* Multi-user support
* Mobile app
* Push notifications
* Stock alerts
* Watchlists
* Goal tracking
* Retirement planning

---

# 16. MVP Scope

Version 1.0 includes:

* Authentication
* Stock management
* Buy/Sell CRUD
* Portfolio dashboard
* Average cost calculation
* Profit/Loss calculation
* B3 integration
* NASDAQ integration
* USD/BRL conversion
* Portfolio charts
* Docker deployment
* PostgreSQL database

Success criteria:

A user can register all stock purchases and sales, visualize current holdings, track profitability, and monitor portfolio evolution across both B3 and NASDAQ from a single dashboard.

