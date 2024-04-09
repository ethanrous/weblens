import {
    Box,
    Button,
    Combobox,
    Indicator,
    Modal,
    Slider,
    Space,
    Switch,
    Tabs,
    Text,
    TextInput,
    useCombobox,
} from "@mantine/core";
import {
    useEffect,
    useReducer,
    useMemo,
    useRef,
    useContext,
    useState,
    memo,
} from "react";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import { IconCheck, IconFilter, IconPlus } from "@tabler/icons-react";

import HeaderBar from "../../components/HeaderBar";
import Presentation from "../../components/Presentation";
import { PhotoGallery } from "../../components/MediaDisplay";
import {
    mediaReducer,
    useKeyDownGallery,
    handleWebsocket,
    GalleryAction,
} from "./GalleryLogic";
import { CreateAlbum, FetchData, GetAlbums } from "../../api/GalleryApi";
import {
    AlbumData,
    MediaDataT,
    MediaStateT,
    UserContextT,
    UserInfoT,
} from "../../types/Types";
import { userContext } from "../../Context";
import { ColumnBox, RowBox } from "../FileBrowser/FilebrowserStyles";
import { Albums } from "./Albums";
import { WeblensButton } from "../../components/WeblensButton";
import useWeblensSocket from "../../api/Websocket";
import { GalleryMenu } from "./GalleryMenu";
import { useMediaType } from "../../components/hooks";

const NoMediaDisplay = () => {
    const nav = useNavigate();
    return (
        <ColumnBox style={{ marginTop: 75, gap: 25, width: "max-content" }}>
            <Text c="white" fw={700} size="31px">
                No media to display
            </Text>
            <Text c="white">Upload files then add them to an album</Text>
            <RowBox style={{ height: "max-content", width: "100%", gap: 10 }}>
                <WeblensButton
                    label="Upload Media"
                    centerContent
                    onClick={() => nav("/files")}
                />
                <WeblensButton
                    label="View Albums"
                    centerContent
                    onClick={() => nav("/files")}
                />
            </RowBox>
        </ColumnBox>
    );
};

const ImageSizeSlider = ({ imageSize, dispatch }) => {
    return (
        <Slider
            color="#4444ff"
            label={`Image Size`}
            defaultValue={300}
            value={imageSize}
            w={200}
            min={100}
            max={500}
            step={10}
            marks={[
                { value: 100, label: "10%" },
                { value: 300, label: "50%" },
                { value: 500, label: "100%" },
            ]}
            onChange={(e) => dispatch({ type: "set_image_size", size: e })}
            onDoubleClick={() =>
                dispatch({ type: "set_image_size", size: 300 })
            }
            style={{ paddingBottom: 10 }}
        />
    );
};

const TimelineControls = ({
    rawSelected,
    selecting,
    albumsFilter,
    imageSize,
    albumsMap,
    dispatch,
}: {
    rawSelected: boolean;
    selecting: boolean;
    albumsFilter: string[];
    imageSize: number;
    albumsMap: Map<string, AlbumData>;
    dispatch: (action: GalleryAction) => void;
}) => {
    const albumNames = useMemo(
        () => Array.from(albumsMap.values()).map((v) => v.Name),
        [albumsMap.size, albumsMap]
    );
    const combobox = useCombobox({
        onDropdownClose: () => {
            combobox.resetSelectedOption();
            dispatch({ type: "set_albums_filter", albumNames: selectedAlbums });
            dispatch({ type: "set_raw_toggle", raw: rawOn });
        },
    });
    const [selectedAlbums, setSelectedAlbums] = useState(albumsFilter);
    const [rawOn, setRawOn] = useState(rawSelected);

    const albumsOptions = useMemo(() => {
        const options = albumNames.map((name) => {
            return (
                <Combobox.Option value={name} key={name}>
                    <RowBox style={{ justifyContent: "space-between" }}>
                        <Text className="menu-item-text">{name}</Text>
                        {selectedAlbums.includes(name) && <IconCheck />}
                    </RowBox>
                </Combobox.Option>
            );
        });
        return options;
    }, [albumNames, selectedAlbums]);

    return (
        <RowBox
            style={{
                flexGrow: 1,
                marginRight: "2vw",
                height: 55,
                alignItems: "center",
                padding: 10,
            }}
        >
            <ImageSizeSlider imageSize={imageSize} dispatch={dispatch} />
            <Space w={20} />
            <Combobox
                store={combobox}
                width={200}
                position="bottom-start"
                withArrow
                withinPortal={false}
                positionDependencies={[selectedAlbums]}
                onOptionSubmit={(val) => {
                    setSelectedAlbums((current) =>
                        current.includes(val)
                            ? current.filter((item) => item !== val)
                            : [...current, val]
                    );
                }}
            >
                <Combobox.Target>
                    <Indicator
                        color="#4444ff"
                        disabled={!selectedAlbums.length && !rawSelected}
                        zIndex={3}
                    >
                        <IconFilter
                            onClick={() => combobox.toggleDropdown()}
                            style={{ cursor: "pointer" }}
                        />
                    </Indicator>
                </Combobox.Target>

                <Combobox.Dropdown style={{ padding: 10 }}>
                    <Combobox.Header>
                        <ColumnBox style={{ paddingBottom: 10 }}>
                            <Text fw={600}>Gallery Filters</Text>
                        </ColumnBox>
                    </Combobox.Header>
                    <Space h={10} />
                    <WeblensButton
                        label="Show RAWs"
                        toggleOn={rawOn}
                        onClick={() => setRawOn(!rawOn)}
                    />
                    <Space h={10} />
                    <Combobox.Options>
                        <Combobox.Group label="Albums">
                            {albumsOptions}
                        </Combobox.Group>
                        <Combobox.Group label="Filetypes">
                            {/* {albumsOptions} */}
                        </Combobox.Group>
                    </Combobox.Options>
                </Combobox.Dropdown>
            </Combobox>
            <Space w={1} flex={2} />
            <WeblensButton
                label="Select"
                centerContent
                toggleOn={selecting}
                onClick={() =>
                    dispatch({ type: "set_selecting", selecting: !selecting })
                }
                style={{ width: 100, padding: 2 }}
            />
        </RowBox>
    );
};

const AlbumsControls = ({ albumId, imageSize, rawSelected, dispatch }) => {
    const [newAlbumModal, setNewAlbumModal] = useState(false);
    const [newAlbumName, setNewAlbumName] = useState("");
    const { authHeader }: UserContextT = useContext(userContext);

    if (!albumId) {
        return (
            <Box>
                <Button
                    color="#4444ff"
                    onClick={() => {
                        dispatch({ type: "set_block_focus", block: true });
                        setNewAlbumModal(true);
                    }}
                    leftSection={<IconPlus />}
                >
                    New Album
                </Button>

                <Modal
                    opened={newAlbumModal}
                    onClose={() => {
                        dispatch({ type: "set_block_focus", block: false });
                        setNewAlbumModal(false);
                    }}
                    title="New Album"
                >
                    <TextInput
                        value={newAlbumName}
                        placeholder="Album name"
                        onChange={(e) => setNewAlbumName(e.currentTarget.value)}
                    />
                    <Space h={"md"} />
                    <Button
                        onClick={() => {
                            CreateAlbum(newAlbumName, authHeader).then(() =>
                                GetAlbums(authHeader).then((val) =>
                                    dispatch({
                                        type: "set_albums",
                                        albums: val,
                                    })
                                )
                            );
                            dispatch({ type: "set_block_focus", block: false });
                            setNewAlbumModal(false);
                        }}
                    >
                        Create
                    </Button>
                </Modal>
            </Box>
        );
    } else {
        return (
            <RowBox style={{ width: "max-content" }}>
                <ImageSizeSlider imageSize={imageSize} dispatch={dispatch} />
                <Space w={20} />
                <WeblensButton
                    label="RAWs"
                    toggleOn={rawSelected}
                    onClick={() =>
                        dispatch({
                            type: "set_raw_toggle",
                            raw: !rawSelected,
                        })
                    }
                    width={"70px"}
                />
            </RowBox>
        );
    }
};

function ViewSwitch({
    mediaState,
    page,
    timeline,
    albums,
    albumId,
    dispatch,
}: {
    mediaState: MediaStateT;
    page: string;
    timeline;
    albums;
    albumId: string;
    dispatch: (action: GalleryAction) => void;
}) {
    const nav = useNavigate();
    const [hovering, setHovering] = useState(false);
    let albumStyle = {};
    if (albumId && hovering) {
        albumStyle = {
            backgroundColor: "#2e2e2e",
            outline: "1px solid #4444ff",
        };
    } else if (albumId) {
        albumStyle = {
            backgroundColor: "#00000000",
            outline: "1px solid #4444ff",
        };
    }
    return (
        <Tabs
            value={page}
            keepMounted={false}
            onChange={(p) => nav(`/${p}`)}
            variant="pills"
            style={{ height: "100%" }}
        >
            <RowBox style={{ height: 60 }}>
                <Tabs.List
                    style={{
                        marginLeft: 20,
                        flexDirection: "row",
                        width: 175,
                        flexShrink: 0,
                    }}
                >
                    <Tabs.Tab value="timeline" color="#4444ff">
                        Timeline
                    </Tabs.Tab>
                    <Tabs.Tab
                        value="albums"
                        color="#4444ff"
                        onMouseOver={() => setHovering(true)}
                        onMouseLeave={() => setHovering(false)}
                        style={albumStyle}
                    >
                        Albums
                    </Tabs.Tab>
                </Tabs.List>
                <Space w={30} />
                {page === "timeline" && (
                    <TimelineControls
                        rawSelected={mediaState.includeRaw}
                        selecting={mediaState.selecting}
                        albumsFilter={mediaState.albumsFilter}
                        imageSize={mediaState.imageSize}
                        albumsMap={mediaState.albumsMap}
                        dispatch={dispatch}
                    />
                )}
                {page === "albums" && (
                    <AlbumsControls
                        albumId={albumId}
                        imageSize={mediaState.imageSize}
                        rawSelected={mediaState.includeRaw}
                        dispatch={dispatch}
                    />
                )}
            </RowBox>

            <Tabs.Panel value="timeline" style={{ height: "98%" }}>
                <ColumnBox>{timeline}</ColumnBox>
            </Tabs.Panel>
            <Tabs.Panel value="albums" style={{ height: "98%" }}>
                <ColumnBox style={{ alignItems: "center" }}>{albums}</ColumnBox>
            </Tabs.Panel>
        </Tabs>
    );
}

export const Timeline = memo(
    ({
        mediaState,
        imageBaseScale,
        selecting,
        loading,
        page,
        dispatch,
    }: {
        mediaState: MediaStateT;
        imageBaseScale: number;
        selecting: boolean;
        loading: string[];
        page: string;
        dispatch: (value) => void;
    }) => {
        const { authHeader }: UserContextT = useContext(userContext);
        const mediaTypeMap = useMediaType();
        useEffect(() => {
            if (!mediaTypeMap) {
                return;
            }
            dispatch({ type: "add_loading", loading: "media" });
            FetchData(mediaState, dispatch, authHeader).then(() =>
                dispatch({ type: "remove_loading", loading: "media" })
            );
        }, [
            mediaState.includeRaw,
            mediaState.albumsFilter,
            page,
            mediaTypeMap,
            authHeader,
        ]);

        const medias = useMemo(() => {
            if (!mediaTypeMap) {
                return [];
            }

            return Array.from(mediaState.mediaMap.values())
                .filter((v) => {
                    v.mediaType = mediaTypeMap.get(v.mimeType);
                    if (v.selected === undefined) {
                        v.selected = false;
                    }
                    if (mediaState.searchContent === "") {
                        return true;
                    }
                    if (!v.recognitionTags) {
                        return false;
                    }
                    for (const tag of v.recognitionTags) {
                        if (tag.includes(mediaState.searchContent)) {
                            return true;
                        }
                    }
                    return false;
                })
                .reverse();
        }, [mediaState.mediaMap.size, mediaState.searchContent, mediaTypeMap]);

        if (loading.includes("media")) {
            return null;
        }

        if (medias.length === 0) {
            return <NoMediaDisplay />;
        }

        return (
            <ColumnBox>
                <PhotoGallery
                    medias={medias}
                    selecting={mediaState.selecting}
                    imageBaseScale={imageBaseScale}
                    title={null}
                    dispatch={dispatch}
                />
            </ColumnBox>
        );
    },
    (prev, next) => {
        if (prev.mediaState.mediaMap !== next.mediaState.mediaMap) {
            return false;
        } else if (prev.selecting !== next.selecting) {
            return false;
        } else if (
            prev.mediaState.albumsFilter !== next.mediaState.albumsFilter
        ) {
            return false;
        } else if (prev.loading !== next.loading) {
            return false;
        } else if (prev.imageBaseScale !== next.imageBaseScale) {
            return false;
        }
        return true;
    }
);

const Gallery = () => {
    const [mediaState, dispatch]: [MediaStateT, React.Dispatch<any>] =
        useReducer(mediaReducer, {
            mediaMap: new Map<string, MediaDataT>(),
            selected: new Map<string, boolean>(),
            mediaMapUpdated: 0,
            albumsMap: new Map<string, AlbumData>(),
            presentingMedia: null,
            albumsFilter: [],
            loading: [],
            includeRaw: false,
            newAlbumDialogue: false,
            blockSearchFocus: false,
            selecting: false,
            imageSize: 300,
            showingCount: 300,
            scanProgress: 0,
            searchContent: "",
            menuOpen: false,
            menuTargetId: "",
            menuPos: { x: 0, y: 0 },
        });

    const nav = useNavigate();
    const { authHeader, usr }: { authHeader; usr: UserInfoT } =
        useContext(userContext);

    const loc = useLocation();
    const page =
        loc.pathname === "/" || loc.pathname === "/timeline"
            ? "timeline"
            : "albums";
    const albumId = useParams()["*"];
    const { lastMessage } = useWeblensSocket();

    const viewportRef: React.Ref<HTMLDivElement> = useRef();

    const searchRef = useRef();
    useKeyDownGallery(mediaState.blockSearchFocus, searchRef, dispatch);
    useEffect(() => {
        handleWebsocket(lastMessage, dispatch);
    }, [lastMessage]);

    useEffect(() => {
        if (usr.isLoggedIn) {
            dispatch({ type: "remove_loading", loading: "login" });
        } else if (usr.isLoggedIn === undefined) {
            dispatch({ type: "add_loading", loading: "login" });
        } else if (usr.isLoggedIn === false) {
            nav("/login");
        }
    }, [usr]);

    useEffect(() => {
        if (authHeader.Authorization !== "" && page !== "albums") {
            dispatch({ type: "add_loading", loading: "albums" });
            GetAlbums(authHeader).then((val) => {
                dispatch({ type: "set_albums", albums: val });
                dispatch({ type: "remove_loading", loading: "albums" });
            });
        }
    }, [authHeader, page]);

    return (
        <Box>
            <HeaderBar
                // searchContent={mediaState.searchContent}
                dispatch={dispatch}
                page={"gallery"}
                // searchRef={searchRef}
                loading={mediaState.loading}
                // progress={mediaState.scanProgress}
            />
            <Presentation
                itemId={mediaState.presentingMedia?.mediaId}
                mediaData={mediaState.presentingMedia}
                element={null}
                dispatch={dispatch}
            />
            <GalleryMenu
                media={mediaState.mediaMap.get(mediaState.menuTargetId)}
                menuPos={mediaState.menuPos}
                open={mediaState.menuOpen}
                setOpen={(o: boolean) =>
                    dispatch({ type: "set_menu_open", open: o })
                }
            />
            <ColumnBox
                style={{ height: "100vh", alignItems: "normal", zIndex: 2 }}
            >
                {/* <GalleryControls mediaState={mediaState} page={page} albumId={albumId} dispatch={dispatch} /> */}
                <Box
                    ref={viewportRef}
                    style={{
                        height: "calc(100% - 80px)",
                        width: "100%",
                        position: "absolute",
                    }}
                >
                    <ViewSwitch
                        mediaState={mediaState}
                        page={page}
                        timeline={
                            <Timeline
                                mediaState={mediaState}
                                imageBaseScale={mediaState.imageSize}
                                selecting={mediaState.selecting}
                                loading={mediaState.loading}
                                page={page}
                                dispatch={dispatch}
                            />
                        }
                        albums={
                            <Albums
                                mediaState={mediaState}
                                selectedAlbum={albumId}
                                dispatch={dispatch}
                            />
                        }
                        albumId={albumId}
                        dispatch={dispatch}
                    />
                </Box>
            </ColumnBox>
        </Box>
    );
};

export default Gallery;
