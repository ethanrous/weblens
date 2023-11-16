import API_ENDPOINT from "./ApiEndpoint"


function PromiseQueue(tasks = [], concurrentCount = 1) {
    this.total = tasks.length;
    this.todo = tasks;
    this.running = [];
    this.complete = [];
    this.count = concurrentCount;
}

PromiseQueue.prototype.runNext = function () {
    return ((this.running.length < this.count) && this.todo.length);
}

PromiseQueue.prototype.run = function () {
    while (this.runNext()) {
        const promiseFunc = this.todo.shift();
        // console.log(promise)
        promiseFunc().then(() => {
            this.complete.push(this.running.shift());
            this.run();
        });
        this.running.push(promiseFunc);
    }
}

const PostFile = (file64, fileName, path, authHeader, dispatch) => {
    const url = new URL(`${API_ENDPOINT}/file`)
    fetch(url.toString(), {
        method: "POST",
        body: JSON.stringify({
            file64: file64,
            fileName: fileName,
            path: path,
        }),
        headers: authHeader
    }).then(() => dispatch({ type: 'remove_from_upload_map', uploadName: fileName }))
}

function readFile(file, dispatch) {
    dispatch({ type: 'add_to_upload_map', uploadName: file.name })
    console.log("Startin")
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

const singleUploadPromise = (fileData, path, authHeader, dispatch) => {
    if (fileData.size > 2000000000) {
        console.log("This upload is going to fail")
    }
    return () => readFile(fileData, dispatch).then((value: { name: String, item64: string }) => { PostFile(value.item64, value.name, path, authHeader, dispatch) })
}

const Upload = (files, path, authHeader, dispatch) => {
    let uploads: (() => Promise<void>)[] = []

    for (const file of files) {
        uploads.push(singleUploadPromise(file, path, authHeader, dispatch))
    }
    const taskQueue = new PromiseQueue(uploads, 5)
    taskQueue.run()
}

export default Upload