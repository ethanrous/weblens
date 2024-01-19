import { fileData } from './types/Types'

export function humanFileSize(bytes, si = true, dp = 1) {
    if (!bytes) {
        return [0, "B"]
    }
    const thresh = si ? 1000 : 1024;

    if (bytes === undefined) {
        return [0, 'B']
    }
    if (Math.abs(bytes) < thresh) {
        return [bytes, 'B'];
    }

    const units = si
        ? ['kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
        : ['KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB'];
    let u = -1;
    const r = 10 ** dp;

    do {
        bytes /= thresh;
        ++u;
    } while (Math.round(Math.abs(bytes) * r) / r >= thresh && u < units.length - 1);

    return [bytes.toFixed(dp), units[u]];
}

export function dateFromFileData(fileData: fileData) {
    var date = new Date(fileData.mediaData.createDate)
    if (date.getFullYear() === 0) {
        date = new Date(fileData.modTime)
    }
    return date.toDateString()
}