export type UIEvent = { time: number; kind: string }

export type UITallyItem = {
  name: string
  type: string
  price: number
  last_update: number
  from: string
  count: number
}

export type UIMap = {
  start: number
  end: number
  durationMs: number
  earnings: number
}

export type UIState = {
  inMap: boolean
  sessionStart: number
  sessionEnd: number
  sessionPaused: boolean
  pausedAt: number
  pausedAccumMs: number
  mapStart: number
  mapEnd: number
  totalDrops: number
  tally: Record<string, UITallyItem>
  recent: UIEvent[]
  maps: UIMap[]
  earningsPerSession: number
  earningsPerHour: number
  avgMapTimeMs: number
}
