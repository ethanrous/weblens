import { useCallback, useEffect, useState } from "react";
import { fetchMediaTypes } from "../api/ApiFetch";
import { mediaType } from "../types/Types";

export const useResize = (elem: HTMLDivElement) => {
    const [size, setSize] = useState({ height: 0, width: 0 });
    useEffect(() => {
        if (elem) {
            const obs = new ResizeObserver((entries) => {
                // only 1 entry
                setSize({ height: elem.clientHeight, width: elem.clientWidth });
            });
            obs.observe(elem);
            return () => obs.disconnect();
        }
    }, [elem]);

    return size;
};

export const useKeyDown = (key: string, callback: (e) => void) => {
    const onKeyDown = useCallback(
        (event) => {
            if (event.key === key) {
                callback(event);
            }
        },
        [key, callback],
    );

    useEffect(() => {
        document.addEventListener("keydown", onKeyDown);
        return () => {
            document.removeEventListener("keydown", onKeyDown);
        };
    }, [onKeyDown]);
};

export const useWindowSize = () => {
    const [windowSize, setWindowSize] = useState({ width: window.innerWidth, height: window.innerHeight });
    const onResize = (e) => setWindowSize({ width: e.target.innerWidth, height: e.target.innerHeight });
    useEffect(() => {
        window.addEventListener("resize", onResize);
        return () => window.removeEventListener("resize", onResize);
    }, []);

    return windowSize;
};

export const useResizeDrag = (resizing: boolean, setResizing, setResizeOffset, flip?: boolean) => {
    const windowSize = useWindowSize();
    window.addEventListener("mouseup", (e) => {
        if (resizing) {
            setResizing(false);
        }
    });

    const onMove = (event) => {
        if (flip) {
            setResizeOffset(windowSize.width - event.clientX);
        } else {
            setResizeOffset(event.clientX);
        }
    };

    useEffect(() => {
        if (resizing) {
            addEventListener("mousemove", onMove);
        }
        return () => removeEventListener("mousemove", onMove);
    }, [resizing]);
};

export const useMediaType = (): Map<string, mediaType> => {
    const [typeMap, setTypeMap] = useState(null)

    useEffect(() => {
        const mediaTypes = new Map<string, mediaType>()
        fetchMediaTypes().then((mt) => {
            const mimes: string[] = Array.from(Object.keys(mt));
            for (const mime of mimes) {
                mediaTypes.set(mime, mt[mime]);
            }
            setTypeMap(mediaTypes)
        });
    }, [])
    return typeMap
}
// export async function VerifyMediaTypeMap() {
//     if (MediaTypes.size === 0) {
//     }
// }

// export function GetMediaType(mimeType: string): mediaType {
//     return MediaTypes.get(mimeType);
// }