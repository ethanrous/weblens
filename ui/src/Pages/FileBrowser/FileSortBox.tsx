import { Box } from "@mantine/core";
import { RowBox } from "./FileBrowserStyles";
import {
    IconCalendar,
    IconFileAnalytics,
    IconSortAscending2,
    IconSortAZ,
    IconSortDescending2,
} from "@tabler/icons-react";
import { useResize } from "../../components/hooks";
import { useState } from "react";
import { SelectIcon } from "../../components/WeblensButton";

const fileSortTypes = [
    { Name: "Name", Icon: <IconSortAZ className="button-icon" /> },
    { Name: "Create Date", Icon: <IconCalendar className="button-icon" /> },
    { Name: "Size", Icon: <IconFileAnalytics className="button-icon" /> },
];

export const FileSortBox = ({ fb, dispatch }) => {
    const [sortFuncBox, setSortFuncBox] = useState(null);
    const sortFuncBoxSize = useResize(sortFuncBox);

    return (
        <RowBox style={{ width: "max-content", flexShrink: 0 }}>
            <Box className="file-sort-box">
                <Box
                    className="sort-direction-box"
                    onClick={() =>
                        dispatch({
                            type: "set_sort",
                            sortDirection: fb.sortDirection * -1,
                        })
                    }
                >
                    {fb.sortDirection === 1 && <IconSortDescending2 />}
                    {fb.sortDirection === -1 && <IconSortAscending2 />}
                </Box>
                <Box
                    style={{
                        height: "100%",
                        width: 1,
                        paddingTop: 4,
                        paddingBottom: 4,
                        backgroundColor: "#333333",
                    }}
                />
                <Box ref={setSortFuncBox}>
                    <Box className="sort-func-selector">
                        {fileSortTypes.map((v, i) => {
                            return (
                                <SelectIcon
                                    key={v.Name}
                                    size={42}
                                    label={v.Name}
                                    icon={v.Icon}
                                    index={i}
                                    selected={fb.sortFunc === v.Name}
                                    selectedIndex={fileSortTypes.findIndex(
                                        (v) => v.Name === fb.sortFunc
                                    )}
                                    expandSize={sortFuncBoxSize.width}
                                    onClick={() => {
                                        dispatch({
                                            type: "set_sort",
                                            sortType: v.Name,
                                        });
                                    }}
                                />
                            );
                        })}
                    </Box>
                </Box>
            </Box>
        </RowBox>
    );
};
