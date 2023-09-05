import IconButton from "@mui/material/IconButton"
import UploadIcon from '@mui/icons-material/Upload'
import Stack from "@mui/material/Stack"

const PostFile = (file, item64) => {
    console.log("UPLOADING")
    var url = new URL("http:localhost:3000/api/item")
    fetch(url, {
        method: "POST",
        body: JSON.stringify({
            filename: file.name,
            item64: item64
        }),
        headers: {
            "Content-type": "application/json; charset=UTF-8"
        }
    })
    console.log("done")
}

const FileUpload = () => {
    const HandleFileUpload = (event) => {
        const file = event.target.files[0]
        const reader = new FileReader()

        reader.onloadend = () => {
            PostFile(file, reader.result)
        }

        console.log(file)

        reader.readAsDataURL(file)
    }

    return (

        <Stack >
            <IconButton
                color="inherit"
                aria-label="upload picture"
                component="label"
            >
                <input
                    id="upload-image"
                    hidden
                    accept="image/*"
                    type="file"
                    onChange={HandleFileUpload}
                />
                <UploadIcon style={{ paddingRight: "10px" }} />
                {"Upload"}
            </IconButton>
        </Stack>

    )
}

export default FileUpload