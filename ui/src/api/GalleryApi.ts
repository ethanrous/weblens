import { AlbumData, MediaData } from '../types/Types'
import API_ENDPOINT from './ApiEndpoint'

export async function FetchData(mediaState, dispatch, authHeader) {
    if (!authHeader || authHeader.Authorization === "") { dispatch({ type: "set_loading", loading: false }); return }
    // console.log("FETCHIN")
    try {
        const url = new URL(`${API_ENDPOINT}/media`)
        url.searchParams.append('raw', mediaState.includeRaw.toString())
        const data = await fetch(url.toString(), { headers: authHeader }).then(res => {
            if (res.status != 200) { return Promise.reject('Failed to get media') }
            else { return res.json() }
        })

        dispatch({
            type: 'set_media',
            media: data.Media,
        })
    } catch (error) {
        console.error("Error fetching data - Ethan you wrote this, its not a js err")
    }
}

export async function CreateAlbum(albumName, authHeader) {
    const url = new URL(`${API_ENDPOINT}/album`)
    const body = {
        name: albumName
    }
    await fetch(url.toString(), { method: "POST", headers: authHeader, body: JSON.stringify(body) })
}

export async function GetAlbums(authHeader): Promise<AlbumData[]> {
    const url = new URL(`${API_ENDPOINT}/albums`)
    const res = await (await fetch(url.toString(), { headers: authHeader })).json()
    return res.albums
}


export async function GetAlbumMedia(albumId, dispatch, authHeader): Promise<{ albumMeta: AlbumData, media: MediaData[] }> {
    dispatch({ type: 'set_loading', loading: true })
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`)
    const res = await fetch(url.toString(), { headers: authHeader })
    if (res.status != 200) {
        return Promise.reject("Album not found")
    }

    const data = await res.json()
    return { albumMeta: data.albumMeta, media: data.medias }
}

// export async function AddMediaToAlbum(albumId, mediaIds, authHeader) {
//     const url = new URL(`${API_ENDPOINT}/album`)
//     const body = {
//         albumId: albumId,
//         media: mediaIds
//     }
//     await fetch(url.toString(), { method: "PUT", headers: authHeader, body: JSON.stringify(body) })
// }

export async function ShareAlbum(albumId: string, users: string[], authHeader) {
    const url = new URL(`${API_ENDPOINT}/share`)
    const body = {
        shareType: 'album',
        content: albumId,
        users: users
    }
    return fetch(url.toString(), { method: "POST", headers: authHeader, body: JSON.stringify(body) })
}

export async function SetAlbumCover(albumId, coverMediaId, authHeader) {
    const url = new URL(`${API_ENDPOINT}/album/${albumId}`)
    const body = {
        cover: coverMediaId
    }
    return fetch(url.toString(), { method: "PUT", headers: authHeader, body: JSON.stringify(body) })

}