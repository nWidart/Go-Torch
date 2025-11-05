import React from 'react'

export function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div style={{ opacity: 0.7, marginBottom: 4 }}>{label}</div>
      <div style={{ fontSize: 20 }}>{value}</div>
    </div>
  )
}

export default Stat
