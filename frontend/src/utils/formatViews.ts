/**
 * Format a view count into a compact, human-readable string.
 * e.g. 0 → "0", 999 → "999", 1200 → "1.2K", 1500000 → "1.5M", 2000000000 → "2B"
 */
export function formatViews(count: number): string {
  if (count < 0) return '0'

  if (count < 1_000) {
    return count.toString()
  }

  if (count < 1_000_000) {
    const value = count / 1_000
    return `${parseFloat(value.toFixed(1))}K`
  }

  if (count < 1_000_000_000) {
    const value = count / 1_000_000
    return `${parseFloat(value.toFixed(1))}M`
  }

  const value = count / 1_000_000_000
  return `${parseFloat(value.toFixed(1))}B`
}
