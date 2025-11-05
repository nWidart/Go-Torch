import React from 'react'
import { card, h2 } from '../uiStyles'
import { UIEvent } from '../types/ui'

export type RecentEventsProps = {
  events?: UIEvent[]
}

export default function RecentEvents({ events }: RecentEventsProps) {
  return (
    <div style={card}>
      <h2 style={h2}>Recent Events</h2>
      <div style={{ maxHeight: 300, overflowY: 'auto' }}>
        {events?.length ? (
          events.slice(-50).map((e, i) => (
            <div key={i} style={{ padding: '4px 0', borderBottom: '1px solid #1e293b' }}>
              <span style={{ opacity: 0.7, marginRight: 8 }}>{new Date(e.time).toLocaleTimeString()}</span>
              <span>{e.kind}</span>
            </div>
          ))
        ) : (
          <div style={{ opacity: 0.7 }}>No events yet.</div>
        )}
      </div>
    </div>
  )
}
