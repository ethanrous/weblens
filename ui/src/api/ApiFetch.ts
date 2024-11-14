// export async function wrapRequest(rq: Promise<Response>): Promise<Response> {
//     return await rq
//         .then((r: Response) => {
//             if (r.status >= 400) {
//                 return Promise.reject(new Error(r.statusText))
//             }
//             return r
//         })
//         .catch((e) => {
//             if (e === 401) {
//                 const user = new User()
//                 useSessionStore.getState().setUser(user)
//
//                 if (window.location.pathname.startsWith('/files/share')) {
//                     console.log('Got 401 on share page')
//                     return
//                 }
//
//                 console.debug('Got 401, going to login')
//                 useSessionStore.getState().nav('/login', {
//                     state: { returnTo: window.location.pathname },
//                 })
//             }
//             return Promise.reject(e)
//         })
// }
//
// export async function fetchJson<T>(
//     url: string,
//     method?: string,
//     body?: object
// ): Promise<T> {
//     if (!method) {
//         method = 'GET'
//     }
//     const init: RequestInit = {
//         method: method,
//     }
//
//     if (body) {
//         init.body = JSON.stringify(body)
//     }
//
//     return await wrapRequest(fetch(url, init)).then((r) => {
//         return r.json()
//     })
// }

// export function clearCache() {
//     return wrapRequest(
//         fetch(`${API_ENDPOINT}/cache`, {
//             method: 'POST',
//         })
//     )
// }

// export async function resetServer() {
//     return wrapRequest(fetch(`${API_ENDPOINT}/reset`, { method: 'POST' }))
// }
