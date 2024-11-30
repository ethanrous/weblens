import {
    IconCalendar,
    IconChevronDown,
    IconChevronLeft,
    IconColumns,
    IconFileAnalytics,
    IconLayoutGrid,
    IconLayoutList,
    IconSortAZ,
    IconSortAscending2,
    IconSortDescending2,
    TablerIconsProps,
} from '@tabler/icons-react'
import { useResize } from '@weblens/components/hooks'
import WeblensButton from '@weblens/lib/WeblensButton'
import dirViewHeaderStyle from '@weblens/pages/FileBrowser/style/dirViewHeader.module.scss'
import { useEffect, useState } from 'react'

import { useFileBrowserStore } from './FBStateControl'
import { DirViewModeT } from './FileBrowserTypes'

const fileSortTypes = [
    { Name: 'Name', Icon: IconSortAZ },
    { Name: 'Date Modified', Icon: IconCalendar },
    { Name: 'Size', Icon: IconFileAnalytics },
]

const dirViewModes: {
    Mode: DirViewModeT
    Icon: (props: TablerIconsProps) => JSX.Element
}[] = [
    { Mode: DirViewModeT.Grid, Icon: IconLayoutGrid },
    { Mode: DirViewModeT.List, Icon: IconLayoutList },
    { Mode: DirViewModeT.Columns, Icon: IconColumns },
]

function FileSortBox() {
    const viewOpts = useFileBrowserStore((state) => state.viewOpts)
    const setViewOpts = useFileBrowserStore((state) => state.setViewOptions)
    const [open, setOpen] = useState(false)
    const [isVertical, setIsVertical] = useState(false)
    const [sortRef, setSortRef] = useState<HTMLDivElement>()
    const size = useResize(sortRef)

    useEffect(() => {
        if (size.width <= size.height) {
            if (isVertical) {
                return
            }
            setIsVertical(true)
            setOpen(false)
        } else if (isVertical) {
            setIsVertical(false)
        }
    }, [size])

    return (
        <div
            ref={setSortRef}
            className={dirViewHeaderStyle['file-sort-box']}
            data-open={open}
        >
            {isVertical && (
                <WeblensButton
                    subtle
                    Left={open ? IconChevronDown : IconChevronLeft}
                    onClick={(e) => {
                        e.stopPropagation()
                        setOpen(!open)
                    }}
                />
            )}
            <WeblensButton
                Left={
                    viewOpts.sortDirection === 1
                        ? IconSortDescending2
                        : IconSortAscending2
                }
                tooltip={
                    viewOpts.sortDirection === 1 ? 'Descending' : 'Ascending'
                }
                onClick={() =>
                    setViewOpts({ sortDirection: viewOpts.sortDirection * -1 })
                }
            />

            <div className={dirViewHeaderStyle['file-sort-divider']} />

            <div className={dirViewHeaderStyle['file-sort-group']}>
                {fileSortTypes.map((v) => {
                    return (
                        <WeblensButton
                            key={v.Name}
                            squareSize={40}
                            Left={v.Icon}
                            toggleOn={v.Name === viewOpts.sortFunc}
                            tooltip={v.Name}
                            onClick={(e) => {
                                e.stopPropagation()
                                setViewOpts({ sortKey: v.Name })
                            }}
                        />
                    )
                })}
            </div>

            <div className={dirViewHeaderStyle['file-sort-divider']} />

            <div className={dirViewHeaderStyle['file-sort-group']}>
                {dirViewModes.map((v) => {
                    return (
                        <WeblensButton
                            key={v.Mode}
                            squareSize={40}
                            Left={v.Icon}
                            toggleOn={v.Mode === viewOpts.dirViewMode}
                            tooltip={String(DirViewModeT[v.Mode])}
                            onClick={(e) => {
                                e.stopPropagation()
                                setViewOpts({ dirViewMode: v.Mode })
                            }}
                        />
                    )
                })}
            </div>
        </div>
    )
}

export default FileSortBox
