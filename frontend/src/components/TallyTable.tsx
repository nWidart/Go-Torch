import React from 'react'
import { card, h2, td, th } from '../uiStyles'

export type TallyTableProps = {
  tally?: Record<string, number>
}

export default function TallyTable({ tally }: TallyTableProps) {
  const hasRows = !!tally && Object.keys(tally).length > 0
  return (
    <div style={card}>
      <h2 style={h2}>Tally (by ConfigBaseID)</h2>
      {hasRows ? (
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr>
              <th style={th}>ConfigBaseID</th>
              <th style={th}>Count</th>
            </tr>
          </thead>
          <tbody>
            {Object.entries(tally!).sort().map(([id, n]) => (
              <tr key={id}>
                <td style={td}>{id}</td>
                <td style={td}>{n}</td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <div style={{ opacity: 0.7 }}>No items counted yet.</div>
      )}
    </div>
  )
}
