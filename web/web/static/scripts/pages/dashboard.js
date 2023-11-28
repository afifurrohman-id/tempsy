import {setupUserNav} from '../components/user-nav.js'
import {setupTheme} from '../utils/theme.js'
import {doFile} from '../components/upload.js'

setupUserNav()
setupTheme()

doFile(async (file, metadata) => {
    const res = await fetch(location.href, {
        method: 'POST',
        headers: {
            'Content-Type': file.type,
            'X-File-Auto-Deleted-At': metadata.autoDeletedAt,
            'X-File-Is-Public': metadata.isPublic,
            'X-File-Name': file.name,
            'X-File-Private-Url-Expires': metadata.xPrivateUrlExpires,
        },
        body: file,
    })


    if (res.ok) {
        alert('Upload successfully')
        location.reload()
    } else {
        const {error_description} = await res.json()
        confirm(error_description+ ', Do you want to refresh the page?') && location.reload()
    }
})