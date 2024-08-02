import API_ENDPOINT from './ApiEndpoint'
import { AuthHeaderT } from '../types/Types'

export function GetUsersInfo(setAllUsersInfo, authHeader: AuthHeaderT) {
    fetch(`${API_ENDPOINT}/users`, { headers: authHeader, method: 'GET' })
        .then((res) => {
            if (res.status !== 200) {
                return Promise.reject(
                    `Could not get user info list: ${res.statusText}`
                )
            } else {
                return res.json()
            }
        })
        .then((data) => setAllUsersInfo(data))
        .catch((r) => console.error(r))
}

export function ActivateUser(username: string, authHeader: AuthHeaderT) {
    return fetch(`${API_ENDPOINT}/user/${username}/activate`, {
        headers: authHeader,
        method: 'PATCH',
    })
}

export function DeleteUser(username: string, authHeader: AuthHeaderT) {
    return fetch(`${API_ENDPOINT}/user/${username}`, {
        headers: authHeader,
        method: 'DELETE',
    })
}

export function UpdatePassword(
    username: string,
    oldPassword: string,
    newPassword: string,
    authHeader: AuthHeaderT
) {
    return fetch(`${API_ENDPOINT}/user/${username}/password`, {
        method: 'PATCH',
        body: JSON.stringify({
            oldPassword: oldPassword,
            newPassword: newPassword,
        }),
        headers: authHeader,
    })
}

export async function SetUserAdmin(
    username: string,
    admin: boolean,
    authHeader: AuthHeaderT
) {
    return fetch(`${API_ENDPOINT}/user/${username}/admin`, {
        method: 'PATCH',
        body: JSON.stringify({ admin: admin }),
        headers: authHeader,
    })
}
