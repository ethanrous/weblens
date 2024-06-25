import {
    IconCalendar,
    IconFileAnalytics,
    IconSortAscending2,
    IconSortAZ,
    IconSortDescending2,
} from '@tabler/icons-react'
import { useResize } from '../../components/hooks'
import { useContext, useState } from 'react'
import { FbContext } from '../../Files/filesContext'
import WeblensButton from '../../components/WeblensButton'

const fileSortTypes = [
    { Name: 'Name', Icon: IconSortAZ },
    { Name: 'Create Date', Icon: IconCalendar },
    { Name: 'Size', Icon: IconFileAnalytics },
]

function FileSortBox() {
    const { fbState, fbDispatch } = useContext(FbContext)
    const [sortFuncBox, setSortFuncBox] = useState(null)
    const sortFuncBoxSize = useResize(sortFuncBox)

    return (
        // <div className="flex flex-row w-max shrink-0 p-2">
        <div className="file-sort-box">
            <WeblensButton
                Left={
                    fbState.sortDirection === 1
                        ? IconSortDescending2
                        : IconSortAscending2
                }
                onClick={() =>
                    fbDispatch({
                        type: 'set_sort',
                        sortDirection: fbState.sortDirection * -1,
                    })
                }
            />

            <div className="h-full w-[1px] pt-1 pb-1 bg-[#333333]" />

            <div className="flex flex-row items-center">
                {fileSortTypes.map((v, i) => {
                    return (
                        <WeblensButton
                            key={v.Name}
                            squareSize={42}
                            Left={v.Icon}
                            toggleOn={v.Name === fbState.sortFunc}
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
        // </div>
    )
}

export default FileSortBox
