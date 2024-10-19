import { ServerInfoT } from '@weblens/types/Types'
import API_ENDPOINT from './ApiEndpoint'
import { wrapRequest } from './ApiFetch'

export async function launchRestore(
    serverInfo: ServerInfoT,
    restoreUrl: string
) {
    const body = {
        restoreId: serverInfo.id,
        restoreUrl: restoreUrl,
    }
    return wrapRequest(
        fetch(`${API_ENDPOINT}/restore`, {
            method: 'POST',
            body: JSON.stringify(body),
        })
    )
}
