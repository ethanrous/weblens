import axios from "axios";
import API_ENDPOINT from "./ApiEndpoint";
import { notifications } from "@mantine/notifications";

const userApiUrl = `${API_ENDPOINT}/user`
const adminUserApiUrl = `${API_ENDPOINT}/admin/user`
const adminUsersApiUrl = `${API_ENDPOINT}/admin/users`

export function GetUsersInfo(setAllUsersInfo, authHeader) {
    fetch(adminUsersApiUrl, { headers: authHeader, method: "GET" })
        .then(res => { if (res.status != 200) { return Promise.reject(`Could not get user info list: ${res.statusText}`) } else { return res.json() } })
        .then(data => setAllUsersInfo(data))
        .catch(r => notifications.show({ message: r, color: 'red' }))
}

export function ActivateUser(username: string, authHeader: { Authentication: string }) {
    return axios.post(adminUserApiUrl, JSON.stringify({ username: username }), { headers: authHeader })
}

export function DeleteUser(username: string, authHeader: { Authentication: string }) {
    return axios.delete(`${adminUserApiUrl}/${username}`, { headers: authHeader })
}