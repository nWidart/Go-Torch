import React, { useMemo } from 'react'
import { card, h2 } from '../uiStyles'
import { fmtDur } from '../lib/format'
import { UIState } from '../types/ui'
import Stat from './Stat'

export type StatsPanelProps = {
  state: UIState | null
}

export default function StatsPanel({ state }: StatsPanelProps) {
  const status = state?.inMap ? 'In Map' : 'Idle'
  const sessionDur = useMemo(() => fmtDur(state?.sessionStart || 0, state?.inMap ? undefined : state?.sessionEnd), [state])
  const itemsPerHour = useMemo(() => {
    if (!state?.inMap && state?.sessionStart && state?.sessionEnd && state.sessionEnd > state.sessionStart) {
      const durH = (state.sessionEnd - state.sessionStart) / 3600000
      const iph = durH > 0 ? (state.totalDrops / durH) : 0
      return iph.toFixed(1)
    }
    return '—'
  }, [state])

  return (
    <section style={card}>
      <h2 style={h2}>Live Stats</h2>
      <div style={{ display: 'flex', gap: 24 }}>
        <Stat label="Status" value={status} />
        <Stat label="Session Duration" value={sessionDur} />
        <Stat label="Session Total" value={String(state?.totalDrops ?? 0)} />
        <Stat label="Items/hour" value={itemsPerHour} />
      </div>
    </section>
  )
}
