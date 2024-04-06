import axios from "axios";
import API_ENDPOINT, { ADMIN_ENDPOINT } from "./ApiEndpoint";
import { notifications } from "@mantine/notifications";
import { AuthHeaderT, UserInfoT } from "../types/Types";

const userApiUrl = `${API_ENDPOINT}/user`;
const adminUserApiUrl = `${API_ENDPOINT}/admin/user`;
const adminUsersApiUrl = `${API_ENDPOINT}/admin/users`;

export function GetUsersInfo(setAllUsersInfo, authHeader: AuthHeaderT) {
    fetch(adminUsersApiUrl, { headers: authHeader, method: "GET" })
        .then((res) => {
            if (res.status !== 200) {
                return Promise.reject(`Could not get user info list: ${res.statusText}`);
            } else {
                return res.json();
            }
        })
        .then((data) => setAllUsersInfo(data))
        .catch((r) => notifications.show({ message: String(r), color: "red" }));
}

export function ActivateUser(username: string, authHeader: AuthHeaderT) {
    return axios.post(adminUserApiUrl, JSON.stringify({ username: username }), { headers: authHeader });
}

export function DeleteUser(username: string, authHeader: AuthHeaderT) {
    return axios.delete(`${adminUserApiUrl}/${username}`, { headers: authHeader });
}

export function UpdatePassword(username: string, oldPassword: string, newPassword: string, authHeader: AuthHeaderT) {
    return fetch(`${userApiUrl}/${username}/password`, {
        method: "PATCH",
        body: JSON.stringify({ oldPassword: oldPassword, newPassword: newPassword }),
        headers: authHeader,
    });
}

export async function SetUserAdmin(username: string, admin: boolean, authHeader: AuthHeaderT) {
    return fetch(`${ADMIN_ENDPOINT}/user/${username}/admin`, {
        method: "PATCH",
        body: JSON.stringify({ admin: admin }),
        headers: authHeader,
    });
}
