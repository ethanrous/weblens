import { RefObject, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { fetchMediaTypes } from '../api/ApiFetch';
import { mediaType } from '../types/Types';
import WeblensMedia from '../Media/Media';
import { MediaContext } from '../Context';

export const useResize = (elem: HTMLDivElement) => {
    const [size, setSize] = useState({ height: -1, width: -1 });
    useEffect(() => {
        if (elem) {
            setSize({ height: elem.clientHeight, width: elem.clientWidth });
            const obs = new ResizeObserver(entries => {
                // only 1 entry
                setSize({ height: elem.clientHeight, width: elem.clientWidth });
            });
            obs.observe(elem);
            return () => obs.disconnect();
        }
    }, [elem]);

    return size;
};

export const useVideo = (elem: HTMLVideoElement) => {
    const [playtime, setPlaytime] = useState(0);
    const [isPlaying, setIsPlaying] = useState(false);
    const [isWaiting, setIsWaiting] = useState(true);

    const updatePlaytime = useCallback(() => {
        setPlaytime(elem.currentTime);
    }, [setPlaytime, elem]);

    const updatePlayState = useCallback(
        e => {
            setIsPlaying(e.type === 'play');
        },
        [setIsPlaying],
    );

    const updateBufferState = useCallback(
        e => {
            if (e.type === 'waiting') {
                setIsWaiting(true);
            } else if (e.type === 'playing') {
                setIsWaiting(false);
            }
        },
        [setIsWaiting],
    );

    useEffect(() => {
        if (!elem) {
            return;
        }
        elem.addEventListener('timeupdate', updatePlaytime);
        elem.addEventListener('play', updatePlayState);
        elem.addEventListener('pause', updatePlayState);
        elem.addEventListener('waiting', updateBufferState);
        elem.addEventListener('playing', updateBufferState);
        return () => {
            elem.removeEventListener('timeupdate', updatePlaytime);
            elem.removeEventListener('play', updatePlayState);
            elem.removeEventListener('pause', updatePlayState);
            elem.removeEventListener('waiting', updateBufferState);
            elem.removeEventListener('playing', updateBufferState);
        };
    }, [updatePlaytime, updatePlayState, elem]);

    return { playtime, isPlaying, isWaiting };
};

export const useKeyDown = (
    key: string | ((e: KeyboardEvent) => boolean),
    callback: (e: KeyboardEvent) => void,
    ignore?: boolean,
) => {
    const onKeyDown = useCallback(
        event => {
            if ((typeof key === 'string' && event.key === key) || (typeof key === 'function' && key(event))) {
                callback(event);
            }
        },
        [key, callback],
    );

    useEffect(() => {
        if (ignore === true) {
            return;
        }
        document.addEventListener('keydown', onKeyDown);
        return () => {
            document.removeEventListener('keydown', onKeyDown);
        };
    }, [onKeyDown, ignore]);
};

export const useWindowSize = () => {
    const [windowSize, setWindowSize] = useState({
        width: window.innerWidth,
        height: window.innerHeight,
    });
    const onResize = e =>
        setWindowSize({
            width: e.target.innerWidth,
            height: e.target.innerHeight,
        });
    useEffect(() => {
        window.addEventListener('resize', onResize);
        return () => window.removeEventListener('resize', onResize);
    }, []);

    return windowSize;
};

export const useResizeDrag = (resizing: boolean, setResizing, setResizeOffset, flip?: boolean) => {
    const windowSize = useWindowSize();
    window.addEventListener('mouseup', e => {
        if (resizing) {
            setResizing(false);
        }
    });

    const onMove = event => {
        if (flip) {
            setResizeOffset(windowSize.width - event.clientX);
        } else {
            setResizeOffset(event.clientX);
        }
    };

    useEffect(() => {
        if (resizing) {
            addEventListener('mousemove', onMove);
        }
        return () => removeEventListener('mousemove', onMove);
    }, [resizing]);
};

export const useMediaType = (): Map<string, mediaType> => {
    const [typeMap, setTypeMap] = useState(null);

    useEffect(() => {
        const mediaTypes = new Map<string, mediaType>();
        fetchMediaTypes().then(mt => {
            const mimes: string[] = Array.from(Object.keys(mt));
            for (const mime of mimes) {
                mediaTypes.set(mime, mt[mime]);
            }
            setTypeMap(mediaTypes);
        });
    }, []);
    return typeMap;
};

export const useClick = (handler: (e) => void, ignore?, disable?: boolean) => {
    const callback = useCallback(
        e => {
            if (disable) {
                return;
            }

            if (ignore && ignore.contains(e.target)) {
                return;
            }

            handler(e);
        },
        [handler, ignore, disable],
    );

    useEffect(() => {
        if (!disable) {
            window.addEventListener('click', callback, true);
            window.addEventListener('contextmenu', callback, true);
        } else {
            return;
        }
        return () => {
            window.removeEventListener('click', callback, true);
            window.removeEventListener('contextmenu', callback, true);
        };
    }, [callback, disable]);
};

export const useMedia = (mediaId: string): WeblensMedia => {
    const { mediaState } = useContext(MediaContext);
    const [mediaData, setMediaData] = useState<WeblensMedia>(null);

    useEffect(() => {
        mediaState.loadNew(mediaId).then(m => setMediaData(m));
    }, [mediaId]);

    return mediaData;
};

export const useIsFocused = (element: HTMLDivElement) => {
    const [active, setActive] = useState<boolean>(false);

    const handleFocusIn = e => {
        setActive(true);
    };

    const handleFocusOut = e => {
        setActive(false);
    };

    useEffect(() => {
        if (!element) {
            return;
        }
        if (document.activeElement === element) {
            setActive(true);
        }
        element.addEventListener('focusin', handleFocusIn);
        element.addEventListener('focusout', handleFocusOut);
        return () => {
            document.removeEventListener('focusin', handleFocusIn);
            document.removeEventListener('focusout', handleFocusOut);
        };
    }, [element]);

    return active;
};

export function useOnScreen(ref: RefObject<HTMLElement>) {
    const [isIntersecting, setIntersecting] = useState(false);

    const observer = useMemo(() => new IntersectionObserver(([entry]) => setIntersecting(entry.isIntersecting)), [ref]);

    useEffect(() => {
        observer.observe(ref.current);
        return () => observer.disconnect();
    }, []);

    return isIntersecting;
}
