import { useCallback, useEffect } from "react";
import { AlbumData, MediaData, MediaStateT, UserInfoT } from "../../types/Types";
import { notifications } from "@mantine/notifications";

export type GalleryAction = {
    type: string;
    medias?: MediaData[];
    albums?: AlbumData[];
    albumId?: string;
    mediaId?: string;
    media?: MediaData;
    albumNames?: string[];
    include?: boolean;
    block?: boolean;
    progress?: number;
    loading?: string;
    search?: string;
    selected?: boolean;
    selecting?: boolean;
    open?: boolean;
    size?: number;
    raw?: boolean;
    targetId?: string;
    pos?: { x: number; y: number };
};

export function mediaReducer(state: MediaStateT, action: GalleryAction): MediaStateT {
    switch (action.type) {
        case "set_media": {
            state.mediaMap.clear();
            if (action.medias) {
                let prev: MediaData;
                for (const m of action.medias) {
                    state.mediaMap.set(m.mediaId, m);
                    if (prev) {
                        prev.Next = m;
                        m.Previous = prev;
                    }
                    prev = m;
                }
            }
            return {
                ...state,
                mediaMap: new Map(state.mediaMap),
            };
        }

        case "set_selected": {
            const item = state.mediaMap.get(action.mediaId);
            if (!item) {
                console.warn("Trying to select media that does not exist");
                return { ...state };
            }
            item.selected = action.selected;
            state.mediaMap.set(action.mediaId, item);
            return { ...state, mediaMap: new Map(state.mediaMap) };
        }

        case "set_selecting": {
            if (!action.selecting) {
                for (const key of state.mediaMap.keys()) {
                    state.mediaMap.get(key).selected = false;
                }
            }
            return { ...state, selecting: action.selecting, mediaMap: new Map(state.mediaMap) };
        }

        case "set_albums": {
            if (!action.albums) {
                return { ...state };
            }
            state.albumsMap.clear();
            for (const a of action.albums) {
                state.albumsMap.set(a.Id, a);
            }
            return { ...state };
        }

        case "set_album_media": {
            const album = state.albumsMap.get(action.albumId);
            album.CoverMedia = action.media;
            state.albumsMap.set(action.albumId, album);
            return { ...state };
        }

        case "set_albums_filter": {
            return {
                ...state,
                albumsFilter: action.albumNames,
            };
        }

        case "set_image_size": {
            return {
                ...state,
                imageSize: action.size,
            };
        }

        case "set_block_focus": {
            return {
                ...state,
                blockSearchFocus: action.block,
            };
        }

        case "set_new_album_open": {
            return {
                ...state,
                blockSearchFocus: action.open,
                newAlbumDialogue: action.open,
            };
        }

        case "delete_from_map": {
            state.mediaMap.delete(action.media.mediaId);
            return { ...state, mediaMap: new Map(state.mediaMap) };
        }

        case "set_scan_progress": {
            return {
                ...state,
                scanProgress: action.progress,
            };
        }

        case "add_loading": {
            const newLoading = state.loading.filter((v) => v !== action.loading);
            newLoading.push(action.loading);
            return {
                ...state,
                loading: newLoading,
            };
        }

        case "remove_loading": {
            const newLoading = state.loading.filter((v) => v !== action.loading);
            return {
                ...state,
                loading: newLoading,
            };
        }

        case "set_menu_open": {
            return { ...state, menuOpen: action.open };
        }

        case "set_menu_target": {
            return { ...state, menuTargetId: action.targetId };
        }

        case "set_menu_pos": {
            return { ...state, menuPos: action.pos, menuOpen: true };
        }

        case "set_raw_toggle": {
            if (action.raw === state.includeRaw) {
                return { ...state };
            }
            window.scrollTo({
                top: 0,
                behavior: "smooth",
            });
            state.mediaMap.clear();
            return {
                ...state,
                includeRaw: action.raw,
                mediaMap: new Map(state.mediaMap),
            };
        }

        case "set_search": {
            return {
                ...state,
                searchContent: action.search,
            };
        }

        case "set_presentation": {
            return {
                ...state,
                presentingMedia: action.media,
            };
        }

        case "presentation_next": {
            return {
                ...state,
                presentingMedia: state.presentingMedia.Next ? state.presentingMedia.Next : state.presentingMedia,
            };
        }

        case "presentation_previous": {
            return {
                ...state,
                presentingMedia: state.presentingMedia.Previous
                    ? state.presentingMedia.Previous
                    : state.presentingMedia,
            };
        }

        case "stop_presenting": {
            if (state.presentingMedia === null) {
                return {
                    ...state,
                };
            }
            try {
                state.presentingMedia.ImgRef.current.scrollIntoView({
                    behavior: "smooth",
                    block: "nearest",
                    inline: "start",
                });
            } catch {
                console.error("No img ref: ", state.presentingMedia);
            }
            return {
                ...state,
                presentingMedia: null,
            };
        }

        default: {
            console.error("Do not have handler for dispatch type", action.type);
            return {
                ...state,
            };
        }
    }
}

export const useKeyDownGallery = (blockSearchFocus: boolean, searchRef, dispatch: (action: GalleryAction) => void) => {
    const onKeyDown = useCallback(
        (event) => {
            if (
                !blockSearchFocus &&
                !event.metaKey &&
                ((event.which >= 65 && event.which <= 90) || event.key === "Backspace")
            ) {
                searchRef.current.focus();
            } else if (event.key === "Escape" && searchRef.current) {
                searchRef.current.blur();
            } else if (event.key === "Escape") {
                dispatch({ type: "set_selecting", selecting: false });
            }
        },
        [blockSearchFocus, searchRef],
    );
    useEffect(() => {
        document.addEventListener("keydown", onKeyDown);
        return () => {
            document.removeEventListener("keydown", onKeyDown);
        };
    }, [onKeyDown]);
};

export function handleWebsocket(lastMessage, dispatch) {
    if (lastMessage) {
        const msgData = JSON.parse(lastMessage.data);
        switch (msgData["type"]) {
            case "item_update": {
                return;
            }
            case "item_deleted": {
                dispatch({
                    type: "delete_from_map",
                    media: msgData["content"].hash,
                });
                return;
            }
            case "scan_directory_progress": {
                dispatch({
                    type: "set_scan_progress",
                    progress: (1 - msgData["remainingTasks"] / msgData["totalTasks"]) * 100,
                });
                return;
            }
            case "error": {
                notifications.show({ message: msgData["error"], color: "red" });
                return;
            }
            default: {
                console.error("Got unexpected websocket message: ", msgData);
                return;
            }
        }
    }
}
