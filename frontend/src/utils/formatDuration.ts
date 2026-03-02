/**
 * Format a duration in seconds to mm:ss (or h:mm:ss for videos >= 1 hour).
 * e.g. 125 → "2:05", 3661 → "1:01:01"
 */
export function formatDuration(seconds: number): string {
  if (seconds < 0 || !Number.isFinite(seconds)) return '0:00'

  const totalSeconds = Math.floor(seconds)
  const h = Math.floor(totalSeconds / 3600)
  const m = Math.floor((totalSeconds % 3600) / 60)
  const s = totalSeconds % 60

  const paddedSeconds = s.toString().padStart(2, '0')

  if (h > 0) {
    const paddedMinutes = m.toString().padStart(2, '0')
    return `${h}:${paddedMinutes}:${paddedSeconds}`
  }

  return `${m}:${paddedSeconds}`
}
