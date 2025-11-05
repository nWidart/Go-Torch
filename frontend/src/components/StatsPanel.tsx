import React, { useMemo } from 'react'
import { card, h2, th, td } from '../uiStyles'
import { fmtDur } from '../lib/format'
import { UIState } from '../types/ui'
import Stat from './Stat'

export type StatsPanelProps = {
  state: UIState | null
}

const fmtMoney = (n: number | undefined | null, digits = 2) => {
  if (n === undefined || n === null || !isFinite(n as number)) return '—'
  return (n as number).toFixed(digits)
}

export default function StatsPanel({ state }: StatsPanelProps) {
  const running = !!(state && state.sessionStart && !state.sessionEnd)
  const status = state?.sessionPaused ? 'Paused' : (state?.inMap ? 'In Map' : (running ? 'Running' : 'Idle'))

  const sessionDur = useMemo(() => {
    const start = state?.sessionStart || 0
    if (!start) return '0s'
    const pausedAccum = state?.pausedAccumMs || 0
    const end = state?.sessionEnd ? state.sessionEnd : (state?.sessionPaused ? state.pausedAt : undefined)
    // Use start adjusted by past paused time; when paused, end=pausedAt so current pause not included
    const startAdj = start + pausedAccum
    return fmtDur(startAdj, end)
  }, [state])

  const mapDur = useMemo(() => {
    if (!state) return '—'
    if (state.inMap && state.maps?.length) {
      const last = state.maps[state.maps.length - 1]
      if (last?.durationMs) {
        const now = Date.now()
        return fmtDur(now - last.durationMs, now)
      }
    }
    if (state.mapStart) return fmtDur(state.mapStart, state.mapEnd || undefined)
    return '—'
  }, [state])

  const eph = useMemo(() => fmtMoney(state?.earningsPerHour, 1), [state])
  const eps = useMemo(() => fmtMoney(state?.earningsPerSession, 2), [state])
  const avgMapDur = useMemo(() => {
    const ms = state?.avgMapTimeMs || 0
    if (!ms) return '—'
    const now = Date.now()
    return fmtDur(now - ms, now)
  }, [state])

  // Show last 5 maps earnings (including current as last)
  const maps = state?.maps || []
  const lastMaps = maps.slice(Math.max(0, maps.length - 5))

  return (
    <section style={card}>
      <h2 style={h2}>Live Stats</h2>
      <div style={{ display: 'flex', gap: 24, flexWrap: 'wrap' }}>
        <Stat label="Status" value={status} />
        <Stat label="Session Duration" value={sessionDur} />
        <Stat label="Map Duration" value={mapDur} />
        <Stat label="Earnings/hour" value={eph} />
        <Stat label="Earnings/session" value={eps} />
        <Stat label="Avg time/map" value={avgMapDur} />
      </div>

      <div style={{ marginTop: 12 }}>
        <div style={{ opacity: 0.7, marginBottom: 4 }}>Earnings per map</div>
        {lastMaps.length ? (
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr>
                <th style={th}>Map #</th>
                <th style={th}>Earnings</th>
                <th style={th}>Duration</th>
              </tr>
            </thead>
            <tbody>
              {lastMaps.reverse().map((m, idx) => {
                const dur = m.durationMs > 0 ? fmtDur(Date.now() - m.durationMs, Date.now()) : '—'
                const i = maps.length - lastMaps.length + idx + 1
                return (
                  <tr key={idx}>
                    <td style={td}>{i}</td>
                    <td style={td}>{fmtMoney(m.earnings)}</td>
                    <td style={td}>{dur}</td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        ) : (
          <div style={{ opacity: 0.7 }}>No maps yet.</div>
        )}
      </div>
    </section>
  )
}
