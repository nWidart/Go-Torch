export type UIEvent = { time: number; kind: string }

export type UIState = {
  inMap: boolean
  sessionStart: number
  sessionEnd: number
  totalDrops: number
  tally: Record<string, number>
  recent: UIEvent[]
}
