import React from 'react'
import { card, h2, td, th } from '../uiStyles'
import type { UITallyItem } from '../types/ui'

export type TallyTableProps = {
  tally?: Record<string, UITallyItem>
}

export default function TallyTable({ tally }: TallyTableProps) {
  const hasRows = !!tally && Object.keys(tally).length > 0
  const rows = hasRows ? Object.entries(tally!) : []
  // Sort by total value descending
  rows.sort((a, b) => (b[1].price * b[1].count) - (a[1].price * a[1].count))

  return (
    <div style={card}>
      <h2 style={h2}>Tally</h2>
      {hasRows ? (
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr>
              <th style={th}>Item</th>
              <th style={th}>Count</th>
              <th style={th}>Total Value</th>
            </tr>
          </thead>
          <tbody>
            {rows.map(([id, item]) => {
              const total = item.price * item.count
              return (
                <tr key={id}>
                  <td style={td}>{item.name}</td>
                  <td style={td}>{item.count}</td>
                  <td style={td}>{total.toFixed(2)}</td>
                </tr>
              )
            })}
          </tbody>
        </table>
      ) : (
        <div style={{ opacity: 0.7 }}>No items counted yet.</div>
      )}
    </div>
  )
}
