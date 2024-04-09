import { Box, Menu, Text } from "@mantine/core";
import { MediaDataT } from "../../types/Types";

export const GalleryMenu = ({
    media,
    menuPos,
    open,
    setOpen,
}: {
    media: MediaDataT;
    menuPos: { x: number; y: number };
    open: boolean;
    setOpen: (o: boolean) => void;
}) => {
    if (!open) {
        return null;
    }
    return (
        <Menu opened={open} onClose={() => setOpen(false)}>
            <Menu.Target>
                <Box
                    style={{
                        position: "fixed",
                        top: menuPos.y,
                        left: menuPos.x,
                    }}
                />
            </Menu.Target>

            <Menu.Dropdown>
                <Menu.Item>
                    <Text>Not impl</Text>
                    {/* <Box
                        style={{
                            height: "100px",
                            width: "100px",
                            backgroundColor: "#333333",
                        }}
                    ></Box> */}
                </Menu.Item>
            </Menu.Dropdown>
        </Menu>
    );
};
