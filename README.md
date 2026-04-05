# Seckill System — Quick Start

## Requirements

- Docker & Docker Compose
- Go 1.25+
- Node.js 20+

## Start Backend Services

```bash
docker compose up --build -d
```

Wait ~15 seconds for Kafka to fully initialise before making requests.

## Start Dashboard

```bash
cd dashboard
npm install
npm run dev
```

Open [http://localhost:5173](http://localhost:5173)

## API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/activity` | Create a flash sale activity |
| GET | `/api/activity/:id` | Get activity and current stock |
| POST | `/api/seckill` | Purchase (atomic, rate-limited) |
| POST | `/api/order/:id/pay` | Pay a pending order |
| GET | `/api/order/:id` | Get order by ID |
| GET | `/api/orders/recent?limit=N` | Get recent orders |
| GET | `/api/stats` | Real-time counters |
| GET | `/api/stats/qps` | QPS for last 60 seconds |

## Testing

**1. Create an activity**

```bash
curl -X POST http://localhost:8080/api/activity \
  -H "Content-Type: application/json" \
  -d '{"name":"flash","stock_total":100,"start_time":"2026-01-01T00:00:00Z","end_time":"2027-01-01T00:00:00Z"}'
```

**2. Run load test**

```bash
go run ./scripts/loadtest -activity <activity_id> -users 10000
```

Expected: 100 succeed, 9,900 rejected. Unpaid orders are automatically cancelled after 30 seconds.

**3. Pay an order**

```bash
curl -X POST http://localhost:8080/api/order/<order_id>/pay
```

## Shut Down

```bash
docker compose down
```
