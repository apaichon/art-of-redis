Backend Folder structure
```sh
sales-analytics/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers.go
│   │   └── server.go
│   ├── models/
│   │   └── types.go
│   └── storage/
│       ├── redis.go
│       └── analytics.go
└── go.mod
```

Frontend Folder structure
```sh
web/
src/
├── components/
│   ├── analytics/
│   │   ├── TicketAnalytics.jsx
│   │   ├── SummaryCards.jsx
│   │   ├── HourlySalesChart.jsx
│   │   ├── CategoryDistribution.jsx
│   │   └── TopConcerts.jsx
│   └── ui/
│       └── LoadingSpinner.jsx
├── hooks/
│   └── useWebSocket.js
├── utils/
│   └── formatters.js
└── constants/
    └── chartConfig.js
```