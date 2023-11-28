import {setupTheme} from "../utils/theme.js";
import {setupUserNav} from "../components/user-nav.js";
import {doFile} from "../components/upload.js";

setupTheme()
setupUserNav()

const uploadArea = document.querySelector('.upload-area')
uploadArea.setAttribute('open', '')

doFile(async (file, metadata) => {
    const res = await fetch(location.href, {
        method: 'PUT',
        body: file,
        headers: {
            'Content-Type': file.type,
            'X-File-Private-Url-Expires': metadata.xPrivateUrlExpires,
            'X-File-Is-Public': metadata.isPublic,
            'X-File-Auto-Deleted-At': metadata.autoDeletedAt
        }
    })

    if (res.ok) {
        alert('Updated File successfully')
        location.reload()
    } else {
        const {error_description} = await res.json()
        confirm(error_description+ ' Do you want to refresh the page?') && location.reload()
    }
})

{
    const deleteBtn = document.querySelector('.delete-btn')
    deleteBtn.addEventListener('click', async () => {
        if (confirm('Are you sure?, this action cannot be undone')) {
            const res = await fetch(location.href, {
                method: 'DELETE'
            })

            if (res.ok) {
                const username = location.pathname.split('/')[2]
                location.href = '/dashboard/' + username
            } else {
                const {error_description} = await res.json()
                confirm(error_description+ ', Do you want to refresh the page?') && location.reload()
            }
        }
    })
}