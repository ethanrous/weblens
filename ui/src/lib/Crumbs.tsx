import {
    IconChevronRight,
    IconCornerDownRight,
    IconHome,
    IconTrash,
    IconUsers,
} from '@tabler/icons-react'
import { useSessionStore } from '@weblens/components/UserInfo'
import { useClick, useResize } from '@weblens/components/hooks'
import { filenameFromPath } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { ShareRoot, useFileBrowserStore } from '@weblens/store/FBStateControl'
import { DraggingStateT } from '@weblens/types/files/FBTypes'
import { WeblensFile } from '@weblens/types/files/File'
import { goToFile } from '@weblens/types/files/FileDragLogic'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import WeblensButton from './WeblensButton'
import crumbStyle from './crumbStyle.module.scss'

type Crumb = {
    path: string
    id: string
    file?: WeblensFile
    visitRoute?: string
    navigable: boolean
}

type breadcrumbProps = {
    crumbInfo: Crumb
    moveSelectedTo: (folderId: string) => void
    isCurrent: boolean
    compact?: boolean
}

const CrumbText = ({ crumb }: { crumb: Crumb }) => {
    const usr = useSessionStore((state) => state.user)

    if (crumb.id === usr.homeId) {
        return (
            <div className={crumbStyle['crumb-icon']}>
                <IconHome className="w-full h-full" />
            </div>
        )
    }
    if (crumb.id === 'shared') {
        return (
            <div className={crumbStyle['crumb-icon']}>
                <IconUsers className="w-full h-full" />
            </div>
        )
    }
    if (crumb.id === usr.trashId) {
        return (
            <div className={crumbStyle['crumb-icon']}>
                <IconTrash className="w-full h-full" />
            </div>
        )
    }

    const { nameText } = filenameFromPath(crumb.path)

    return (
        <p className={`${crumbStyle['crumb-text']} w-max max-w-full`}>
            {nameText}
        </p>
    )
}

export const StyledBreadcrumb = ({
    crumbInfo,
    compact = false,
    moveSelectedTo,
    isCurrent,
}: breadcrumbProps) => {
    const dragging = useFileBrowserStore((state) => state.draggingState)

    const setMoveDest = useFileBrowserStore((state) => state.setMoveDest)
    const clearSelected = useFileBrowserStore((state) => state.clearSelected)
    const nav = useNavigate()

    const [copyOpen, setCopyOpen] = useState(false)
    const [buttonRef, setButtonRef] = useState<HTMLDivElement>()
    useClick(() => setCopyOpen(false), buttonRef, !copyOpen)

    if (!crumbInfo) {
        console.error('NO CRUMB')
        return null
    }

    return (
        <div
            className={crumbStyle['crumb-box']}
            data-navigable={crumbInfo.navigable}
            data-compact={compact}
            data-dragging={dragging === DraggingStateT.InternalDrag}
            data-current={isCurrent}
            onContextMenu={(e) => {
                e.stopPropagation()
                e.preventDefault()
                setCopyOpen((c) => !c)
            }}
            onMouseOver={() => {
                if (dragging && !isCurrent && setMoveDest) {
                    setMoveDest(crumbInfo.path)
                }
            }}
            onMouseLeave={() => {
                if (dragging && setMoveDest) {
                    setMoveDest('')
                }
            }}
            onClick={(e) => {
                e.stopPropagation()
                if (dragging === DraggingStateT.NoDrag) {
                    if (crumbInfo.id === 'shared') {
                        goToFile(ShareRoot, true)
                    } else if (crumbInfo.file) {
                        clearSelected()
                        goToFile(crumbInfo.file)
                    } else if (crumbInfo.visitRoute) {
                        nav(crumbInfo.visitRoute)
                    } else {
                        console.error('No file or visit route in crumb')
                    }
                }
            }}
            onMouseUp={(e) => {
                e.stopPropagation()
                if (!crumbInfo.navigable) {
                    return
                }

                if (dragging !== DraggingStateT.NoDrag) {
                    moveSelectedTo(crumbInfo.id)
                }
                setMoveDest('')
            }}
        >
            <CrumbText crumb={crumbInfo} />
            {copyOpen && (
                <div ref={setButtonRef} className="absolute z-20">
                    <WeblensButton
                        label="Copy"
                        onClick={(e) => {
                            e.stopPropagation()
                        }}
                    />
                </div>
            )}
        </div>
    )
}

function LoafOverflowMenu({
    open,
    reff,
    setOpen,
    crumbs,
}: {
    open: boolean
    reff: HTMLDivElement
    setOpen: (b: boolean) => void
    crumbs: Crumb[]
}) {
    useClick(() => setOpen(false), reff, !open)
    return (
        <div className={crumbStyle['overflow-menu']} data-open={open}>
            {crumbs.map((item, i: number) => {
                return (
                    <div
                        key={`crumb-overflow-${i}`}
                        className="flex flex-row items-center min-w-0"
                        style={{
                            paddingLeft: 12 + (i - 1) * 28,
                        }}
                    >
                        {i !== 0 && <IconCornerDownRight />}
                        <StyledBreadcrumb
                            crumbInfo={item}
                            moveSelectedTo={() => {}}
                            isCurrent={i === crumbs.length - 1}
                        />
                    </div>
                )
            })}
        </div>
    )
}

export const StyledLoaf = ({
    crumbs,
    moveSelectedTo,
}: {
    crumbs: Crumb[]
    moveSelectedTo: (folderId: string) => void
}) => {
    const [widths, setWidths] = useState<number[]>(new Array(crumbs.length))
    // const [squished, setSquished] = useState(0)
    const [crumbsRef, setCrumbRef] = useState<HTMLDivElement>(null)
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
        <div ref={setCrumbRef} className={crumbStyle['loaf']}>
            {crumbs.map((c, i) => (
                <div
                    key={c.id}
                    className="flex flex-row items-center min-w-[50px]"
                >
                    <StyledBreadcrumb
                        crumbInfo={c}
                        moveSelectedTo={moveSelectedTo}
                        isCurrent={i === crumbs.length - 1}
                    />
                    {i !== crumbs.length - 1 && (
                        <IconChevronRight
                            style={{ width: '20px', minWidth: '20px' }}
                        />
                    )}
                </div>
            ))}

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
                        <p className={crumbStyle['crumb-text'] + 'trun'}>...</p>
                        <LoafOverflowMenu
                            open={overflowMenu}
                            reff={overflowRef}
                            setOpen={setOverflowMenu}
                            crumbs={crumbs}
                        />
                    </div>
                </div>
            )}
        </div>
    )
}

function Crumbs({
    moveSelectedTo,
    navOnLast,
}: {
    navOnLast: boolean
    moveSelectedTo?: (folderId: string) => void
}) {
    const user = useSessionStore((state) => state.user)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const crumbs: Crumb[] = []

    if (!user || !folderInfo?.Id()) {
        return <StyledLoaf crumbs={crumbs} moveSelectedTo={moveSelectedTo} />
    }

    crumbs.push(
        ...folderInfo.FormatParents().map((parent) => {
            return {
                path: parent.portablePath,
                id: parent.id,
                file: parent,
                navigable: true,
            }
        })
    )

    if (folderInfo.IsInTrash()) {
        crumbs.shift()
    }

    console.log(folderInfo.GetFilename())
    // Add the current folder, which is not always navigable
    crumbs.push({
        path: folderInfo.GetFilename(),
        id: folderInfo.Id(),
        file: folderInfo,
        navigable: navOnLast,
    })

    return <StyledLoaf crumbs={crumbs} moveSelectedTo={moveSelectedTo} />
}

export default Crumbs
