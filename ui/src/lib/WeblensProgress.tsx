import { useResize } from '@weblens/lib/hooks'
import progressStyle from '@weblens/lib/weblensProgress.module.scss'
import { clamp } from '@weblens/util'
import { CSSProperties, useEffect, useRef, useState } from 'react'

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
    className?: string
}

function WeblensProgress({
    value,
    secondaryValue,
    complete = false,
    orientation = 'horizontal',
    loading = false,
    disabled = false,
    failure = false,
    seekCallback,
    primaryColor,
    secondaryColor,
    className,
}: progressProps) {
    const [dragging, setDragging] = useState(false)
    const [percentage, setPercentage] = useState(clamp(value, 0, 100))
    const [hoverPercent, setHoverPercent] = useState(clamp(value, 0, 100))
    const boxRef = useRef<HTMLDivElement>(null)
    const size = useResize(boxRef)

    useEffect(() => {
        if (seekCallback) {
            seekCallback(percentage, dragging)
        }
    }, [percentage, dragging, seekCallback])

    useEffect(() => {
        if (dragging) {
            if (!boxRef.current) {
                return
            }

            const rect = boxRef.current.getBoundingClientRect()

            const update = (e: MouseEvent) => {
                setPercentage(
                    clamp(
                        ((e.clientX - rect.left) / (rect.right - rect.left)) *
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

    const style = {
        '--color-progress-bar-primary': '',
        '--color-progress-bar-secondary': '',
    }

    if (primaryColor) {
        style['--color-progress-bar-primary'] = primaryColor
    }

    if (secondaryColor) {
        style['--color-progress-bar-secondary'] = secondaryColor
    }

    return (
        <div
            className={
                'pointer-events-auto relative flex h-3 items-center ' +
                className
            }
            data-scrubbing={dragging}
            style={style as CSSProperties}
            data-seekable={seekCallback !== undefined}
        >
            {seekCallback !== undefined && (
                <div
                    className={progressStyle.sliderHandle}
                    style={{
                        left: `${clamp((value / 100) * size.width, 6, size.width - 6)}px`,
                    }}
                    onClick={(e) => e.stopPropagation()}
                    onMouseDown={() => setDragging(true)}
                    onMouseUp={() => setDragging(false)}
                />
            )}
            <div
                className={progressStyle.weblensProgress}
                ref={boxRef}
                data-loading={loading}
                data-disabled={disabled}
                data-complete={complete}
                data-failure={failure}
                onMouseUp={() => setDragging(false)}
                onMouseMove={(e) => {
                    const rect = (
                        e.target as HTMLDivElement
                    ).getBoundingClientRect()
                    let v = (e.clientX - rect.left) / (rect.right - rect.left)
                    if (v < 0) {
                        v = 0
                    }
                    setHoverPercent(v * 100)
                }}
                onMouseLeave={() => setHoverPercent(0)}
                onMouseDown={(e) => {
                    e.stopPropagation()
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
                    className="bg-progress-bar-primary pointer-events-none relative z-20 h-full shadow-lg transition-colors"
                    data-complete={complete}
                    style={{
                        height: orientation === 'horizontal' ? '' : `${value}%`,
                        width: orientation === 'horizontal' ? `${value}%` : '',
                    }}
                />
                {seekCallback !== undefined && (
                    <div
                        className="bg-progress-bar-primary/50 pointer-events-none absolute z-20 h-full shadow-lg transition-colors"
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
                        }}
                    />
                )}
                <div
                    className="bg-progress-bar-secondary pointer-events-none absolute z-10 h-full shadow-lg transition-colors"
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
                    }}
                />
            </div>
        </div>
    )
}
export default WeblensProgress
