import { notifications } from '@mantine/notifications';
import API_ENDPOINT from '../api/ApiEndpoint';
import WeblensMedia, { MediaDataT } from '../Media/Media';
import { AlbumData, AuthHeaderT } from '../types/Types';

export async function getAlbums(includeShared: Boolean, authHeader: AuthHeaderT): Promise<AlbumData[]> {
    const url = new URL(`${API_ENDPOINT}/albums`);
    url.searchParams.append('includeShared', includeShared.toString());

    return await fetch(url, { headers: authHeader })
        .then(r => r.json())
        .then(j => j.albums)
        .catch(r => console.error(r));
}

export async function createAlbum(albumName: string, authHeader: AuthHeaderT) {
    if (albumName === '') {
        return Promise.reject('No album title');
    }

    const url = new URL(`${API_ENDPOINT}/album`);
    const body = {
        name: albumName,
    };
    await fetch(url.toString(), {
        method: 'POST',
        headers: authHeader,
        body: JSON.stringify(body),
    });
}

export async function getAlbumMedia(
    albumId: string,
    includeRaw: boolean,
    authHeader: AuthHeaderT,
): Promise<{ albumMeta: AlbumData; media: WeblensMedia[] }> {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`);
    url.searchParams.append('raw', includeRaw.toString());
    const res = await fetch(url.toString(), { headers: authHeader });
    if (res.status === 404) {
        return Promise.reject(404);
    } else if (res.status !== 200) {
        return Promise.reject('Unknown error');
    }

    const data = await res.json();
    const medias = data.medias?.map((m: MediaDataT) => {
        return new WeblensMedia(m);
    });
    return { albumMeta: data.albumMeta, media: medias };
}

export async function addMediaToAlbum(
    albumId: string,
    mediaIds: string[],
    folderIds: string[],
    authHeader: AuthHeaderT,
) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`);
    const body = {
        newMedia: mediaIds,
        newFolders: folderIds,
    };
    return (
        await fetch(url.toString(), {
            method: 'PATCH',
            headers: authHeader,
            body: JSON.stringify(body),
        })
    ).json();
}

export async function RemoveMediaFromAlbum(albumId: string, mediaIds: string[], authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`);
    const body = {
        removeMedia: mediaIds,
    };
    await fetch(url.toString(), {
        method: 'PATCH',
        headers: authHeader,
        body: JSON.stringify(body),
    });
}

export async function ShareAlbum(albumId: string, authHeader: AuthHeaderT, users?: string[], removeUsers?: string[]) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`);
    const body = {
        users: users,
        removeUsers: removeUsers,
    };
    return fetch(url.toString(), {
        method: 'PATCH',
        headers: authHeader,
        body: JSON.stringify(body),
    });
}

export async function LeaveAlbum(albumId: string, authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}/leave`);
    return fetch(url, { method: 'POST', headers: authHeader });
}

export async function SetAlbumCover(albumId: string, coverMediaId: string, authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`);
    const body = {
        cover: coverMediaId,
    };
    return fetch(url.toString(), {
        method: 'PATCH',
        headers: authHeader,
        body: JSON.stringify(body),
    });
}

export async function RenameAlbum(albumId: string, newName: string, authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`);
    const body = {
        newName: newName,
    };
    return fetch(url.toString(), {
        method: 'PATCH',
        headers: authHeader,
        body: JSON.stringify(body),
    })
        .then(res => {
            if (res.status !== 200) {
                return res.json();
            }
        })
        .then(json => {
            if (json?.error) {
                return Promise.reject(json.error);
            }
        })
        .catch(r =>
            notifications.show({
                title: 'Failed to rename album',
                message: String(r),
                color: 'red',
            }),
        );
}

export async function DeleteAlbum(albumId: string, authHeader: AuthHeaderT) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`);

    return await fetch(url.toString(), {
        method: 'DELETE',
        headers: authHeader,
    });
}