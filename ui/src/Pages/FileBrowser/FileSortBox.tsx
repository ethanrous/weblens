import {
    IconCalendar,
    IconFileAnalytics,
    IconSortAscending2,
    IconSortAZ,
    IconSortDescending2,
} from '@tabler/icons-react'
import { useResize } from '../../components/hooks'
import { useContext, useState } from 'react'
import { SelectIcon } from '../../components/WeblensButton'
import { FbContext } from './FileBrowser'

const fileSortTypes = [
    { Name: 'Name', Icon: <IconSortAZ className="button-icon" /> },
    { Name: 'Create Date', Icon: <IconCalendar className="button-icon" /> },
    { Name: 'Size', Icon: <IconFileAnalytics className="button-icon" /> },
]

function FileSortBox() {
    const { fbState, fbDispatch } = useContext(FbContext)
    const [sortFuncBox, setSortFuncBox] = useState(null)
    const sortFuncBoxSize = useResize(sortFuncBox)

    return (
        <div className="flex flex-row w-max shrink-0">
            <div className="file-sort-box">
                <div
                    className="sort-direction-box"
                    onClick={() =>
                        fbDispatch({
                            type: 'set_sort',
                            sortDirection: fbState.sortDirection * -1,
                        })
                    }
                >
                    {fbState.sortDirection === 1 && <IconSortDescending2 />}
                    {fbState.sortDirection === -1 && <IconSortAscending2 />}
                </div>
                <div className="h-full w-1 pt-1 pb-4 bg-[#333333]" />
                <div ref={setSortFuncBox}>
                    <div className="sort-func-selector">
                        {fileSortTypes.map((v, i) => {
                            return (
                                <SelectIcon
                                    key={v.Name}
                                    size={42}
                                    label={v.Name}
                                    icon={v.Icon}
                                    index={i}
                                    selected={fbState.sortFunc === v.Name}
                                    selectedIndex={fileSortTypes.findIndex(
                                        (v) => v.Name === fbState.sortFunc
                                    )}
                                    expandSize={sortFuncBoxSize.width}
                                    onClick={() => {
                                        fbDispatch({
                                            type: 'set_sort',
                                            sortType: v.Name,
                                        })
                                    }}
                                />
                            )
                        })}
                    </div>
                </div>
            </div>
        </div>
    )
}

export default FileSortBox
