import { CSSProperties } from '@mantine/core'
import { useResize } from '@weblens/components/hooks'
import progressStyle from '@weblens/lib/weblensProgress.module.scss'
import { clamp } from '@weblens/util'
import { memo, useEffect, useState } from 'react'

type progressProps = {
    value: number
    secondaryValue?: number
    height?: number
    complete?: boolean
    orientation?: 'horizontal' | 'vertical'
    loading?: boolean
    disabled?: boolean
    failure?: boolean
    seekCallback?: (v: number, seeking?: boolean) => void
    style?: CSSProperties
    primaryColor?: string
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
        primaryColor,
        secondaryColor,
    }: progressProps) => {
        const [dragging, setDragging] = useState(false)
        const [percentage, setPercentage] = useState(clamp(value, 0, 100))
        const [hoverPercent, setHoverPercent] = useState(clamp(value, 0, 100))
        const [boxRef, setBoxRef] = useState<HTMLDivElement>(null)
        const size = useResize(boxRef)

        useEffect(() => {
            if (seekCallback) {
                seekCallback(percentage, dragging)
            }
        }, [percentage, dragging])

        useEffect(() => {
            if (dragging) {
                const rect = boxRef.getBoundingClientRect()
                const update = (e: MouseEvent) => {
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
                className={progressStyle.weblensProgressContainer}
                data-scrubbing={dragging}
                style={{
                    height: orientation === 'vertical' ? '100%' : height,
                }}
                data-seekable={seekCallback !== undefined}
            >
                {seekCallback !== undefined && (
                    <div
                        className={progressStyle.sliderHandle}
                        style={{
                            left: `${clamp((value / 100) * size.width, 6, size.width - 6)}px`,
                            height: height,
                            width: height,
                        }}
                        onClick={(e) => e.stopPropagation()}
                        onMouseDown={() => setDragging(true)}
                        onMouseUp={() => setDragging(false)}
                    />
                )}
                <div
                    className={progressStyle.weblensProgress}
                    ref={setBoxRef}
                    data-loading={loading}
                    data-disabled={disabled}
                    data-complete={complete}
                    data-failure={failure}
                    onMouseUp={() => setDragging(false)}
                    onMouseMove={(e) => {
                        if (e.target instanceof HTMLDivElement) {
                            const rect = e.target.getBoundingClientRect()
                            let v =
                                (e.clientX - rect.left) /
                                (rect.right - rect.left)
                            if (v < 0) {
                                v = 0
                            }
                            setHoverPercent(v * 100)
                        }
                    }}
                    onMouseLeave={() => setHoverPercent(0)}
                    onMouseDown={(e) => {
                        e.stopPropagation()
                        // if (e.target instanceof HTMLDivElement) {
                        //     const rect = e.target.getBoundingClientRect()
                        //     let v =
                        //         (e.clientX - rect.left) /
                        //         (rect.right - rect.left)
                        //     if (v < 0) {
                        //         v = 0
                        //     }
                        // }
                        setPercentage(hoverPercent)
                        setDragging(true)
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
                        className={progressStyle.weblensProgressBar}
                        data-complete={complete}
                        style={{
                            height:
                                orientation === 'horizontal' ? '' : `${value}%`,
                            width:
                                orientation === 'horizontal' ? `${value}%` : '',
                            backgroundColor: primaryColor ? primaryColor : '',
                        }}
                    />
                    {seekCallback !== undefined && (
                        <div
                            className={progressStyle.weblensProgressBar}
                            data-seek-hint={true}
                            style={{
                                height:
                                    orientation === 'horizontal'
                                        ? ''
                                        : `${hoverPercent}%`,
                                width:
                                    orientation === 'horizontal'
                                        ? `${hoverPercent}%`
                                        : '',
                                // backgroundColor: primaryColor
                                //     ? primaryColor
                                //     : '',
                            }}
                        />
                    )}
                    <div
                        className={progressStyle.weblensProgressBar}
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
        if (prev.primaryColor !== next.primaryColor) {
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
