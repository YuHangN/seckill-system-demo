import { useState, useEffect, useCallback } from 'react'
import {
  LineChart, Line, BarChart, Bar,
  XAxis, YAxis, CartesianGrid, Tooltip,
  ResponsiveContainer, Cell
} from 'recharts'

const API = '/api'
const POLL_MS = 500

const CARDS = [
  { key: 'total_requests', label: 'Total Requests', accent: 'var(--cyan)',   unit: 'requests' },
  { key: 'success_count',  label: 'Succeeded',      accent: 'var(--green)',  unit: 'orders'   },
  { key: 'fail_count',     label: 'Failed',          accent: 'var(--red)',    unit: 'rejected' },
  { key: 'cancel_count',   label: 'Cancelled',       accent: 'var(--amber)',  unit: 'timeout'  },
  { key: 'paid_count',     label: 'Paid',            accent: 'var(--purple)', unit: 'settled'  },
]

const BAR_COLORS = {
  success_count: 'var(--green)',
  fail_count:    'var(--red)',
  cancel_count:  'var(--amber)',
}

function CustomQPSTooltip({ active, payload, label }) {
  if (!active || !payload?.length) return null
  return (
    <div className="custom-tooltip">
      <div className="label">T−{60 - label}s</div>
      <div style={{ color: 'var(--cyan)' }}>{payload[0].value} req/s</div>
    </div>
  )
}

function CustomBarTooltip({ active, payload, label }) {
  if (!active || !payload?.length) return null
  return (
    <div className="custom-tooltip">
      <div className="label">{label}</div>
      <div style={{ color: payload[0].fill }}>{payload[0].value.toLocaleString()}</div>
    </div>
  )
}

export default function App() {
  const [stats, setStats]   = useState({ total_requests: 0, success_count: 0, fail_count: 0, cancel_count: 0, paid_count: 0 })
  const [qps, setQps]       = useState([])
  const [clock, setClock]   = useState(new Date())
  const [error, setError]   = useState(false)

  const fetchData = useCallback(async () => {
    try {
      const [statsRes, qpsRes] = await Promise.all([
        fetch(`${API}/stats`),
        fetch(`${API}/stats/qps`),
      ])
      if (!statsRes.ok || !qpsRes.ok) throw new Error('non-200')
      const statsData = await statsRes.json()
      const qpsData   = await qpsRes.json()

      setStats(statsData)
      // Normalize: use index as x-axis position (0 = oldest, 59 = newest)
      setQps((qpsData || []).map((p, i) => ({ idx: i, count: p.count ?? 0 })))
      setError(false)
    } catch {
      setError(true)
    }
  }, [])

  useEffect(() => {
    fetchData()
    const id = setInterval(fetchData, POLL_MS)
    return () => clearInterval(id)
  }, [fetchData])

  useEffect(() => {
    const id = setInterval(() => setClock(new Date()), 1000)
    return () => clearInterval(id)
  }, [])

  const barData = [
    { name: 'Success',   value: stats.success_count, key: 'success_count' },
    { name: 'Failed',    value: stats.fail_count,    key: 'fail_count'    },
    { name: 'Cancelled', value: stats.cancel_count,  key: 'cancel_count'  },
  ]

  const currentQPS = qps.length > 0 ? qps[qps.length - 1].count : 0

  return (
    <div className="dashboard">
      {/* Header */}
      <header className="header">
        <div className="header-left">
          <h1>SECKILL OPS</h1>
          <p>Flash Sale · Real-time Operations Monitor</p>
        </div>
        <div className="header-right">
          <div className="live-badge">
            <span className="live-dot" />
            {error ? 'OFFLINE' : 'LIVE'}
          </div>
          <div className="timestamp">
            {clock.toLocaleTimeString('en-GB', { hour12: false })}
          </div>
        </div>
      </header>

      {/* Stat Cards */}
      <div className="stat-grid">
        {CARDS.map(({ key, label, accent, unit }) => (
          <div
            key={key}
            className="stat-card"
            style={{ '--accent': accent }}
          >
            <div className="stat-label">{label}</div>
            <div className="stat-value">
              {(stats[key] ?? 0).toLocaleString()}
            </div>
            <div className="stat-unit">{unit}</div>
          </div>
        ))}
      </div>

      {/* Charts */}
      <div className="chart-grid">
        {/* QPS Line Chart */}
        <div className="chart-card">
          <div className="chart-header">
            <span className="chart-title">REQUESTS / SECOND</span>
            <span className="chart-subtitle">
              NOW&nbsp;&nbsp;
              <span style={{ color: 'var(--cyan)', fontWeight: 500 }}>
                {currentQPS}
              </span>
              &nbsp;req/s
            </span>
          </div>
          {qps.length === 0 ? (
            <div className="no-data">awaiting data…</div>
          ) : (
            <ResponsiveContainer width="100%" height={220}>
              <LineChart data={qps} margin={{ top: 4, right: 4, left: -24, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis
                  dataKey="idx"
                  tick={false}
                  axisLine={{ stroke: 'var(--bg-border)' }}
                  tickLine={false}
                />
                <YAxis
                  tick={{ fontFamily: 'DM Mono', fontSize: 10, fill: 'var(--text-dim)' }}
                  axisLine={false}
                  tickLine={false}
                  width={40}
                />
                <Tooltip content={<CustomQPSTooltip />} />
                <Line
                  type="monotone"
                  dataKey="count"
                  stroke="var(--cyan)"
                  strokeWidth={2}
                  dot={false}
                  isAnimationActive={false}
                />
              </LineChart>
            </ResponsiveContainer>
          )}
        </div>

        {/* Bar Chart */}
        <div className="chart-card">
          <div className="chart-header">
            <span className="chart-title">ORDER BREAKDOWN</span>
            <span className="chart-subtitle">cumulative</span>
          </div>
          <ResponsiveContainer width="100%" height={220}>
            <BarChart data={barData} margin={{ top: 4, right: 4, left: -24, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" vertical={false} />
              <XAxis
                dataKey="name"
                tick={{ fontFamily: 'DM Mono', fontSize: 10, fill: 'var(--text-dim)' }}
                axisLine={false}
                tickLine={false}
              />
              <YAxis
                tick={{ fontFamily: 'DM Mono', fontSize: 10, fill: 'var(--text-dim)' }}
                axisLine={false}
                tickLine={false}
                width={40}
              />
              <Tooltip content={<CustomBarTooltip />} cursor={{ fill: 'rgba(255,255,255,0.03)' }} />
              <Bar dataKey="value" radius={[2, 2, 0, 0]} isAnimationActive={false}>
                {barData.map((entry) => (
                  <Cell key={entry.key} fill={BAR_COLORS[entry.key]} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  )
}
