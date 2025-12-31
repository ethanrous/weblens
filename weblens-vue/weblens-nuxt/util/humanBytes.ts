export function humanBytes(bytes: number, si = true, dp = 1): [string, string] {
    if (!bytes) {
        return ['0', 'B']
    }
    const thresh = si ? 1000 : 1024

    if (bytes === undefined) {
        return ['0', 'B']
    }
    if (Math.abs(bytes) < thresh) {
        return [bytes.toString(), 'B']
    }

    const units = si
        ? ['kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
        : ['KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB']
    let u = -1
    const r = 10 ** dp

    do {
        bytes /= thresh
        ++u
    } while (Math.round(Math.abs(bytes) * r) / r >= thresh && u < units.length - 1)

    return [bytes.toFixed(dp), units[u]]
}

export function humanBytesStr(bytes: number): string {
    return humanBytes(bytes).join('')
}

export function humanDuration(s: number, doMs: boolean = false): string {
    if (!s) {
        return '-'
    }

    const ms = s % 1000
    s = (s - ms) / 1000
    const secs = s % 60
    s = (s - secs) / 60
    const mins = s % 60
    const hrs = (s - mins) / 60

    if (!doMs && hrs + mins + secs === 0) {
        return '<1s'
    }

    return (
        (hrs ? hrs + 'h ' : '') +
        (mins ? mins + 'm ' : '') +
        (secs ? secs + 's ' : '') +
        (doMs ? Math.trunc(ms) + 'ms' : '')
    )
}
