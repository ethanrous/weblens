import { UserInfoT } from './types/Types'

export function humanFileSize(
    bytes: number,
    si = true,
    dp = 1
): [string, string] {
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
    } while (
        Math.round(Math.abs(bytes) * r) / r >= thresh &&
        u < units.length - 1
    )

    return [bytes.toFixed(dp), units[u]]
}

const NS_IN_MILLISECOND = 1000 * 1000
const NS_IN_SECOND = NS_IN_MILLISECOND * 1000
const NS_IN_MINUTE = NS_IN_SECOND * 60
const NS_IN_HOUR = NS_IN_MINUTE * 60

export function nsToHumanTime(ns: number) {
    let timeStr = ''

    const hours = Math.floor(ns / NS_IN_HOUR)
    if (hours >= 1) {
        timeStr += hours + ' Hours '
        ns = ns % NS_IN_HOUR
    }

    const minutes = Math.floor(ns / NS_IN_MINUTE)
    if (minutes >= 1) {
        timeStr += minutes + ' Minutes '
        ns = ns % NS_IN_MINUTE
    }

    const seconds = Math.floor(ns / NS_IN_SECOND)
    if (seconds >= 1) {
        timeStr += seconds + ' Second'
        if (seconds >= 2) {
            timeStr += 's'
        }
        timeStr += ' '
        ns = ns % NS_IN_SECOND
    }

    if (seconds === 0) {
        const milliseconds = Math.floor(ns / NS_IN_MILLISECOND)
        if (milliseconds >= 1) {
            timeStr += milliseconds + 'ms '
        }
    }

    if (timeStr.length === 0) {
        timeStr = '<1ms'
    }

    return timeStr
}

export const clamp = (value: number, min: number, max: number) =>
    Math.min(Math.max(value, min), max)

export function getRandomInt(min: number, max: number): number {
    return Math.floor(min + Math.random() * (max - min))
}

export function friendlyFolderName(
    folderName: string,
    folderId: string,
    usr: UserInfoT
): string {
    if (folderId === usr.homeId) {
        return 'Home'
    } else if (folderId === usr.trashId) {
        return 'Trash'
    } else if (folderName === usr.username) {
        return 'Home'
    } else if (folderName === '.user_trash') {
        return 'Trash'
    }
    return folderName
}

export function binarySearch<T>(
    values: T[],
    target: T,
    cmp: (a: T, b: T) => number
): number {
    let left: number = 0
    let right: number = values.length - 1

    while (left <= right) {
        const mid: number = Math.floor((left + right) / 2)

        const result = cmp(values[mid], target)
        if (result === 0) {
            return mid
        } else if (result > 0) {
            right = mid - 1
        } else {
            left = mid + 1
        }
    }

    return -1
}
