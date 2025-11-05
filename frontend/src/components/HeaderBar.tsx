import React from 'react'
import { btnStyle } from '../uiStyles'

export type HeaderBarProps = {
  onStart: () => void
  onReset: () => void
}

export default function HeaderBar({ onStart, onReset }: HeaderBarProps) {
  return (
    <header style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
      <h1 style={{ margin: 0, fontSize: 20 }}>GoTorch — Torchlight Infinite Tracker</h1>
      <div>
        <button onClick={onReset} style={btnStyle}>Reset</button>
        <button onClick={onStart} style={{ ...btnStyle, marginLeft: 8 }}>Start</button>
      </div>
    </header>
  )
}
