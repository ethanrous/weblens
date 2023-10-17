import API_ENDPOINT from './ApiEndpoint'

export const fetchMetadata = (fileHash, setMediaData) => {
    var url = new URL(`${API_ENDPOINT}/item/${fileHash}`)
    url.searchParams.append('meta', 'true')

    fetch(url.toString()).then((res) => res.json()).then((data) => setMediaData(data))
}

export function fetchThumb64(fileHash, setMediaData: (thumb64: string) => void) {
    var url = new URL(`${API_ENDPOINT}/item/${fileHash}`)
    url.searchParams.append('thumbnail', 'true')

    fetch(url)
        .then(response => response.blob())
        .then(blob => new Promise((resolve, reject) => {
            const reader = new FileReader()
            reader.onloadend = () => resolve(reader.result)
            reader.onerror = reject
            reader.readAsDataURL(blob)
        }))
        .then(setMediaData)
}

export function fetchFullres64(fileHash, setMediaData: (fullres64: string) => void) {
    var url = new URL(`${API_ENDPOINT}/item/${fileHash}`)
    url.searchParams.append('fullres', 'true')

    fetch(url)
        .then(response => response.blob())
        .then(blob => new Promise((resolve, reject) => {
            const reader = new FileReader()
            reader.onloadend = () => resolve(reader.result)
            reader.onerror = reject
            reader.readAsDataURL(blob)
        }))
        .then(setMediaData)
}

export function login(user: string, pass: string) {
    var url = new URL(`${API_ENDPOINT}/login`)
    url.searchParams.append('user', user)
    url.searchParams.append('pass', pass)

    fetch(url, { method: "POST" }).then(res => res.json()).then(data => console.log(data))
}