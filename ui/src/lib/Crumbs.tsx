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
import {
    FbModeT,
    useFileBrowserStore,
} from '@weblens/pages/FileBrowser/FBStateControl'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { WeblensFile } from '@weblens/types/files/File'
import { goToFile } from '@weblens/types/files/FileDragLogic'

type Crumb = {
    name: string
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
            <div className="h-[30px] w-[30px]">
                <IconHome className="w-full h-full" />
            </div>
        )
    }
    if (crumb.name === 'Shared') {
        return (
            <div className="h-[30px] w-[30px]">
                <IconUsers className="w-full h-full" />
            </div>
        )
    }
    if (crumb.id === usr.trashId) {
        return (
            <div className="h-[30px] w-[30px]">
                <IconTrash className="w-full h-full" />
            </div>
        )
    }

    return <p className="crumb-text w-max max-w-full">{crumb.name}</p>
}

export const StyledBreadcrumb = ({
    crumbInfo,
    compact = false,
    moveSelectedTo,
    isCurrent,
}: breadcrumbProps) => {
    if (!crumbInfo) {
        console.error('NO CRUMB')
        return null
    }

    const setMoveDest = useFileBrowserStore((state) => state.setMoveDest)
    const dragging = useFileBrowserStore((state) => state.draggingState)
    const clearSelected = useFileBrowserStore((state) => state.clearSelected)
    const nav = useNavigate()

    return (
        <div
            className={'crumb-box'}
            data-navigable={crumbInfo.navigable}
            data-compact={compact}
            data-dragging={dragging === 1}
            data-current={isCurrent}
            onMouseOver={() => {
                if (dragging && !isCurrent && setMoveDest) {
                    setMoveDest(crumbInfo.name)
                }
            }}
            onMouseLeave={() => {
                if (dragging && setMoveDest) {
                    setMoveDest('')
                }
            }}
            onMouseUp={(e) => {
                e.stopPropagation()
                if (!crumbInfo.navigable) {
                    return
                }

                if (dragging !== 0) {
                    moveSelectedTo(crumbInfo.id)
                } else {
                    if (crumbInfo.file) {
                        clearSelected()
                        goToFile(crumbInfo.file)
                    } else if (crumbInfo.visitRoute) {
                        nav(crumbInfo.visitRoute)
                    } else {
                        console.error('NO FILE OR VISIT ROUTE IN CRUMB')
                    }
                }
                setMoveDest('')
            }}
        >
            <CrumbText crumb={crumbInfo} />
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
        <div className="overflow-menu" data-open={open}>
            {crumbs.map((item, i: number) => {
                return (
                    <div
                        key={`crumb-overflow-${i}`}
                        className="flex flex-row items-center "
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
            {crumbs.map((c, i) => (
                <div key={c.name} className="flex flex-row items-center">
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
                        <p className="crumb-text">...</p>
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

    const mode = useFileBrowserStore((state) => state.fbMode)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)

    const crumbs: Crumb[] = []

    // Add the share crumb before checking if we have folder info
    // because the base share page does not have a folderInfo, and
    // so we won't render anything past the check after this
    if (mode == FbModeT.share) {
        crumbs.push({
            name: 'Shared',
            id: 'shared',
            visitRoute: '/files/shared',
            navigable: folderInfo !== null,
        })
    }

    if (!user || !folderInfo?.Id()) {
        return <StyledLoaf crumbs={crumbs} moveSelectedTo={moveSelectedTo} />
    }

    crumbs.push(
        ...folderInfo.FormatParents().map((parent) => {
            return {
                name: parent.GetFilename(),
                id: parent.id,
                file: parent,
                navigable: true,
            }
        })
    )

    if (folderInfo.IsInTrash()) {
        crumbs.shift()
    }

    // Add the current folder, which is not always navigable
    crumbs.push({
        name: folderInfo.GetFilename(),
        id: folderInfo.Id(),
        file: folderInfo,
        navigable: navOnLast,
    })

    return <StyledLoaf crumbs={crumbs} moveSelectedTo={moveSelectedTo} />
}

export default Crumbs
