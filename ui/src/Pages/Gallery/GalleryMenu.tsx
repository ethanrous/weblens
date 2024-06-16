import WeblensMedia from '../../Media/Media'
import { WeblensButton } from '../../components/WeblensButton'
import { useClick, useKeyDown } from '../../components/hooks'
import { memo, useCallback, useContext, useMemo, useState } from 'react'
import { UserContext } from '../../Context'
import {
    IconCalendarTime,
    IconCaretDown,
    IconCaretUp,
    IconEyeOff,
    IconPolaroid,
} from '@tabler/icons-react'
import { GalleryContext } from './Gallery'
import { adjustMediaTime, hideMedia } from '../../api/ApiFetch'
import { GalleryDispatchT, newTimeOffset, TimeOffset } from '../../types/Types'
import { SetAlbumCover } from '../../Albums/AlbumQuery'

const mediaDate = (timestamp: number) => {
    const dateObj = new Date(timestamp)
    const options: Intl.DateTimeFormatOptions = {
        month: 'long',
        day: 'numeric',
        minute: 'numeric',
        hour: 'numeric',
        timeZoneName: 'short',
    }
    if (dateObj.getFullYear() !== new Date().getFullYear()) {
        options.year = 'numeric'
    }
    return dateObj.toLocaleDateString('en-US', options)
}

function TimeSlice({
    value,
    isAnchor,
    incDecFunc,
}: {
    value: any
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
    adjustTime: (d: Date) => Promise<string>
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
                        value={offsetDate.getDate()}
                        isAnchor={isAnchor}
                        incDecFunc={dayTicker}
                    />

                    {/* Year */}
                    <TimeSlice
                        value={offsetDate.getFullYear()}
                        isAnchor={isAnchor}
                        incDecFunc={yearTicker}
                    />
                </div>
                <div className="p-3" />
                <div className="flex flex-row items-center justify-center">
                    {/* Hour */}
                    <TimeSlice
                        value={offsetDate.getHours()}
                        isAnchor={isAnchor}
                        incDecFunc={hourTicker}
                    />
                    <p className="select-none">:</p>
                    {/* Minute */}
                    <TimeSlice
                        value={offsetDate.getMinutes()}
                        isAnchor={isAnchor}
                        incDecFunc={minuteTicker}
                    />
                    <p className="select-none">:</p>
                    {/* Second */}
                    <TimeSlice
                        value={offsetDate.getSeconds()}
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
                            adjustTime(offsetDate)
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
        albumId,
    }: {
        media: WeblensMedia
        open: boolean
        setOpen: (o: boolean) => void
        updateAlbum?: () => void
        albumId?: string
    }) => {
        const { authHeader } = useContext(UserContext)
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

        useKeyDown('Escape', (e) => {
            if (open) {
                setOpen(false)
                galleryDispatch({
                    type: 'set_time_offset',
                    offset: null,
                })
            }
        })

        const hide = useCallback(
            async (e) => {
                e.stopPropagation()

                let medias = []
                if (galleryState.selecting) {
                    medias = Array.from(galleryState.selected.keys())
                }
                medias.push(media.Id())
                const r = hideMedia(medias, authHeader)

                galleryDispatch({
                    type: 'delete_from_map',
                    mediaIds: medias,
                })

                if ((await r).status !== 200) {
                    return false
                }

                return true
            },
            [galleryState.selected]
        )

        const adjustTime = useCallback(
            async (newDate: Date) => {
                const r = await adjustMediaTime(
                    media.Id(),
                    newDate,
                    Array.from(galleryState.selected.keys()),
                    authHeader
                )
                galleryDispatch({
                    type: 'set_time_offset',
                    offset: null,
                })

                return r
            },
            [media, authHeader, galleryState.selected]
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
                            disabled={!albumId}
                            style={{ opacity: open ? '100%' : '0%' }}
                            onClick={async (e) => {
                                e.stopPropagation()
                                const r = await SetAlbumCover(
                                    albumId,
                                    media.Id(),
                                    authHeader
                                )
                                updateAlbum()
                                return r.status === 200
                            }}
                        />
                        <WeblensButton
                            label="Hide"
                            centerContent
                            danger
                            fillWidth
                            Left={IconEyeOff}
                            squareSize={40}
                            textMin={100}
                            disabled={media.IsHidden()}
                            style={hideStyle}
                            onClick={hide}
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
        if (prev.albumId !== next.albumId) {
            return false
        }
        if (prev.updateAlbum !== next.updateAlbum) {
            return false
        }

        return true
    }
)
