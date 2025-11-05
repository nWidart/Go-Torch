export const fmtDur = (startMs: number, endMs?: number): string => {
  if (!startMs) return '0s'
  const end = endMs && endMs > 0 ? endMs : Date.now()
  const ms = Math.max(0, end - startMs)
  const s = Math.floor(ms / 1000)
  const m = Math.floor(s / 60)
  const h = Math.floor(m / 60)
  const rems = s % 60
  const remm = m % 60
  const parts: string[] = []
  if (h) parts.push(`${h}h`)
  if (h || remm) parts.push(`${remm}m`)
  parts.push(`${rems}s`)
  return parts.join(' ')
}
