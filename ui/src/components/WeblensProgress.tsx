import { CSSProperties } from '@mantine/core'
import { memo } from 'react'
import '../style/weblensProgress.scss'

type progressProps = {
    value: number
    complete?: boolean
    orientation?: 'horizontal' | 'vertical'
    loading?: boolean
    failure?: boolean
    seekCallback?: (v: number) => void
    style?: CSSProperties
}

export const WeblensProgress = memo(
    ({
        value,
        complete = false,
        orientation = 'horizontal',
        loading = false,
        failure = false,
        seekCallback,
        style,
    }: progressProps) => {
        return (
            <div
                className="weblens-progress"
                data-loading={loading}
                data-complete={complete}
                data-failure={failure}
                onClick={(e) => {
                    if (!seekCallback) {
                        return
                    }

                    if (e.target instanceof HTMLDivElement) {
                        const rect = e.target.getBoundingClientRect()
                        console.log(e.clientX, rect.left, rect.right)
                        seekCallback(
                            (e.clientX - rect.left) / (rect.right - rect.left)
                        )
                    }
                }}
                style={{
                    justifyContent:
                        orientation === 'horizontal'
                            ? 'flex-start'
                            : 'flex-end',
                    flexDirection:
                        orientation === 'horizontal' ? 'row' : 'column',
                    ...style,
                    cursor: seekCallback ? 'pointer' : 'default',
                }}
            >
                <div
                    className="weblens-progress-bar"
                    data-complete={complete}
                    style={{
                        height:
                            orientation === 'horizontal' ? '100%' : `${value}%`,
                        width:
                            orientation === 'horizontal' ? `${value}%` : '100%',
                    }}
                />
            </div>
        )
    },
    (prev, next) => {
        if (prev.value !== next.value) {
            return false
        }
        if (prev.complete !== next.complete) {
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
