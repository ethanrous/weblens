

const PostFile = (files, path, sendMessage) => {
    let msg = JSON.stringify({
        type: 'file_upload',
        content: {
            path: path,
            files: files
        },
    })
    sendMessage(msg)
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


const HandleFileUpload = (filesData, path, sendMessage) => {
    let readers = []

    for (let file of filesData) {
        readers.push(readFile(file))
    }

    Promise.all(readers).then((values) => { PostFile(values, path, sendMessage) })

}

export default HandleFileUpload