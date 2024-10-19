import { fetchJson, wrapRequest } from '@weblens/api/ApiFetch'
import { UserInfoT } from '@weblens/types/Types'
import API_ENDPOINT from './ApiEndpoint'

export function GetUserInfo() {
    return fetchJson<UserInfoT>(`${API_ENDPOINT}/user`)
}

export function GetUsersInfo() {
    const url = `${API_ENDPOINT}/users`
    return fetchJson<UserInfoT[]>(url)
}

export function ActivateUser(username: string) {
    const url = `${API_ENDPOINT}/user/${username}/activate`
    return fetchJson(url, 'PATCH')
}

export function DeleteUser(username: string) {
    const url = `${API_ENDPOINT}/user/${username}`
    return wrapRequest(fetch(url, { method: 'DELETE' }))
}

export function UpdatePassword(
    username: string,
    oldPassword: string,
    newPassword: string
) {
    return wrapRequest(
        fetch(`${API_ENDPOINT}/user/${username}/password`, {
            method: 'PATCH',
            body: JSON.stringify({
                oldPassword: oldPassword,
                newPassword: newPassword,
            }),
        })
    )
}

export async function SetUserAdmin(username: string, admin: boolean) {
    return wrapRequest(
        fetch(`${API_ENDPOINT}/user/${username}/admin`, {
            method: 'PATCH',
            body: JSON.stringify({ admin: admin }),
        })
    )
}
