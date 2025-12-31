export function relativeTimeAgo(timestampMs: number): string {
    const now = Date.now()
    const diffMs = now - timestampMs

    if (diffMs < 0) {
        return 'in the future'
    }

    const seconds = Math.floor(diffMs / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)
    const days = Math.floor(hours / 24)
    const weeks = Math.floor(days / 7)
    const months = Math.floor(days / 30)
    const years = Math.floor(days / 365)

    if (seconds < 60) {
        return `${seconds} sec.ago`
    }

    if (minutes < 60) {
        return `${minutes} min. ago`
    }

    if (hours < 24) {
        return `${hours} hr. ago`
    }

    if (days < 7) {
        return days === 1 ? '1 day ago' : `${days} days ago`
    }

    if (weeks < 4) {
        return weeks === 1 ? '1 week ago' : `${weeks} weeks ago`
    }

    if (months < 12) {
        return `${months} mo. ago`
    }

    return `${years} yr. ago`
}
