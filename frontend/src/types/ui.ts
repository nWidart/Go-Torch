export type UIEvent = { time: number; kind: string }

export type UITallyItem = {
  name: string
  type: string
  price: number
  last_update: number
  from: string
  count: number
}

export type UIState = {
  inMap: boolean
  sessionStart: number
  sessionEnd: number
  totalDrops: number
  tally: Record<string, UITallyItem>
  recent: UIEvent[]
}
