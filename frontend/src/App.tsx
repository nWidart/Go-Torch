import React, { useEffect, useMemo, useState } from 'react'

// Wails will inject a global 'window.runtime' when built/run by wails dev/build.
// We guard accesses so the app still renders in plain browser preview.

type UIEvent = { time: number; kind: string }

type UIState = {
  inMap: boolean
  sessionStart: number
  sessionEnd: number
  totalDrops: number
  tally: Record<string, number>
  recent: UIEvent[]
}

const hasRuntime = () => typeof (window as any).runtime !== 'undefined'

const fmtDur = (startMs: number, endMs?: number) => {
  if (!startMs) return '0s'
  const end = endMs && endMs > 0 ? endMs : Date.now()
  const ms = Math.max(0, end - startMs)
  const s = Math.floor(ms / 1000)
  const m = Math.floor(s / 60)
  const h = Math.floor(m / 60)
  const rems = s % 60
  const remm = m % 60
  const parts: string[] = []
  if (h) parts.push(`${h}h`)
  if (h || remm) parts.push(`${remm}m`)
  parts.push(`${rems}s`)
  return parts.join(' ')
}

export default function App() {
  const [uid, setUid] = useState('')
  const [state, setState] = useState<UIState | null>(null)
  const [logPath, setLogPath] = useState('')
  const [readFromStart, setReadFromStart] = useState<boolean>(() => {
    try {
      const saved = localStorage.getItem('readFromStart')
      if (saved !== null) return saved === '1'
    } catch {}
    // Default: in dev, read old lines; in prod, tail only
    return import.meta.env.DEV
  })

  useEffect(() => {
    if (!hasRuntime()) return
    const rt = (window as any).runtime
    const off = rt.EventsOn('state', (s: UIState) => setState(s))
    return () => {
      try { off() } catch {}
    }
  }, [])

  const status = state?.inMap ? 'In Map' : 'Idle'
  const sessionDur = useMemo(() => fmtDur(state?.sessionStart || 0, state?.inMap ? undefined : state?.sessionEnd), [state])

  const startTracking = async () => {
    if (!hasRuntime()) return alert('This UI must be run via Wails to access backend.')
    try {
      let path = logPath
      if (!path) {
        // Open file dialog via backend
        const backend = (window as any).go?.app?.App
        if (!backend) throw new Error('Backend not available')
        const res = await backend.SelectLogFile()
        if (!res) return
        path = res
        setLogPath(res)
      }
      // Persist preference
      try { localStorage.setItem('readFromStart', readFromStart ? '1' : '0') } catch {}
      const backend = (window as any).go?.app?.App
      if (backend?.StartTrackingWithOptions) {
        await backend.StartTrackingWithOptions(path, readFromStart)
      } else if (backend?.StartTracking) {
        // fallback: older backend without options will default to tailing from end
        await backend.StartTracking(path)
      } else {
        throw new Error('Backend not available')
      }
    } catch (e) {
      console.error(e)
      alert('Failed to start tracking: ' + e)
    }
  }

  const reset = async () => {
    if (!hasRuntime()) return
    try {
      await (window as any).go.app.App.Reset()
    } catch (e) {
      console.error(e)
    }
  }

  return (
    <div style={{ fontFamily: 'Inter, system-ui, Arial', padding: 24, color: '#e2e8f0', background: '#0f172a', minHeight: '100vh' }}>
      <header style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
        <h1 style={{ margin: 0, fontSize: 20 }}>GoTorch — Torchlight Infinite Tracker</h1>
        <div>
          <button onClick={reset} style={btnStyle}>Reset</button>
          <button onClick={startTracking} style={{ ...btnStyle, marginLeft: 8 }}>Start</button>
        </div>
      </header>

      <section style={card}>
        <h2 style={h2}>Welcome</h2>
        <div style={{ display: 'flex', gap: 12, alignItems: 'center', flexWrap: 'wrap' }}>
          <label>User ID:</label>
          <input value={uid} onChange={e => setUid(e.target.value)} placeholder="Enter user id (no-op)" style={inputStyle} />
          <label>Log Path:</label>
          <input value={logPath} onChange={e => setLogPath(e.target.value)} placeholder="Select or paste UE_game.log path" style={inputStyle} />
          <label style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <input type="checkbox" checked={readFromStart} onChange={e => setReadFromStart(e.target.checked)} />
            <span>Read old lines (dev)</span>
          </label>
          <button onClick={startTracking} style={btnStyle}>Start</button>
        </div>
      </section>

      <section style={card}>
        <h2 style={h2}>Live Stats</h2>
        <div style={{ display: 'flex', gap: 24 }}>
          <Stat label="Status" value={status} />
          <Stat label="Session Duration" value={sessionDur} />
          <Stat label="Session Total" value={String(state?.totalDrops ?? 0)} />
          <Stat label="Items/hour" value={(() => {
            if (!state?.inMap && state?.sessionStart && state?.sessionEnd && state.sessionEnd > state.sessionStart) {
              const durH = (state.sessionEnd - state.sessionStart) / 3600000
              const iph = durH > 0 ? (state.totalDrops / durH) : 0
              return iph.toFixed(1)
            }
            return '—'
          })()} />
        </div>
      </section>

      <section style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
        <div style={card}>
          <h2 style={h2}>Tally (by ConfigBaseID)</h2>
          {state && Object.keys(state.tally).length ? (
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr><th style={th}>ConfigBaseID</th><th style={th}>Count</th></tr>
              </thead>
              <tbody>
                {Object.entries(state.tally).sort().map(([id, n]) => (
                  <tr key={id}><td style={td}>{id}</td><td style={td}>{n}</td></tr>
                ))}
              </tbody>
            </table>
          ) : (
            <div style={{ opacity: 0.7 }}>No items counted yet.</div>
          )}
        </div>
        <div style={card}>
          <h2 style={h2}>Recent Events</h2>
          <div style={{ maxHeight: 300, overflowY: 'auto' }}>
            {state?.recent?.length ? state.recent.slice(-50).map((e, i) => (
              <div key={i} style={{ padding: '4px 0', borderBottom: '1px solid #1e293b' }}>
                <span style={{ opacity: 0.7, marginRight: 8 }}>{new Date(e.time).toLocaleTimeString()}</span>
                <span>{e.kind}</span>
              </div>
            )) : <div style={{ opacity: 0.7 }}>No events yet.</div>}
          </div>
        </div>
      </section>

      <footer style={{ marginTop: 24, opacity: 0.7 }}>v0 — Parsing BagMgr deltas; mapping to names/prices will be added later.</footer>
    </div>
  )
}

function Stat({ label, value }: { label: string, value: string }) {
  return (
    <div>
      <div style={{ opacity: 0.7, marginBottom: 4 }}>{label}</div>
      <div style={{ fontSize: 20 }}>{value}</div>
    </div>
  )
}

const card: React.CSSProperties = {
  background: '#0b1220',
  border: '1px solid #1e293b',
  borderRadius: 8,
  padding: 16,
  marginBottom: 16,
}

const h2: React.CSSProperties = { margin: '0 0 12px 0', fontSize: 16 }

const btnStyle: React.CSSProperties = {
  background: '#1e293b',
  color: '#e2e8f0',
  border: '1px solid #334155',
  borderRadius: 6,
  padding: '8px 12px',
  cursor: 'pointer'
}

const inputStyle: React.CSSProperties = {
  background: '#0f172a',
  color: '#e2e8f0',
  border: '1px solid #334155',
  borderRadius: 6,
  padding: '8px 12px',
  minWidth: 300,
}
