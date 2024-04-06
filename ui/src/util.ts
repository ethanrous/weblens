import { FileInfoT } from "./types/Types";

export function humanFileSize(bytes, si = true, dp = 1) {
    if (!bytes) {
        return [0, "B"];
    }
    const thresh = si ? 1000 : 1024;

    if (bytes === undefined) {
        return [0, "B"];
    }
    if (Math.abs(bytes) < thresh) {
        return [bytes, "B"];
    }

    const units = si
        ? ["kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"]
        : ["KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"];
    let u = -1;
    const r = 10 ** dp;

    do {
        bytes /= thresh;
        ++u;
    } while (Math.round(Math.abs(bytes) * r) / r >= thresh && u < units.length - 1);

    return [bytes.toFixed(dp), units[u]];
}

const NS_IN_MILLISECOND = 1000 * 1000;
const NS_IN_SECOND = NS_IN_MILLISECOND * 1000;
const NS_IN_MINUTE = NS_IN_SECOND * 60;
const NS_IN_HOUR = NS_IN_MINUTE * 60;

export function nsToHumanTime(ns: number) {
    let timeStr = "";

    const hours = Math.floor(ns / NS_IN_HOUR);
    if (hours >= 1) {
        timeStr += hours + " Hours ";
        ns = ns % NS_IN_HOUR;
    }

    const minutes = Math.floor(ns / NS_IN_MINUTE);
    if (minutes >= 1) {
        timeStr += minutes + " Minutes ";
        ns = ns % NS_IN_MINUTE;
    }

    const seconds = Math.floor(ns / NS_IN_SECOND);
    if (seconds >= 1) {
        timeStr += seconds + " Seconds ";
        ns = ns % NS_IN_SECOND;
    }

    const milliseconds = Math.floor(ns / NS_IN_MILLISECOND);
    if (milliseconds >= 1) {
        timeStr += milliseconds + "ms ";
    }

    return timeStr;
}

export function dateFromFileData(fileData: FileInfoT) {
    var date = new Date(fileData.mediaData.createDate);
    if (date.getFullYear() === 0) {
        date = new Date(fileData.modTime);
    }
    return date.toDateString();
}

export const clamp = (value: number, min: number, max: number) => Math.min(Math.max(value, min), max);

export function hexToComplimentary(hex: string){

    // Convert hex to rgb
    // Credit to Denis http://stackoverflow.com/a/36253499/4939630
    const rgbStr = 'rgb(' + (hex = hex.replace('#', '')).match(new RegExp('(.{' + hex.length/3 + '})', 'g')).map(function(l) { return parseInt(hex.length%2 ? l+l : l, 16); }).join(',') + ')';

    // Get array of RGB values
    const rgb = rgbStr.replace(/[^\d,]/g, '').split(',');

    var r = Number(rgb[0]), g = Number(rgb[1]), b = Number(rgb[2]);

    // Convert RGB to HSL
    // Adapted from answer by 0x000f http://stackoverflow.com/a/34946092/4939630
    r /= 255.0;
    g /= 255.0;
    b /= 255.0;
    var max = Math.max(r, g, b);
    var min = Math.min(r, g, b);
    var h, s, l = (max + min) / 2.0;

    if(max == min) {
        h = s = 0;  //achromatic
    } else {
        var d = max - min;
        s = (l > 0.5 ? d / (2.0 - max - min) : d / (max + min));

        if(max == r && g >= b) {
            h = 1.0472 * (g - b) / d ;
        } else if(max == r && g < b) {
            h = 1.0472 * (g - b) / d + 6.2832;
        } else if(max == g) {
            h = 1.0472 * (b - r) / d + 2.0944;
        } else if(max == b) {
            h = 1.0472 * (r - g) / d + 4.1888;
        }
    }

    h = h / 6.2832 * 360.0 + 0;

    // Shift hue to opposite side of wheel and convert to [0-1] value
    h+= 180;
    if (h > 360) { h -= 360; }
    h /= 360;

    // Convert h s and l values into r g and b values
    // Adapted from answer by Mohsen http://stackoverflow.com/a/9493060/4939630
    if(s === 0){
        r = g = b = l; // achromatic
    } else {
        var hue2rgb = function hue2rgb(p, q, t){
            if(t < 0) t += 1;
            if(t > 1) t -= 1;
            if(t < 1/6) return p + (q - p) * 6 * t;
            if(t < 1/2) return q;
            if(t < 2/3) return p + (q - p) * (2/3 - t) * 6;
            return p;
        };

        var q = l < 0.5 ? l * (1 + s) : l + s - l * s;
        var p = 2 * l - q;

        r = hue2rgb(p, q, h + 1/3);
        g = hue2rgb(p, q, h);
        b = hue2rgb(p, q, h - 1/3);
    }

    r = Math.round(r * 255);
    g = Math.round(g * 255);
    b = Math.round(b * 255);

    // Convert r b and g values to hex
    const rgbFinal = b | (g << 8) | (r << 16);
    return "#" + (0x1000000 | rgbFinal).toString(16).substring(1);
}
