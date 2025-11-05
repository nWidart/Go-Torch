import React from 'react'
import { btnStyle, card, h2, inputStyle } from '../uiStyles'

export type WelcomeSectionProps = {
  uid: string
  onUidChange: (v: string) => void
  logPath: string
  onLogPathChange: (v: string) => void
  readFromStart: boolean
  onToggleReadFromStart: (v: boolean) => void
  onStart: () => void
  selectLogFile: () => void
}

export default function WelcomeSection(props: WelcomeSectionProps) {
  const { uid, onUidChange, logPath, onLogPathChange, readFromStart, onToggleReadFromStart, onStart, selectLogFile } = props
  return (
    <section style={card}>
      <h2 style={h2}>Welcome</h2>
      <div style={{ display: 'flex', gap: 12, alignItems: 'center', flexWrap: 'wrap' }}>
        <label>User ID:</label>
        <input value={uid} onChange={e => onUidChange(e.target.value)} placeholder="Enter user id (no-op)" style={inputStyle} />
        <label>Log Path:</label>
        <button onClick={() => selectLogFile()}>Select log file</button>
        <input value={logPath} onChange={e => onLogPathChange(e.target.value)} placeholder="Select or paste UE_game.log path" style={inputStyle} />
        <label style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <input type="checkbox" checked={readFromStart} onChange={e => onToggleReadFromStart(e.target.checked)} />
          <span>Read old lines (dev)</span>
        </label>
        <button onClick={onStart} style={btnStyle}>Start</button>
      </div>
    </section>
  )
}
