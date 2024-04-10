import { notifications } from "@mantine/notifications";
import API_ENDPOINT, { ADMIN_ENDPOINT, PUBLIC_ENDPOINT } from "./ApiEndpoint";
import { AuthHeaderT, MediaDataT } from "../types/Types";

export function login(user: string, pass: string) {
    var url = new URL(`${PUBLIC_ENDPOINT}/login`);
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

export function clearCache(authHeader: AuthHeaderT) {
    return fetch(`${ADMIN_ENDPOINT}/cache`, {
        method: "POST",
        headers: authHeader,
    });
}

export async function getMedia(mediaId, authHeader: AuthHeaderT): Promise<MediaDataT> {
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
    const mediaMeta: MediaDataT = await fetch(url, { headers: authHeader }).then((r) => r.json());
    return mediaMeta;
}

export async function fetchMediaTypes() {
    const url = new URL(`${PUBLIC_ENDPOINT}/media/types`);
    return await fetch(url).then((r) => r.json());
}

export async function newApiKey(authHeader: AuthHeaderT) {
    const url = new URL(`${ADMIN_ENDPOINT}/apiKey`);
    return await fetch(url, {headers: authHeader, method: "POST"}).then((r) => r.json());
}

export async function deleteApiKey(key: string, authHeader: AuthHeaderT) {
    const url = new URL(`${ADMIN_ENDPOINT}/apiKey`);
    return await fetch(url, {headers: authHeader, method: "DELETE", body: JSON.stringify({key: key})});
}

export async function getApiKeys(authHeader: AuthHeaderT) {
    const url = new URL(`${ADMIN_ENDPOINT}/apiKeys`);
    return await fetch(url, {headers: authHeader}).then((r) => r.json());
}

export async function getRandomThumbs() {
    const url = new URL(`${PUBLIC_ENDPOINT}/media/random`);
    url.searchParams.append("count", "50")
    return await fetch(url).then((r) => r.json());
}

export async function initServer(serverName: string, role: "core" | "backup", username: string, password: string, coreAddress: string, coreKey: string) {
    const url = new URL(`${PUBLIC_ENDPOINT}/initialize`);
    const body = {name: serverName, role: role, username: username, password: password, coreAddress: coreAddress, coreKey: coreKey}
    return await fetch(url,{ body: JSON.stringify(body), method: "POST"});
}

export async function getServerInfo() {
    const url = new URL(`${PUBLIC_ENDPOINT}/info`);
    return await fetch(url).then(r => {if (r.status !== 200) {return r.status} else {return r.json()}})
}

export async function getUsersPublic() {
    const url = new URL(`${PUBLIC_ENDPOINT}/users`);
    return await fetch(url).then(r => {if (r.status !== 200) {return r.status} else {return r.json()}})
}

export async function doBackup(authHeader: AuthHeaderT) {
    const url = new URL(`${ADMIN_ENDPOINT}/backup`);
    return await fetch(url, {method: "POST", headers: authHeader}).then(r => {if (r.status !== 200) {return r.status} else {return r.json()}})
}

export async function getRemotes(authHeader: AuthHeaderT) {
    const url = new URL(`${ADMIN_ENDPOINT}/remotes`);
    return await fetch(url, {method: "GET", headers: authHeader}).then(r => {if (r.status !== 200) {return r.status} else {return r.json()}})
}

export async function deleteRemote(remoteId: string, authHeader: AuthHeaderT) {
    const url = new URL(`${ADMIN_ENDPOINT}/remote`);
    return await fetch(url, {method: "DELETE", headers: authHeader, body: JSON.stringify({remoteId: remoteId})})
}