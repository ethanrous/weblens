import { useNavigate } from 'react-router-dom'

import { Text } from '@mantine/core'
import {
    IconChevronRight,
    IconHome,
    IconTrash,
    IconUsers,
} from '@tabler/icons-react'
import { memo, MouseEventHandler, useContext, useEffect, useState } from 'react'
import { UserContextT } from '../types/Types'
import { UserContext } from '../Context'
import { useResize } from './hooks'
import { FbContext, FbModeT } from '../Pages/FileBrowser/FileBrowser'

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
    alwaysOn = false,
    fontSize = 25,
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

export const StyledLoaf = ({ crumbs, postText }) => {
    const [widths, setWidths] = useState(new Array(crumbs.length))
    const [squished, setSquished] = useState(0)
    const [crumbsRef, setCrumbRef] = useState(null)
    const size = useResize(crumbsRef)

    useEffect(() => {
        if (widths.length !== crumbs.length) {
            const newWidths = [...widths.slice(0, crumbs.length)]
            setWidths(newWidths)
        }
    }, [crumbs.length])

    useEffect(() => {
        if (!widths || widths[0] == undefined) {
            return
        }
        let total = widths.reduce((acc, v) => acc + v)
        let squishCount

        // - 20 to account for width of ... text
        for (squishCount = 0; total > size.width - 20; squishCount++) {
            total -= widths[squishCount]
        }
        setSquished(squishCount)
    }, [size, widths])

    return (
        <div ref={setCrumbRef} className="loaf">
            <div ref={setCrumbRef} className="flex items-center w-max">
                {crumbs[0]}
            </div>

            {squished !== 0 && (
                <IconChevronRight style={{ width: '20px', minWidth: '20px' }} />
            )}
            {squished !== 0 && <Text className="crumb-text">...</Text>}

            {crumbs.slice(1).map((c, i) => (
                <Crumbcatenator
                    key={i}
                    crumb={c}
                    index={i}
                    squished={squished}
                    setWidth={(index: string | number, width: any) =>
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
        const { usr }: UserContextT = useContext(UserContext)
        const { fbState, fbDispatch } = useContext(FbContext)

        let crumbs = []
        if (
            fbState.fbMode === FbModeT.share &&
            fbState.folderInfo.GetOwner() !== usr.username
        ) {
            crumbs.push(
                <StyledBreadcrumb
                    key={'shared-crumb'}
                    label={'Shared'}
                    onClick={(e) => {
                        e.stopPropagation()
                        nav('/files/shared')
                    }}
                    dragging={dragging}
                    onMouseUp={(e) => {
                        e.stopPropagation()
                    }}
                    setMoveDest={setMoveDest}
                    isCurrent={!fbState.contentId}
                />
            )
        }

        if (!usr || !fbState.folderInfo?.Id()) {
            return <StyledLoaf crumbs={crumbs} postText={''} />
        }

        if (!fbState.folderInfo.IsTrash()) {
            const parents = fbState.folderInfo.FormatParents().map((parent) => {
                return (
                    <StyledBreadcrumb
                        key={parent.Id()}
                        label={parent.GetFilename()}
                        onClick={(e) => {
                            e.stopPropagation()
                            const route = parent.GetVisitRoute(
                                fbState.fbMode,
                                fbState.shareId,
                                fbDispatch
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
                key={fbState.folderInfo.Id()}
                label={fbState.folderInfo.GetFilename()}
                onClick={(e) => {
                    e.stopPropagation()
                    if (!navOnLast) {
                        return
                    }
                    const route = fbState.folderInfo.GetVisitRoute(
                        fbState.fbMode,
                        fbState.shareId,
                        fbDispatch
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
        return !(prev.dragging !== next.dragging)
    }
)

export default Crumbs
