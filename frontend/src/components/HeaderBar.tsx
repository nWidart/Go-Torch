import React from 'react'
import { btnStyle } from '../uiStyles'

export type HeaderBarProps = {
  running: boolean
  paused: boolean
  onStart: () => void
  onStop: () => void
  onPause: () => void
  onResume: () => void
  onReset: () => void
}

export default function HeaderBar({ running, paused, onStart, onStop, onPause, onResume, onReset }: HeaderBarProps) {
  const dangerBtn: React.CSSProperties = { ...btnStyle, marginLeft: 8, background: '#7f1d1d', borderColor: '#b91c1c' }
  return (
    <header style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
      <h1 style={{ margin: 0, fontSize: 20 }}>GoTorch — Torchlight Infinite Tracker</h1>
      <div>
        <button onClick={onReset} style={btnStyle}>Reset</button>
        {!running ? (
          <button onClick={onStart} style={{ ...btnStyle, marginLeft: 8 }}>Start Session</button>
        ) : (
          <>
            <button onClick={onStop} style={dangerBtn}>Stop Session</button>
            {!paused ? (
              <button onClick={onPause} style={{ ...btnStyle, marginLeft: 8 }}>Pause</button>
            ) : (
              <button onClick={onResume} style={{ ...btnStyle, marginLeft: 8 }}>Resume</button>
            )}
          </>
        )}
      </div>
    </header>
  )
}
