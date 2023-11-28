const MAX_FILE_NAME_LENGTH = 80

function doUpload(file, handler) {
    const fileInfo = document.getElementById('file-info')
    fileInfo.innerHTML = `
                     ${file ? `<h4>Name: <code>${file.name.length > MAX_FILE_NAME_LENGTH ? `${file.name.substring(0, MAX_FILE_NAME_LENGTH)}...` : file.name}</code></h4>`:`<h4>No file selected</h4>`}
                     <label>
                            Automatic Deleted At: <input type="datetime-local">
                     </label>
                     <label>
                         Private Url Expires (Seconds): <input type="number" max="604801">
                    </label>
                     <p>File size: <code>${file.size} bytes</code></p>
                     <label>
                            Public: 
                            <input type="checkbox" name="public" id="public">
                    </label>
                    <button type="submit" class="btn-upload">Upload</button>
                    `
let metadata = {
        autoDeletedAt: Date.now(),
        isPublic: false,
        xPrivateUrlExpires: 10 // 10 seconds
}

    document.querySelector('#file-info input[type="datetime-local"]').addEventListener('input', ({currentTarget}) => {
        metadata = {
            ...metadata,
            autoDeletedAt: Date.parse(currentTarget.value)
        }
    })

    document.querySelector('#file-info input[type="checkbox"]').addEventListener('input', ({currentTarget}) => {
        metadata = {
            ...metadata,
            isPublic: currentTarget.checked
        }
    })

    document.querySelector('#file-info input[type="number"]').addEventListener('input', ({currentTarget}) => {
        metadata = {
            ...metadata,
            xPrivateUrlExpires: currentTarget.value
        }
    })

    document.querySelector('#file-info button').addEventListener('click', async (event) => {
        event.preventDefault()
        await handler(file, metadata)
    })
}

/**
 * @param {function} handler is a function that will be called when the user press the upload button
 */
export function doFile(handler) {
    const uploadBox = document.getElementById('file-upload-box')
    const input = uploadBox.querySelector('#input-file')

    input.addEventListener('input', ({currentTarget}) => doUpload(currentTarget.files[0], handler))

    uploadBox.addEventListener('dragover', (event) => {
        event.preventDefault()
        uploadBox.classList.add('drag')
    })

    uploadBox.addEventListener('dragleave', (event) => {
        event.preventDefault()
        uploadBox.classList.remove('drag')
    })

    uploadBox.addEventListener('drop', (event) => {
        event.preventDefault()
        doUpload(event.dataTransfer.files[0], handler)

        uploadBox.classList.remove('drag')
    })

    uploadBox.addEventListener('paste', ({clipboardData}) => doUpload(clipboardData.files[0], handler))
}