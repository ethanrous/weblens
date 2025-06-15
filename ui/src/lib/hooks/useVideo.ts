import { useCallback, useEffect, useState } from 'react'

export function useVideo(elem: HTMLVideoElement | null) {
    const [playtime, setPlaytime] = useState(0)
    const [isPlaying, setIsPlaying] = useState(false)
    const [isWaiting, setIsWaiting] = useState(true)

    console.log('Using video')

    const updatePlaytime = useCallback(() => {
        if (!elem) {
            return
        }

        setPlaytime(elem.currentTime)
    }, [elem])

    const updatePlayState = (e: Event) => {
        if (e.type === 'canplaythrough') {
            if (isWaiting) {
                setIsWaiting(false)
            }
            return
        } else if (e.type === 'play') {
            setIsPlaying(true)
        } else if (e.type === 'pause') {
            setIsPlaying(false)
        } else {
            console.error('unknown play event', e.type)
        }
    }

    const updateBufferState = useCallback(
        (e: Event) => {
            if (e.type === 'waiting') {
                setIsWaiting(true)
            } else if (e.type === 'playing') {
                setIsWaiting(false)
            }
        },
        [setIsWaiting]
    )

    const error = useCallback((e: Event) => {
        console.log('video error', e)
    }, [])

    useEffect(() => {
        if (!elem) {
            return
        }

        elem.addEventListener('timeupdate', updatePlaytime)
        elem.addEventListener('play', updatePlayState)
        elem.addEventListener('pause', updatePlayState)
        // elem.addEventListener('canplay', updatePlayState)
        elem.addEventListener('canplaythrough', updatePlayState)
        elem.addEventListener('waiting', updateBufferState)
        elem.addEventListener('playing', updateBufferState)
        elem.addEventListener('error', error)
        return () => {
            elem.removeEventListener('timeupdate', updatePlaytime)
            elem.removeEventListener('play', updatePlayState)
            elem.removeEventListener('pause', updatePlayState)
            // elem.removeEventListener('canplay', updatePlayState)
            elem.removeEventListener('canplaythrough', updatePlayState)
            elem.removeEventListener('waiting', updateBufferState)
            elem.removeEventListener('playing', updateBufferState)
            elem.removeEventListener('error', error)
        }
    }, [updatePlaytime, updatePlayState, elem])

    return { playtime, isPlaying, isWaiting }
}
