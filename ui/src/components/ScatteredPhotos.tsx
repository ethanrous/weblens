import { useResizeWindow } from '@weblens/lib/hooks'
import WeblensMedia, { PhotoQuality } from '@weblens/types/media/Media'
import { GetMediaRows } from '@weblens/types/media/MediaRows'
import { MediaImage } from '@weblens/types/media/PhotoContainer'
import { gsap } from 'gsap'
import { FC, useEffect, useMemo, useRef } from 'react'

gsap.registerPlugin()

interface GridMotionProps {
    items?: (string | WeblensMedia)[]
    gradientColor?: string
}

const GridMotion: FC<GridMotionProps> = ({
    items = [],
    gradientColor = 'black',
}) => {
    const gridRef = useRef<HTMLDivElement>(null)
    const rowRefs = useRef<(HTMLDivElement | null)[]>([])
    const mouseXRef = useRef<number>(window.innerWidth / 2)
    const windowSise = useResizeWindow()

    const mediaRows = useMemo(() => {
        if (!(items[0] instanceof WeblensMedia)) {
            return
        }

        return GetMediaRows(
            items as WeblensMedia[],
            380,
            windowSise.width * 1.5,
            4
        )
    }, [items, windowSise.width])

    console.log('mediaRows', mediaRows)

    useEffect(() => {
        gsap.ticker.lagSmoothing(0)

        const handleMouseMove = (e: MouseEvent): void => {
            mouseXRef.current = e.clientX
        }

        const updateMotion = (): void => {
            const maxMoveAmount = 300
            const baseDuration = 0.8 // Base duration for inertia
            const inertiaFactors = [0.6, 0.4, 0.3, 0.2] // Different inertia for each row, outer rows slower

            rowRefs.current.forEach((row, index) => {
                if (row) {
                    const direction = index % 2 === 0 ? 1 : -1
                    const moveAmount =
                        ((mouseXRef.current / window.innerWidth) *
                            maxMoveAmount -
                            maxMoveAmount / 2) *
                        direction

                    gsap.to(row, {
                        x: moveAmount,
                        duration:
                            baseDuration +
                            inertiaFactors[index % inertiaFactors.length],
                        ease: 'power3.out',
                        overwrite: 'auto',
                    })
                }
            })
        }

        const removeAnimationLoop = gsap.ticker.add(updateMotion)
        window.addEventListener('mousemove', handleMouseMove)

        return () => {
            window.removeEventListener('mousemove', handleMouseMove)
            removeAnimationLoop()
        }
    }, [])

    return (
        <div ref={gridRef} className="h-full w-full overflow-hidden">
            <section
                className="relative flex h-screen w-full items-center justify-center overflow-hidden"
                style={{
                    background: `radial-gradient(circle, ${gradientColor} 0%, transparent 100%)`,
                }}
            >
                {/* <div className="pointer-events-none absolute inset-0 z-[4] bg-[length:250px]"></div> */}
                <div className="relative z-[2] flex h-[150vh] w-[150vw] origin-center rotate-[-15deg] flex-col justify-center gap-4">
                    {mediaRows &&
                        mediaRows.map((row, rowIndex) => (
                            <div
                                key={rowIndex}
                                className="relative flex gap-4"
                                style={{
                                    willChange: 'transform, filter',
                                    gridTemplateColumns: `repeat(${row.items.length}, minmax(0, 1fr))`,
                                    height: row.rowHeight,
                                }}
                                ref={(el) => {
                                    rowRefs.current[rowIndex] = el
                                }}
                            >
                                {row.items.map((rowItem) => {
                                    return (
                                        <MediaImage
                                            key={rowItem.m.Id()}
                                            media={rowItem.m}
                                            quality={PhotoQuality.LowRes}
                                            containerStyle={{
                                                width: rowItem.w,
                                            }}
                                        />
                                    )
                                })}
                            </div>
                        ))}
                </div>
                <div className="pointer-events-none relative top-0 left-0 h-full w-full"></div>
            </section>
        </div>
    )
}

export default GridMotion
