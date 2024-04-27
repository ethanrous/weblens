import { Box } from "@mantine/core";
import WeblensMedia from "../../classes/Media";
import { WeblensButton } from "../../components/WeblensButton";
import { useClick, useKeyDown } from "../../components/hooks";
import { memo, useContext, useState } from "react";
import { SetAlbumCover } from "../../api/GalleryApi";
import { UserContext } from "../../Context";
import { IconEyeOff } from "@tabler/icons-react";
import { UserContextT } from "../../types/Types";

export const GalleryMenu = memo(
    ({
        media,
        height,
        width,
        open,
        setOpen,
        updateAlbum,
        albumId,
    }: {
        media: WeblensMedia;
        height: number;
        width: number;
        open: boolean;
        setOpen: (o: boolean) => void;
        updateAlbum?: () => void;
        albumId?: string;
    }) => {
        const { authHeader }: UserContextT = useContext(UserContext);
        const [menuRef, setMenuRef] = useState(null);
        useClick(
            () => {
                setOpen(false);
            },
            menuRef,
            !open
        );
        useKeyDown("Escape", (e) => {
            if (open) {
                setOpen(false);
            }
        });

        return (
            <Box
                ref={setMenuRef}
                className="media-menu-container"
                mod={{ open: open }}
                onClick={(e) => {
                    e.stopPropagation();
                    setOpen(false);
                }}
                onContextMenu={(e) => {
                    e.stopPropagation();
                    e.preventDefault();
                    setOpen(false);
                }}
            >
                <WeblensButton
                    label="Hide"
                    Left={<IconEyeOff />}
                    disabled={media.IsHidden()}
                    subtle
                    height={48}
                    style={{ opacity: open ? "100%" : "0%" }}
                    onClick={async (e) => {
                        e.stopPropagation();
                        const r = await media.Hide(authHeader);
                        if (r.status !== 200) {
                            return false;
                        }
                        return true;
                    }}
                />
                <WeblensButton
                    label="Set as Cover"
                    disabled={!albumId}
                    subtle
                    height={48}
                    style={{ opacity: open ? "100%" : "0%" }}
                    onClick={async (e) => {
                        e.stopPropagation();
                        const r = await SetAlbumCover(
                            albumId,
                            media.Id(),
                            authHeader
                        );
                        updateAlbum();
                        if (r.status !== 200) {
                            return false;
                        }
                        return true;
                    }}
                />
            </Box>
        );
    },
    (prev, next) => {
        if (prev.height !== next.height) {
            return false;
        }
        if (prev.width !== next.width) {
            return false;
        }
        if (prev.open !== next.open) {
            return false;
        }

        return true;
    }
);
