import { notifications } from "@mantine/notifications";
import { AlbumData, AuthHeaderT, MediaStateT } from "../types/Types";
import API_ENDPOINT from "./ApiEndpoint";
import WeblensMedia from "../classes/Media";

export async function FetchData(
    mediaState: MediaStateT,
    dispatch,
    authHeader: AuthHeaderT
) {
    if (!authHeader || authHeader.Authorization === "") {
        return;
    }

    try {
        const url = new URL(`${API_ENDPOINT}/media`);
        url.searchParams.append("raw", mediaState.includeRaw.toString());
        if (mediaState.albumsFilter) {
            url.searchParams.append(
                "albums",
                JSON.stringify(
                    Array.from(mediaState.albumsMap.values())
                        .filter((v) => mediaState.albumsFilter.includes(v.Name))
                        .map((v) => v.Id)
                )
            );
        }
        const data = await fetch(url.toString(), { headers: authHeader }).then(
            (res) => {
                if (res.status !== 200) {
                    return Promise.reject("Failed to get media");
                } else {
                    return res.json();
                }
            }
        );

        console.log(data.Media);
        const medias = data.Media.map((m) => {
            return new WeblensMedia(m);
        });

        dispatch({
            type: "set_media",
            medias: medias,
        });
    } catch (error) {
        console.error(
            "Error fetching data - Ethan you wrote this, its not a js err"
        );
    }
}

export async function CreateAlbum(albumName, authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/album`);
    const body = {
        name: albumName,
    };
    await fetch(url.toString(), {
        method: "POST",
        headers: authHeader,
        body: JSON.stringify(body),
    });
}

export async function GetAlbums(authHeader): Promise<AlbumData[]> {
    const url = new URL(`${API_ENDPOINT}/albums`);

    const res = await fetch(url.toString(), { headers: authHeader }).catch(
        (r) => console.error(r)
    );
    if (!res) {
        return;
    }

    if (res.status === 200) {
        return (await res.json()).albums;
    }
    notifications.show({
        title: "Failed to get albums",
        message: res.statusText,
        color: "red",
    });
}

export async function GetAlbumMedia(
    albumId,
    includeRaw,
    authHeader
): Promise<{ albumMeta: AlbumData; media: WeblensMedia[] }> {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`);
    url.searchParams.append("raw", includeRaw);
    const res = await fetch(url.toString(), { headers: authHeader });
    if (res.status === 404) {
        return Promise.reject(404);
    } else if (res.status !== 200) {
        return Promise.reject("Unknown error");
    }

    const data = await res.json();
    const medias = data.medias.map((m) => {
        return new WeblensMedia(m);
    });
    return { albumMeta: data.albumMeta, media: medias };
}

export async function AddMediaToAlbum(
    albumId,
    mediaIds,
    folderIds,
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`);
    const body = {
        newMedia: mediaIds,
        newFolders: folderIds,
    };
    return (
        await fetch(url.toString(), {
            method: "PATCH",
            headers: authHeader,
            body: JSON.stringify(body),
        })
    ).json();
}

export async function RemoveMediaFromAlbum(
    albumId,
    mediaIds,
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`);
    const body = {
        removeMedia: mediaIds,
    };
    await fetch(url.toString(), {
        method: "PATCH",
        headers: authHeader,
        body: JSON.stringify(body),
    });
}

export async function ShareAlbum(
    albumId: string,
    authHeader: AuthHeaderT,
    users?: string[],
    removeUsers?: string[]
) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`);
    const body = {
        users: users,
        removeUsers: removeUsers,
    };
    return fetch(url.toString(), {
        method: "PATCH",
        headers: authHeader,
        body: JSON.stringify(body),
    });
}

export async function SetAlbumCover(
    albumId,
    coverMediaId,
    authHeader: AuthHeaderT
) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`);
    const body = {
        cover: coverMediaId,
    };
    return fetch(url.toString(), {
        method: "PATCH",
        headers: authHeader,
        body: JSON.stringify(body),
    });
}

export async function RenameAlbum(albumId, newName, authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`);
    const body = {
        newName: newName,
    };
    return fetch(url.toString(), {
        method: "PATCH",
        headers: authHeader,
        body: JSON.stringify(body),
    })
        .then((res) => {
            if (res.status !== 200) {
                return res.json();
            }
        })
        .then((json) => {
            if (json?.error) {
                return Promise.reject(json.error);
            }
        })
        .catch((r) =>
            notifications.show({
                title: "Failed to rename album",
                message: String(r),
                color: "red",
            })
        );
}

export async function DeleteAlbum(albumId, authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`);

    let ret = await fetch(url.toString(), {
        method: "DELETE",
        headers: authHeader,
    });
    return ret;
}
