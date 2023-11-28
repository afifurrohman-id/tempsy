export function setupUserNav() {
    const btnToggleMenu = document.getElementById('btn-user-toggle-menu')
    if (!btnToggleMenu) return

    const navbar = document.querySelector('.main-nav')
    const navMenu = navbar.querySelector('#nav-menu')
    const links = navMenu.querySelectorAll('a')

    links.forEach(link=> {
        switch (true) {
            case new URL(link.href).pathname === location.pathname:
                link.classList.add('active')
                break
            default:
                link.classList.remove('active')
                break
        }
    })

    btnToggleMenu.addEventListener('click', () => navMenu.classList.toggle('open'))

    addEventListener('keydown', (event) => event.key === 'Escape' && navMenu.classList.remove('open'))

    addEventListener('click', (event) => !navMenu.contains(event.target) && !btnToggleMenu.contains(event.target) && navMenu.classList.remove('open'))

}