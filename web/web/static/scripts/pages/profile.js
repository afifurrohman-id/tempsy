import {setupTheme} from "../utils/theme.js";
import {setupUserNav} from "../components/user-nav.js";

setupTheme()
setupUserNav()

const deleteBtn = document.getElementById('delete-btn')

deleteBtn.addEventListener('click', async () => {
    if (confirm('Are you sure?, this action cannot be undone')) {

            const res = await fetch(location.href, {
                method: 'DELETE'
            })

            if (res.ok) {
                if (confirm('Delete successful, do you want to Logout as well?')) location.href = '/auth/logout'
                else location.reload()
            } else {
                const { error_description } = await res.json()
                confirm(error_description+ ', Do you want to refresh the page?') && location.reload()
        }
    }
})
