import React, { useEffect, useState } from 'react'
import HeaderBar from './components/HeaderBar'
import WelcomeSection from './components/WelcomeSection'
import StatsPanel from './components/StatsPanel'
import TallyTable from './components/TallyTable'
import RecentEvents from './components/RecentEvents'
import { hasRuntime, getBackend } from './lib/runtime'
import type { UIState } from './types/ui'

export default function App() {
  const [uid, setUid] = useState('')
  const [state, setUiState] = useState<UIState | null>(null)
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
    const off = rt.EventsOn('state', (s: UIState) => setUiState(s))
    return () => {
      try { off() } catch {}
    }
  }, [])

  const startTracking = async () => {
    if (!hasRuntime()) return alert('This UI must be run via Wails to access backend.')
    try {
      let path = logPath
      const backend = getBackend()
      if (!backend) throw new Error('Backend not available')
      if (!path) {
        const res = await backend.SelectLogFile()
        if (!res) return
        path = res
        setLogPath(res)
      }
      // Persist preference
      try { localStorage.setItem('readFromStart', readFromStart ? '1' : '0') } catch {}
      if (backend.StartTrackingWithOptions) {
        await backend.StartTrackingWithOptions(path, readFromStart)
      } else if (backend.StartTracking) {
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
      const backend = getBackend()
      await backend?.Reset()
    } catch (e) {
      console.error(e)
    }
  }

  return (
    <div style={{ fontFamily: 'Inter, system-ui, Arial', padding: 24, color: '#e2e8f0', background: '#0f172a', minHeight: '100vh' }}>
      <HeaderBar onStart={startTracking} onReset={reset} />

      <WelcomeSection
        uid={uid}
        onUidChange={setUid}
        logPath={logPath}
        onLogPathChange={setLogPath}
        readFromStart={readFromStart}
        onToggleReadFromStart={setReadFromStart}
        onStart={startTracking}
      />

      <StatsPanel state={state} />

      <section style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
        <TallyTable tally={state?.tally} />
        <RecentEvents events={state?.recent} />
      </section>

      <footer style={{ marginTop: 24, opacity: 0.7 }}>v0 — Parsing BagMgr deltas; mapping to names/prices will be added later.</footer>
    </div>
  )
}
