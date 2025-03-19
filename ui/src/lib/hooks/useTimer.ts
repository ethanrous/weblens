import { useEffect, useRef, useState } from 'react'

export const useTimer = (startTime: Date, startPaused?: boolean) => {
    const [elapsedTime, setElapsedTime] = useState(
        startTime ? Date.now() - startTime.getTime() : 0
    )
    const [isRunning, setIsRunning] = useState(false)
    const countRef = useRef<NodeJS.Timeout>(null)

    const handleStart = () => {
        if (isRunning) {
            return
        }
        setIsRunning(true)
        setElapsedTime(0)
        const startNum = startTime ? startTime.getTime() : 0
        countRef.current = setInterval(() => {
            setElapsedTime(Date.now() - startNum)
        }, 27)
    }

    useEffect(() => {
        clearInterval(countRef.current)
        if (!startTime || startPaused === false) {
            return
        }
        handleStart()
    }, [startTime])

    const handlePause = () => {
        clearInterval(countRef.current)
        setIsRunning(false)
    }

    const handleReset = () => {
        clearInterval(countRef.current)
        setIsRunning(false)
        setElapsedTime(0)
    }

    return { elapsedTime, isRunning, handleStart, handlePause, handleReset }
}
