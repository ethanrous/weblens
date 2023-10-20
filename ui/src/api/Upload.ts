

const PostFiles = (files, path, wsSend) => {
    for (let file of files) {
        PostFile(file, path, wsSend)
    }
}

const PostFile = (file, path, wsSend) => {
    let msg = JSON.stringify({
        type: 'file_upload',
        content: {
            path: path,
            file: file
        },
    })
    wsSend(msg)
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

const HandleFileUpload = (filesData, path, wsSend) => {
    for (let file of filesData) {
        if (file.size > 2000000000) {
            console.log("This upload is going to fail")
        }
        readFile(file).then(value => PostFile(value, path, wsSend))
    }

}

export default HandleFileUpload