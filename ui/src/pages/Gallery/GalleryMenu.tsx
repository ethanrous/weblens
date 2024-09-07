import {
    IconCalendarTime,
    IconCaretDown,
    IconCaretUp,
    IconEyeOff,
    IconPolaroid,
} from '@tabler/icons-react'
import { useClick, useKeyDown } from '@weblens/components/hooks'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensButton from '@weblens/lib/WeblensButton'
import { SetAlbumCover } from '@weblens/types/albums/AlbumQuery'
import WeblensMedia from '@weblens/types/media/Media'
import { hideMedia } from '@weblens/types/media/MediaQuery'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { memo, useCallback, useContext, useMemo, useState } from 'react'
import {
    GalleryDispatchT,
    newTimeOffset,
    TimeOffset,
} from '@weblens/types/Types'
import { GalleryContext } from './GalleryLogic'

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
        <div className="flex flex-col w-max h-max p-2 items-center">
            {isAnchor && (
                <IconCaretUp
                    className="cursor-pointer"
                    onClick={(e) => {
                        e.stopPropagation()
                        incDecFunc(1)
                    }}
                />
            )}
            <div className="flex flex-col w-max h-max p-2 rounded select-none bg-[#00000077] text-white">
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
    galleryDispatch,
}: {
    media: WeblensMedia
    isAnchor: boolean
    close: () => void
    offset: TimeOffset
    adjustTime: (d: Date) => Promise<boolean>
    galleryDispatch: GalleryDispatchT
}) {
    const [date, setDate] = useState<Date>(null)

    const monthTicker = useCallback(
        (n) => {
            offset.month += n
            galleryDispatch({ type: 'set_time_offset', offset: offset })
        },
        [offset, galleryDispatch]
    )

    const dayTicker = useCallback(
        (n) => {
            offset.day += n
            galleryDispatch({ type: 'set_time_offset', offset: offset })
        },
        [offset, galleryDispatch]
    )

    const yearTicker = useCallback(
        (n) => {
            offset.year += n
            galleryDispatch({ type: 'set_time_offset', offset: offset })
        },
        [offset, galleryDispatch]
    )
    const hourTicker = useCallback(
        (n) => {
            offset.hour += n
            galleryDispatch({ type: 'set_time_offset', offset: offset })
        },
        [offset, galleryDispatch]
    )
    const minuteTicker = useCallback(
        (n) => {
            offset.minute += n
            galleryDispatch({ type: 'set_time_offset', offset: offset })
        },
        [offset, galleryDispatch]
    )
    const secondTicker = useCallback(
        (n) => {
            offset.second += n
            galleryDispatch({ type: 'set_time_offset', offset: offset })
        },
        [offset, galleryDispatch]
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
        <div className="flex flex-col h-full items-center justify-evenly">
            <div className="flex flex-col w-full items-center justify-around p-1">
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
        updateAlbum,
    }: {
        media: WeblensMedia
        open: boolean
        setOpen: (o: boolean) => void
        updateAlbum?: () => void
    }) => {
        const auth = useSessionStore((state) => state.auth)

        const { galleryState, galleryDispatch } = useContext(GalleryContext)
        const [menuRef, setMenuRef] = useState(null)

        useClick(
            () => {
                setOpen(false)
                galleryDispatch({
                    type: 'set_time_offset',
                    offset: null,
                })
            },
            menuRef,
            !open || galleryState.timeAdjustOffset !== null
        )

        useKeyDown('Escape', () => {
            if (open) {
                setOpen(false)
                galleryDispatch({
                    type: 'set_time_offset',
                    offset: null,
                })
            }
        })

        const hide = useCallback(
            async (e, hidden: boolean) => {
                e.stopPropagation()

                let medias: string[] = []
                if (galleryState.selecting) {
                    medias = [...useMediaStore.getState().selectedMap.keys()]
                }
                medias.push(media.Id())
                const r = await hideMedia(medias, hidden, auth)

                if (r.status !== 200) {
                    return false
                }

                useMediaStore.getState().hideMedias(medias, hidden)

                galleryDispatch({
                    type: 'set_selecting',
                    selecting: false,
                })

                galleryDispatch({
                    type: 'set_menu_target',
                    targetId: '',
                })

                return true
            },
            [galleryState.selecting]
        )

        const adjustTime = useCallback(
            async (newDate: Date) => {
                console.error('adjust time not impl')
                // const r = await adjustMediaTime(
                //     media.Id(),
                //     newDate,
                //     mediaState.getAllSelectedIds(),
                //     auth
                // )
                // galleryDispatch({
                //     type: 'set_time_offset',
                //     offset: null,
                // })
                //
                // return r
                return false
            },
            [media, auth]
        )

        const hideStyle = useMemo(() => {
            return { opacity: open ? '100%' : '0%' }
        }, [open])

        return (
            <div
                ref={setMenuRef}
                className="media-menu-container"
                data-open={
                    open ||
                    (media.IsSelected() &&
                        galleryState.timeAdjustOffset !== null)
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
                {(media.IsSelected() || open) &&
                    galleryState.timeAdjustOffset !== null && (
                        <TimeDialogue
                            media={media}
                            isAnchor={media.Id() === galleryState.menuTargetId}
                            close={() =>
                                galleryDispatch({
                                    type: 'set_time_offset',
                                    offset: null,
                                })
                            }
                            adjustTime={adjustTime}
                            offset={galleryState.timeAdjustOffset}
                            galleryDispatch={galleryDispatch}
                        />
                    )}
                {open && galleryState.timeAdjustOffset === null && (
                    <div className="flex flex-col items-center w-full max-w-[300px] p-2">
                        <WeblensButton
                            squareSize={40}
                            fillWidth
                            Left={IconCalendarTime}
                            label="Adjust Time"
                            centerContent
                            onClick={(e) => {
                                e.stopPropagation()
                                galleryDispatch({
                                    type: 'set_time_offset',
                                    offset: newTimeOffset(),
                                })
                            }}
                        />
                        <WeblensButton
                            label="Set as Cover"
                            centerContent
                            fillWidth
                            Left={IconPolaroid}
                            squareSize={40}
                            textMin={100}
                            disabled={!galleryState.albumId}
                            style={{ opacity: open ? '100%' : '0%' }}
                            onClick={async (e) => {
                                e.stopPropagation()
                                const r = await SetAlbumCover(
                                    galleryState.albumId,
                                    media.Id(),
                                    auth
                                )
                                return r.status === 200
                            }}
                        />
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
        if (prev.updateAlbum !== next.updateAlbum) {
            return false
        }

        return true
    }
)
