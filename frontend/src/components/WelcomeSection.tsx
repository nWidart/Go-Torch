import React from 'react'
import { btnStyle, card, h2, inputStyle } from '../uiStyles'

export type WelcomeSectionProps = {
  uid: string
  onUidChange: (v: string) => void
  logPath: string
  onLogPathChange: (v: string) => void
  readFromStart: boolean
  onToggleReadFromStart: (v: boolean) => void
  running: boolean
  paused: boolean
  onStart: () => void
  onStop: () => void
  onPause: () => void
  onResume: () => void
  selectLogFile: () => void
}

export default function WelcomeSection(props: WelcomeSectionProps) {
  const { uid, onUidChange, logPath, onLogPathChange, readFromStart, onToggleReadFromStart, running, paused, onStart, onStop, onPause, onResume, selectLogFile } = props
  const dangerBtn: React.CSSProperties = { ...btnStyle, background: '#7f1d1d', borderColor: '#b91c1c' }
  return (
    <section style={card}>
      <h2 style={h2}>Welcome</h2>
      <div style={{ display: 'flex', gap: 12, alignItems: 'center', flexWrap: 'wrap' }}>
        <label>Log Path:</label>
        <button onClick={() => selectLogFile()}>Select log file</button>
        <input value={logPath} onChange={e => onLogPathChange(e.target.value)} placeholder="Select or paste UE_game.log path" style={inputStyle} />
        <label style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <input type="checkbox" checked={readFromStart} onChange={e => onToggleReadFromStart(e.target.checked)} />
          <span>Read old lines (dev)</span>
        </label>
        {!running ? (
          <button onClick={onStart} style={btnStyle}>Start Session</button>
        ) : (
          <>
            <button onClick={onStop} style={dangerBtn}>Stop Session</button>
            {!paused ? (
              <button onClick={onPause} style={btnStyle}>Pause</button>
            ) : (
              <button onClick={onResume} style={btnStyle}>Resume</button>
            )}
          </>
        )}
      </div>
    </section>
  )
}
