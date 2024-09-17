import API_ENDPOINT from '@weblens/api/ApiEndpoint'
import { fetchJson } from '@weblens/api/ApiFetch'

export async function getFoldersMedia(folderIds: string[]): Promise<string[]> {
    if (folderIds.length === 0) {
        return []
    }
    const url = new URL(`${API_ENDPOINT}/folders/media`)
    url.searchParams.append('folderIds', JSON.stringify(folderIds))
    return fetch(url.toString(), {})
        .then((res) => res.json())
        .then((j) => j.medias)
        .catch((r) => console.error(r))
}

export async function restoreFiles(
    fileIds: string[],
    newParentId: string,
    time: Date
): Promise<{ newParentId: string }> {
    return fetchJson<{ newParentId: string }>(
        `${API_ENDPOINT}/files/restore`,
        'POST',
        {
            fileIds: fileIds,
            newParentId: newParentId,
            timestamp: time.getTime(),
        }
    )
}
