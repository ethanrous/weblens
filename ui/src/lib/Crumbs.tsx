import { Text } from '@mantine/core'
import {
    IconChevronRight,
    IconCornerDownRight,
    IconHome,
    IconTrash,
    IconUsers,
} from '@tabler/icons-react'
import { useClick, useResize } from '@weblens/components/hooks'

import '@weblens/lib/crumbs.scss'
import { useSessionStore } from '@weblens/components/UserInfo'
import { useFileBrowserStore } from '@weblens/pages/FileBrowser/FBStateControl'
import { memo, MouseEventHandler, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

type breadcrumbProps = {
    label: string
    onClick?: MouseEventHandler<HTMLDivElement>
    onMouseUp?: (e) => void
    dragging?: number
    fontSize?: number
    compact?: boolean
    alwaysOn?: boolean
    setMoveDest?
    isCurrent: boolean
}

const CrumbText = ({ label }) => {
    if (label === 'Home') {
        return (
            <div className="h-[30px] w-[30px]">
                <IconHome className="w-full h-full" />
            </div>
        )
    }
    if (label === 'Shared') {
        return (
            <div className="h-[30px] w-[30px]">
                <IconUsers className="w-full h-full" />
            </div>
        )
    }
    if (label === 'Trash') {
        return (
            <div className="h-[30px] w-[30px]">
                <IconTrash className="w-full h-full" />
            </div>
        )
    }

    return <p className="crumb-text w-max max-w-full">{label}</p>
}

export const StyledBreadcrumb = ({
    label,
    onClick,
    dragging,
    onMouseUp = () => {},
    compact = false,
    setMoveDest,
    isCurrent,
}: breadcrumbProps) => {
    return (
        <div
            className={'crumb-box'}
            data-compact={compact}
            data-dragging={dragging === 1}
            data-current={isCurrent}
            onMouseOver={() => {
                if (dragging && setMoveDest) {
                    setMoveDest(label)
                }
            }}
            onMouseLeave={() => {
                if (dragging && setMoveDest) {
                    setMoveDest('')
                }
            }}
            onMouseUp={(e) => {
                onMouseUp(e)
                setMoveDest('')
            }}
            onClick={onClick}
        >
            <CrumbText label={label} />
        </div>
    )
}

// The crumb concatenateor, the Crumbcatenator
const Crumbcatenator = ({ crumb, index, squished, setWidth }) => {
    const [crumbRef, setCrumbRef] = useState(null)
    const size = useResize(crumbRef)
    useEffect(() => {
        setWidth(index, size.width)
    }, [size])

    if (squished >= index + 1) {
        return null
    }

    return (
        <div className="flex items-center w-max" ref={setCrumbRef}>
            {size.width > 10 && (
                <IconChevronRight style={{ width: '20px', minWidth: '20px' }} />
            )}
            {crumb}
        </div>
    )
}

function LoafOverflowMenu({ open, reff, setOpen, crumbs }) {
    useClick(() => setOpen(false), reff, !open)
    return (
        <div className="overflow-menu" data-open={open}>
            {crumbs.map((item, i) => {
                return (
                    <div
                        key={`crumb-overflow-${i}`}
                        className="flex flex-row items-center "
                        style={{
                            paddingLeft: 12 + (i - 1) * 28,
                        }}
                    >
                        {i !== 0 && <IconCornerDownRight />}
                        {item}
                    </div>
                )
            })}
        </div>
    )
}

export const StyledLoaf = ({ crumbs, postText }) => {
    const [widths, setWidths] = useState(new Array(crumbs.length))
    // const [squished, setSquished] = useState(0)
    const [crumbsRef, setCrumbRef] = useState(null)
    const size = useResize(crumbsRef)
    const [overflowMenu, setOverflowMenu] = useState(false)
    const [overflowRef, setOverflowRef] = useState<HTMLDivElement>()

    useEffect(() => {
        setOverflowMenu(false)
        if (widths.length !== crumbs.length) {
            const newWidths = [...widths.slice(0, crumbs.length)]
            setWidths(newWidths)
        }
    }, [crumbs.length])

    let squished = 0
    if (!widths || widths[0] == undefined) {
        squished = -1
        // return squished
    } else {
        let total = widths.reduce((acc, v) => acc + v)

        // - 20 to account for width of "..." text
        for (squished = 0; total > size.width - 20; squished++) {
            total -= widths[squished]
        }
    }

    return (
        <div ref={setCrumbRef} className="loaf">
            <div className="flex items-center w-max">{crumbs[0]}</div>

            {squished > 0 && (
                <div
                    ref={setOverflowRef}
                    className="flex flex-row items-center h-max w-max cursor-pointer"
                    onClick={(e) => {
                        e.stopPropagation()
                        setOverflowMenu((o) => {
                            return !o
                        })
                    }}
                >
                    <IconChevronRight
                        style={{ width: '20px', minWidth: '20px' }}
                    />
                    <div>
                        <Text className="crumb-text">...</Text>
                        <LoafOverflowMenu
                            open={overflowMenu}
                            reff={overflowRef}
                            setOpen={setOverflowMenu}
                            crumbs={crumbs}
                        />
                    </div>
                </div>
            )}

            {crumbs.slice(1).map((c, i) => (
                <Crumbcatenator
                    key={i}
                    crumb={c}
                    index={i}
                    squished={squished}
                    setWidth={(index: string | number, width: number) =>
                        setWidths((p) => {
                            p[index] = width
                            return [...p]
                        })
                    }
                />
            ))}
            <p className="crumb-text ml-[20px] text-[#c4c4c4] text-xl">
                {postText}
            </p>
        </div>
    )
}

const Crumbs = memo(
    ({
        postText,
        moveSelectedTo,
        navOnLast,
        setMoveDest,
        dragging,
    }: {
        postText?: string
        navOnLast: boolean
        moveSelectedTo?: (folderId: string) => void
        setMoveDest?: (itemName: string) => void
        dragging?: number
    }) => {
        const nav = useNavigate()
        const user = useSessionStore((state) => state.user)

        const mode = useFileBrowserStore((state) => state.fbMode)
        const folderInfo = useFileBrowserStore((state) => state.folderInfo)
        const shareId = useFileBrowserStore((state) => state.shareId)

        const setPresentationTarget = useFileBrowserStore(
            (state) => state.setPresentationTarget
        )

        const crumbs = []

        if (!user || !folderInfo?.Id()) {
            return <StyledLoaf crumbs={crumbs} postText={''} />
        }

        if (!folderInfo.IsTrash()) {
            const parents = folderInfo.FormatParents().map((parent) => {
                return (
                    <StyledBreadcrumb
                        key={parent.Id()}
                        label={parent.GetFilename()}
                        onClick={(e) => {
                            e.stopPropagation()
                            const route = parent.GetVisitRoute(
                                mode,
                                shareId,
                                setPresentationTarget
                            )
                            nav(route)
                        }}
                        dragging={dragging}
                        onMouseUp={() => {
                            if (dragging !== 0) {
                                moveSelectedTo(parent.Id())
                            }
                        }}
                        setMoveDest={setMoveDest}
                        isCurrent={false}
                    />
                )
            })
            crumbs.push(...parents)
        }

        crumbs.push(
            <StyledBreadcrumb
                key={folderInfo.Id()}
                label={folderInfo.GetFilename()}
                onClick={(e) => {
                    e.stopPropagation()
                    if (!navOnLast) {
                        return
                    }
                    const route = folderInfo.GetVisitRoute(
                        mode,
                        shareId,
                        setPresentationTarget
                    )
                    nav(route)
                }}
                setMoveDest={setMoveDest}
                isCurrent={true}
            />
        )

        return <StyledLoaf crumbs={crumbs} postText={postText} />
    },
    (prev, next) => {
        if (prev.postText !== next.postText) {
            return false
        } else if (prev.dragging !== next.dragging) {
            return false
        }
        return true
    }
)

export default Crumbs
