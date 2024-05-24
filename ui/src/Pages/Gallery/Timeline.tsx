import { Combobox, Indicator, Space, Text, useCombobox } from "@mantine/core";
import { PhotoGallery } from "../../components/MediaDisplay";
import { WeblensButton } from "../../components/WeblensButton";
import { useNavigate } from "react-router-dom";
import {
    memo,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useState,
} from "react";
import WeblensSlider from "../../components/WeblensSlider";
import { MediaImage } from "../../components/PhotoContainer";
import { GalleryContext } from "./Gallery";
import { IconFilter } from "@tabler/icons-react";
import { UserContextT } from "../../types/Types";
import { UserContext } from "../../Context";
import { FetchData } from "../../api/GalleryApi";
import WeblensMedia from "../../classes/Media";

const TimelineControls = () => {
    const { galleryState: mediaState, galleryDispatch: mediaDispatch } =
        useContext(GalleryContext);
    const combobox = useCombobox({
        onDropdownClose: () => {
            combobox.resetSelectedOption();
            mediaDispatch({
                type: "set_albums_filter",
                albumNames: selectedAlbums,
            });
            mediaDispatch({ type: "set_raw_toggle", raw: rawOn });
        },
    });

    const [selectedAlbums, setSelectedAlbums] = useState(
        mediaState.albumsFilter
    );
    const [rawOn, setRawOn] = useState(mediaState.includeRaw);

    const albumsOptions = useMemo(() => {
        return Array.from(mediaState.albumsMap.values()).map((a) => {
            const included = selectedAlbums.includes(a.Name);
            if (!a.CoverMedia) {
                a.CoverMedia = new WeblensMedia({ mediaId: a.Cover });
            }
            return (
                <WeblensButton
                    key={a.Name}
                    label={a.Name}
                    allowRepeat
                    subtle
                    toggleOn={!included}
                    height={32}
                    Left={
                        <MediaImage
                            media={a.CoverMedia}
                            quality="thumbnail"
                            containerStyle={{ borderRadius: 4 }}
                        />
                    }
                    onClick={() =>
                        setSelectedAlbums((p) => {
                            if (included) {
                                p.splice(p.indexOf(a.Name));
                                return [...p];
                            } else {
                                p.push(a.Name);
                                return [...p];
                            }
                        })
                    }
                />
            );
        });

    }, [mediaState.albumsMap, selectedAlbums]);

    const rawClick = useCallback(() => setRawOn(!rawOn), [rawOn, setRawOn]);
    const selectClick = useCallback(() => {
        mediaDispatch({
            type: "set_selecting",
            selecting: !mediaState.selecting,
        });
    }, [mediaDispatch, mediaState.selecting]);

    return (
        <div className="flex flex-row items-center grow m-2 h-14 w-11/12">
            <WeblensSlider
                value={mediaState.imageSize}
                width={200}
                height={35}
                min={150}
                max={500}
                callback={(s) =>
                    mediaDispatch({ type: "set_image_size", size: s })
                }
            />
            <Space w={20} />
            <Combobox
                arrowSize={0}
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
                        disabled={
                            !selectedAlbums.length && !mediaState.includeRaw
                        }
                        zIndex={3}
                    >
                        <IconFilter
                            onClick={() => combobox.toggleDropdown()}
                            style={{ cursor: "pointer" }}
                        />
                    </Indicator>
                </Combobox.Target>

                <Combobox.Dropdown className="options-dropdown">
                    <Combobox.Header>
                        <div style={{ paddingBottom: 10 }}>
                            <Text fw={600}>Gallery Filters</Text>
                        </div>
                    </Combobox.Header>
                    <Space h={10} />
                    <WeblensButton
                        label="Show RAWs"
                        height={40}
                        allowRepeat
                        toggleOn={rawOn}
                        onClick={rawClick}
                    />
                    <Space h={10} />
                    <Combobox.Options>
                        <Combobox.Group
                            label="Album Filter"
                            className="flex flex-col items-center"
                        >
                            {albumsOptions}
                        </Combobox.Group>
                        <Combobox.Group label="Filetypes">
                            {/* {albumsOptions} */}
                        </Combobox.Group>
                    </Combobox.Options>
                </Combobox.Dropdown>
            </Combobox>
            <div className="flex grow w-0 justify-end">
                <WeblensButton
                    label="Select"
                    allowRepeat
                    height={40}
                    width={75}
                    centerContent
                    toggleOn={mediaState.selecting}
                    onClick={selectClick}
                />
            </div>
        </div>
    );
};

const NoMediaDisplay = () => {
    const nav = useNavigate();
    return (
        <div className="flex flex-col items-center w-full">
            <div className="mt-20 gap-6 w-max">
                <Text c="white" fw={700} size="31px">
                    No media to display
                </Text>
                <Text c="white">Upload files then add them to an album</Text>
                <div style={{ height: "max-content", width: "100%", gap: 10 }}>
                    <WeblensButton
                        height={48}
                        label="Upload Media"
                        centerContent
                        onClick={() => nav("/files")}
                    />
                    <WeblensButton
                        height={48}
                        label="View Albums"
                        centerContent
                        onClick={() => nav("/albums")}
                    />
                </div>
            </div>
        </div>
    );
};

export const Timeline = memo(
    ({ page }: { page: string }) => {
        const { galleryState, galleryDispatch } = useContext(GalleryContext);
        const { authHeader }: UserContextT = useContext(UserContext);

        useEffect(() => {
            if (!galleryState) {
                return;
            }

            galleryDispatch({ type: "add_loading", loading: "media" });
            FetchData(galleryState, galleryDispatch, authHeader).then(() =>
                galleryDispatch({ type: "remove_loading", loading: "media" })
            );
        }, [
            galleryState?.includeRaw,
            galleryState?.albumsFilter,
            page,
            authHeader,
        ]);

        const medias = useMemo(() => {
            if (!galleryState) {
                return [];
            }

            return Array.from(galleryState.mediaMap.values())
                .filter((m) => {
                    if (galleryState.searchContent === "") {
                        return true;
                    }

                    return m.MatchRecogTag(galleryState.searchContent);
                })
                .reverse();
        }, [galleryState?.mediaMap, galleryState?.searchContent]);

        if (galleryState.loading.includes("media")) {
            return null;
        }

        if (medias.length === 0) {
            return <NoMediaDisplay />;
        }

        return (
            <div className="flex flex-col items-center">
                <TimelineControls />
                <PhotoGallery medias={medias} />
            </div>
        );
    },
    (prev, next) => {
        return prev.page === next.page;
    }
);
