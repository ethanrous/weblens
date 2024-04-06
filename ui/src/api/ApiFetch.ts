import { notifications } from "@mantine/notifications";
import API_ENDPOINT, { ADMIN_ENDPOINT } from "./ApiEndpoint";
import { AuthHeaderT, MediaData } from "../types/Types";

export function login(user: string, pass: string) {
    var url = new URL(`${API_ENDPOINT}/login`);
    let data = {
        username: user,
        password: pass,
    };

    return fetch(url.toString(), {
        method: "POST",
        body: JSON.stringify(data),
    });
}

export function createUser(username, password) {
    const url = new URL(`${API_ENDPOINT}/user`);
    return fetch(url, {
        method: "POST",
        body: JSON.stringify({ username: username, password: password }),
    }).then((res) => {
        if (res.status !== 201) {
            return Promise.reject(`${res.statusText}`);
        }
    });
}

export function adminCreateUser(username, password, admin, authHeader?: AuthHeaderT) {
    const url = new URL(`${ADMIN_ENDPOINT}/user`);
    return fetch(url, {
        headers: authHeader,
        method: "POST",
        body: JSON.stringify({
            username: username,
            password: password,
            admin: admin,
            autoActivate: true,
        }),
    })
        .then((res) => {
            if (res.status !== 201) {
                return Promise.reject(`${res.statusText}`);
            }
        })
        .catch((r) => {
            notifications.show({
                message: `Failed to create new user: ${String(r)}`,
            });
        });
}

export function clearCache(authHeader) {
    return fetch(`${ADMIN_ENDPOINT}/cache`, {
        method: "POST",
        headers: authHeader,
    });
}

export function cleanMedias(authHeader) {
    return fetch(`${ADMIN_ENDPOINT}/cleanup/medias`, {
        method: "POST",
        headers: authHeader,
    });
}

export async function getMedia(mediaId, authHeader: AuthHeaderT): Promise<MediaData> {
    if (!mediaId) {
        console.error("trying to get media with no mediaId");
        notifications.show({
            title: "Failed to get media",
            message: "no media id provided",
            color: "red",
        });
        return;
    }
    const url = new URL(`${API_ENDPOINT}/media/${mediaId}`);
    url.searchParams.append("meta", "true");
    const mediaMeta: MediaData = await fetch(url, { headers: authHeader }).then((r) => r.json());
    return mediaMeta;
}

export async function fetchMediaTypes() {
    const url = new URL(`${API_ENDPOINT}/media/types`);
    return await fetch(url).then((r) => r.json());
}

export async function newApiKey(authHeader) {
    const url = new URL(`${ADMIN_ENDPOINT}/apiKey`);
    return await fetch(url, {headers: authHeader, method: "POST"}).then((r) => r.json());
}

export async function getApiKeys(authHeader) {
    const url = new URL(`${ADMIN_ENDPOINT}/apiKeys`);
    return await fetch(url, {headers: authHeader}).then((r) => r.json());
}