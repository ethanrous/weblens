import { Box, Text } from "@mantine/core";
import { IconCaretDown, IconCaretUp, IconChevronCompactDown } from "@tabler/icons-react";
import { useEffect, useState } from "react";

const Dropdown = ({ data, selected, setSelected }) => {
    return (
        <Box className="weblens-dropdown-input-drop">
            {data.map(v => {
                return (
                    <Box key={v} className="weblens-dropdown-input-item" mod={{ "data-selected": Boolean(v === selected).toString() }} onClick={() => setSelected(v)}>
                        <Text className="weblens-dropdown-text">
                            {v}
                        </Text>
                    </Box>
                );
            })}

        </Box>
    );
};

export function WeblensSelect({ data, value, onChange }: { data: string[], value: string, onChange; }) {
    const [dropdownOpen, setDropdownOpen] = useState(false);

    useEffect(() => {
        if (dropdownOpen) {
            const handler = () => { setDropdownOpen(false); };
            window.addEventListener("click", handler);
            return () => window.removeEventListener("click", handler);
        }
    }, [dropdownOpen]);

    return (
        <Box style={{ width: "100%" }}>
            <Box className="weblens-dropdown-input" onClick={e => { e.stopPropagation(); setDropdownOpen(!dropdownOpen); }}>
                <Text truncate='end' className="weblens-dropdown-text">
                    {value}
                </Text>
                {!dropdownOpen && (
                    <IconCaretDown size={'20px'} style={{ marginRight: 8 }} color="#aaaaaa" />
                )}
                {dropdownOpen && (
                    <IconCaretUp size={'20px'} style={{ marginRight: 8 }} color="#aaaaaa" />
                )}
            </Box>
            {dropdownOpen && (
                <Dropdown data={data} selected={value} setSelected={onChange} />
            )}
        </Box>
    );
}