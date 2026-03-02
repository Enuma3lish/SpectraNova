/**
 * Format an ISO date string into a human-readable form.
 * e.g. "2025-06-15T10:30:00Z" → "Jun 15, 2025"
 */
export function formatDate(isoDate: string): string {
  const date = new Date(isoDate)
  if (isNaN(date.getTime())) return isoDate

  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}
