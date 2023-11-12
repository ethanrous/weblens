import API_ENDPOINT from "./ApiEndpoint"


// const PostFiles = (files, path, wsSend) => {
//     for (let file of files) {
//         PostFile(file, path, wsSend)
//     }
// }

const PostFile = (file64, fileName, path, wsSend, authHeader) => {
    const url = new URL(`${API_ENDPOINT}/file`)
    fetch(url.toString(), {
        method: "POST",
        body: JSON.stringify({
            file64: file64,
            fileName: fileName,
            path: path,
        }),
        headers: authHeader
    })
}

function readFile(file) {
    return new Promise(function (resolve, reject) {
        let fr = new FileReader();

        fr.onload = function () {
            resolve({ name: file.name, item64: fr.result });
        };

        fr.onerror = function () {
            reject(fr);
        };

        fr.readAsDataURL(file);
    });
}

const HandleFileUpload = (fileData, path, wsSend, authHeader) => {
    if (fileData.size > 2000000000) {
        console.log("This upload is going to fail")
    }
    return readFile(fileData).then((value: { name: String, item64: string }) => { PostFile(value.item64, value.name, path, wsSend, authHeader) })
}


export default HandleFileUpload