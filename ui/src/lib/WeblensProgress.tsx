import { CSSProperties } from '@mantine/core'
import { clamp } from '@weblens/util'
import { memo, useEffect, useState } from 'react'
import '@weblens/lib/weblensProgress.scss'

type progressProps = {
    value: number
    secondaryValue?: number
    height?: number
    complete?: boolean
    orientation?: 'horizontal' | 'vertical'
    loading?: boolean
    disabled?: boolean
    failure?: boolean
    seekCallback?: (v: number) => void
    style?: CSSProperties
    secondaryColor?: string
}

const WeblensProgress = memo(
    ({
        value,
        secondaryValue,
        height = 25,
        complete = false,
        orientation = 'horizontal',
        loading = false,
        disabled = false,
        failure = false,
        seekCallback,
        style,
        secondaryColor,
    }: progressProps) => {
        const [dragging, setDragging] = useState(false)
        const [percentage, setPercentage] = useState(clamp(value, 0, 100))
        const [secondaryPercentage, setSecondaryPercentage] = useState(
            clamp(secondaryValue, 0, 100)
        )
        const [boxRef, setBoxRef] = useState(null)

        useEffect(() => {
            if (seekCallback) {
                seekCallback(percentage)
            }
        }, [percentage])

        useEffect(() => {
            if (dragging) {
                const rect = boxRef.getBoundingClientRect()
                const update = (e) => {
                    setPercentage(
                        clamp(
                            ((e.clientX - rect.left) /
                                (rect.right - rect.left)) *
                                100,
                            0,
                            100
                        )
                    )
                }
                const stop = () => setDragging(false)
                window.addEventListener('mousemove', update)
                window.addEventListener('mouseup', stop)
                return () => {
                    window.removeEventListener('mousemove', update)
                    window.removeEventListener('mouseup', stop)
                }
            }
        }, [dragging])

        return (
            <div
                className="weblens-progress-container"
                data-scrubbing={dragging}
                data-seekable={seekCallback !== undefined}
                style={{
                    height: height,
                    cursor: seekCallback ? 'pointer' : 'default',
                }}
            >
                {seekCallback !== undefined && (
                    <div
                        className="slider-handle"
                        style={{
                            left: `${value}%`,
                            height: height,
                            width: height,
                        }}
                        onClick={(e) => e.stopPropagation()}
                        onMouseDown={() => setDragging(true)}
                        onMouseUp={() => setDragging(false)}
                    />
                )}
                <div
                    className="weblens-progress"
                    ref={setBoxRef}
                    data-loading={loading}
                    data-disabled={disabled}
                    data-complete={complete}
                    data-failure={failure}
                    onMouseUp={() => setDragging(false)}
                    onMouseDown={(e) => {
                        if (e.target instanceof HTMLDivElement) {
                            const rect = e.target.getBoundingClientRect()
                            let v =
                                (e.clientX - rect.left) /
                                (rect.right - rect.left)
                            if (v < 0) {
                                v = 0
                            }
                            setPercentage(v * 100)
                            setDragging(true)
                        }
                    }}
                    style={{
                        justifyContent:
                            orientation === 'horizontal'
                                ? 'flex-start'
                                : 'flex-end',
                        ...style,
                    }}
                >
                    <div
                        className="weblens-progress-bar"
                        data-complete={complete}
                        style={{
                            height:
                                orientation === 'horizontal' ? '' : `${value}%`,
                            width:
                                orientation === 'horizontal' ? `${value}%` : '',
                        }}
                    />
                    <div
                        className="weblens-progress-bar"
                        data-secondary={true}
                        style={{
                            height:
                                orientation === 'horizontal'
                                    ? ''
                                    : `${secondaryValue}%`,
                            width:
                                orientation === 'horizontal'
                                    ? `${secondaryValue}%`
                                    : '',
                            backgroundColor: secondaryColor
                                ? secondaryColor
                                : '',
                        }}
                    />
                </div>
            </div>
        )
    },
    (prev, next) => {
        if (prev.value !== next.value) {
            return false
        }
        if (prev.secondaryValue !== next.secondaryValue) {
            return false
        }
        if (prev.complete !== next.complete) {
            return false
        }
        if (prev.disabled !== next.disabled) {
            return false
        }
        if (prev.loading !== next.loading) {
            return false
        }
        if (prev.failure !== next.failure) {
            return false
        }
        if (prev.orientation !== next.orientation) {
            return false
        }
        return true
    }
)

export default WeblensProgress