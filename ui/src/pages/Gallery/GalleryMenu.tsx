import {
    IconCalendarTime,
    IconCaretDown,
    IconCaretUp,
    IconEyeOff,
} from '@tabler/icons-react'
import MediaApi from '@weblens/api/MediaApi'
import { useClick, useKeyDown } from '@weblens/lib/hooks'
import WeblensButton from '@weblens/lib/WeblensButton'
import { TimeOffset, newTimeOffset } from '@weblens/types/Types'
import WeblensMedia from '@weblens/types/media/Media'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { MouseEvent, memo, useCallback, useMemo, useState } from 'react'

import { useGalleryStore } from './GalleryLogic'

function TimeSlice({
    value,
    isAnchor,
    incDecFunc,
}: {
    value: string
    isAnchor: boolean
    incDecFunc: (n: number) => void
}) {
    return (
        <div className="flex h-max w-max flex-col items-center p-2">
            {isAnchor && (
                <IconCaretUp
                    className="cursor-pointer"
                    onClick={(e) => {
                        e.stopPropagation()
                        incDecFunc(1)
                    }}
                />
            )}
            <div className="flex h-max w-max select-none flex-col rounded-sm bg-[#00000077] p-2 text-white">
                <p>{value}</p>
            </div>
            {isAnchor && (
                <IconCaretDown
                    className="cursor-pointer"
                    onClick={(e) => {
                        e.stopPropagation()
                        incDecFunc(-1)
                    }}
                />
            )}
        </div>
    )
}

function TimeDialogue({
    media,
    isAnchor,
    close,
    offset,
    adjustTime,
}: {
    media: WeblensMedia
    isAnchor: boolean
    close: () => void
    offset: TimeOffset
    adjustTime: (d: Date) => Promise<boolean>
}) {
    const [date, setDate] = useState<Date>(null)
    const setTimeOffset = useGalleryStore((state) => state.setTimeOffset)

    const monthTicker = useCallback(
        (n: number) => {
            offset.month += n
            setTimeOffset(offset)
        },
        [offset]
    )

    const dayTicker = useCallback(
        (n: number) => {
            offset.day += n
            setTimeOffset(offset)
        },
        [offset]
    )

    const yearTicker = useCallback(
        (n: number) => {
            offset.year += n
            setTimeOffset(offset)
        },
        [offset]
    )
    const hourTicker = useCallback(
        (n: number) => {
            offset.hour += n
            setTimeOffset(offset)
        },
        [offset]
    )
    const minuteTicker = useCallback(
        (n: number) => {
            offset.minute += n
            setTimeOffset(offset)
        },
        [offset]
    )
    const secondTicker = useCallback(
        (n: number) => {
            offset.second += n
            setTimeOffset(offset)
        },
        [offset]
    )

    const offsetDate = useMemo(() => {
        if (date == null || !offset) {
            return null
        }
        let offsetDate = new Date(date)

        offsetDate = new Date(
            offsetDate.setSeconds(date.getSeconds() + offset.second)
        )
        offsetDate = new Date(
            offsetDate.setMinutes(date.getMinutes() + offset.minute)
        )
        offsetDate = new Date(
            offsetDate.setHours(date.getHours() + offset.hour)
        )
        offsetDate = new Date(offsetDate.setDate(date.getDate() + offset.day))

        // Months are indexed from 0, for some reason, so we must add 1
        offsetDate = new Date(
            offsetDate.setMonth(date.getMonth() + offset.month)
        )

        offsetDate = new Date(
            offsetDate.setFullYear(date.getFullYear() + offset.year)
        )

        return offsetDate
    }, [offset, date])

    if (date === null) {
        setDate(new Date(media.GetCreateDate()))
        return null
    }

    if (offsetDate == null) {
        return null
    }

    return (
        <div className="flex h-full flex-col items-center justify-evenly">
            <div className="flex w-full flex-col items-center justify-around p-1">
                <div className="flex flex-row items-center justify-center">
                    {/* Month */}
                    <TimeSlice
                        value={offsetDate.toLocaleString('default', {
                            month: 'short',
                        })}
                        isAnchor={isAnchor}
                        incDecFunc={monthTicker}
                    />

                    {/* Day */}
                    <TimeSlice
                        value={offsetDate.getDate().toString()}
                        isAnchor={isAnchor}
                        incDecFunc={dayTicker}
                    />

                    {/* Year */}
                    <TimeSlice
                        value={offsetDate.getFullYear().toString()}
                        isAnchor={isAnchor}
                        incDecFunc={yearTicker}
                    />
                </div>
                <div className="p-3" />
                <div className="flex flex-row items-center justify-center">
                    {/* Hour */}
                    <TimeSlice
                        value={offsetDate.getHours().toString()}
                        isAnchor={isAnchor}
                        incDecFunc={hourTicker}
                    />
                    <p className="select-none">:</p>
                    {/* Minute */}
                    <TimeSlice
                        value={offsetDate.getMinutes().toString()}
                        isAnchor={isAnchor}
                        incDecFunc={minuteTicker}
                    />
                    <p className="select-none">:</p>
                    {/* Second */}
                    <TimeSlice
                        value={offsetDate.getSeconds().toString()}
                        isAnchor={isAnchor}
                        incDecFunc={secondTicker}
                    />
                </div>
            </div>
            {isAnchor && (
                <div className="flex flex-row items-center">
                    <WeblensButton
                        label="Cancel"
                        subtle
                        fontSize={'xl'}
                        squareSize={50}
                        onClick={(e) => {
                            e.stopPropagation()
                            close()
                        }}
                    />
                    <div className="p-3" />
                    <WeblensButton
                        label="Confirm"
                        fontSize={'xl'}
                        squareSize={50}
                        onClick={(e) => {
                            e.stopPropagation()
                            return adjustTime(offsetDate)
                        }}
                    />
                </div>
            )}
        </div>
    )
}

export const GalleryMenu = memo(
    ({
        media,
        open,
        setOpen,
    }: {
        media: WeblensMedia
        open: boolean
        setOpen: (o: boolean) => void
    }) => {
        const selecting = useGalleryStore((state) => state.selecting)
        const menuTargetId = useGalleryStore((state) => state.menuTargetId)
        // const albumId = useGalleryStore((state) => state.albumId)
        const timeAdjustOffset = useGalleryStore(
            (state) => state.timeAdjustOffset
        )
        const setSelecting = useGalleryStore((state) => state.setSelecting)
        const setTimeOffset = useGalleryStore((state) => state.setTimeOffset)
        const setMenuTarget = useGalleryStore((state) => state.setMenuTarget)

        const [menuRef, setMenuRef] = useState<HTMLDivElement>(null)

        useClick(
            () => {
                setOpen(false)
                setTimeOffset(null)
            },
            menuRef,
            !open || timeAdjustOffset !== null
        )

        useKeyDown('Escape', () => {
            if (open) {
                setOpen(false)
                setTimeOffset(null)
            }
        })

        const hide = useCallback(
            async (e: MouseEvent, hidden: boolean) => {
                e.stopPropagation()

                let medias: string[] = []
                if (selecting) {
                    medias = [...useMediaStore.getState().selectedMap.keys()]
                }
                medias.push(media.Id())
                const res = await MediaApi.setMediaVisibility(hidden, {
                    mediaIds: medias,
                })

                if (res.status !== 200) {
                    return false
                }

                useMediaStore.getState().hideMedias(medias, hidden)

                setSelecting(false)
                setMenuTarget('')

                return true
            },
            [selecting]
        )

        // const adjustTime = useCallback(async () => {
        //     console.error('adjust time not impl')
        //     // const r = await adjustMediaTime(
        //     //     media.Id(),
        //     //     newDate,
        //     //     mediaState.getAllSelectedIds(),
        //     //     auth
        //     // )
        //     // galleryDispatch({
        //     //     type: 'set_time_offset',
        //     //     offset: null,
        //     // })
        //     //
        //     // return r
        //     return false
        // }, [media])

        const hideStyle = useMemo(() => {
            return { opacity: open ? '100%' : '0%' }
        }, [open])

        return (
            <div
                ref={setMenuRef}
                className="media-menu-container"
                data-open={
                    open || (media.IsSelected() && timeAdjustOffset !== null)
                }
                onClick={(e) => {
                    e.stopPropagation()
                    setOpen(false)
                }}
                onContextMenu={(e) => {
                    e.stopPropagation()
                    e.preventDefault()
                    setOpen(false)
                }}
            >
                {(media.IsSelected() || open) && timeAdjustOffset !== null && (
                    <TimeDialogue
                        media={media}
                        isAnchor={media.Id() === menuTargetId}
                        close={() => setTimeOffset(null)}
                        adjustTime={(d: Date) =>
                            new Promise(() => {
                                console.error('Adjust time not impl', d)
                                return false
                            })
                        }
                        offset={timeAdjustOffset}
                    />
                )}
                {open && timeAdjustOffset === null && (
                    <div className="flex w-full max-w-[300px] flex-col items-center p-2">
                        <WeblensButton
                            squareSize={40}
                            fillWidth
                            Left={IconCalendarTime}
                            label="Adjust Time"
                            centerContent
                            onClick={(e) => {
                                e.stopPropagation()
                                setTimeOffset(newTimeOffset())
                            }}
                        />
                        {/* <WeblensButton */}
                        {/*     label="Set as Cover" */}
                        {/*     centerContent */}
                        {/*     fillWidth */}
                        {/*     Left={IconPolaroid} */}
                        {/*     squareSize={40} */}
                        {/*     textMin={100} */}
                        {/*     disabled={!albumId} */}
                        {/*     style={{ opacity: open ? '100%' : '0%' }} */}
                        {/*     onClick={(e) => { */}
                        {/*         e.stopPropagation() */}
                        {/*         return AlbumsApi.updateAlbum( */}
                        {/*             albumId, */}
                        {/*             media.Id() */}
                        {/*         ) */}
                        {/*     }} */}
                        {/* /> */}
                        <WeblensButton
                            label={media.IsHidden() ? 'Unhide' : 'Hide'}
                            centerContent
                            danger
                            fillWidth
                            Left={IconEyeOff}
                            squareSize={40}
                            textMin={100}
                            style={hideStyle}
                            onClick={(e) => hide(e, !media.IsHidden())}
                        />
                    </div>
                )}
            </div>
        )
    },
    (prev, next) => {
        if (prev.media !== next.media) {
            return false
        }
        if (prev.open !== next.open) {
            return false
        }
        if (prev.setOpen !== next.setOpen) {
            return false
        }

        return true
    }
)
