import API_ENDPOINT from '@weblens/api/ApiEndpoint'

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
