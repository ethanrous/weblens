import { AuthHeaderT } from '../types/Types'
import API_ENDPOINT from '../api/ApiEndpoint'

export async function getFoldersMedia(
    folderIds: string[],
    authHeader: AuthHeaderT
): Promise<string[]> {
    if (folderIds.length === 0) {
        return []
    }
    const url = new URL(`${API_ENDPOINT}/folders/media`)
    url.searchParams.append('folderIds', JSON.stringify(folderIds))
    return fetch(url.toString(), {
        headers: authHeader,
    })
        .then((res) => res.json())
        .then((j) => j.medias)
        .catch((r) => console.error(r))
}
