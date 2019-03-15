document.addEventListener('DOMContentLoaded', () => {
    const THEME = 'ui.theme';
    const THEME_DARK = 'ui-dark';
    const SWITCHER_ID = 'theme-switcher';
    const currentTheme = localStorage.getItem(THEME);

    if (currentTheme == 'ui-dark') {
        document.body.classList.add(currentTheme);
    }

    const switcher = document.getElementById(SWITCHER_ID);
    if (!switcher) {
        console.warn(`Warning: theme switcher control not found`);
        return
    }

    switcher.addEventListener('click', () => {
        const classList = document.body.classList;
        classList.toggle(THEME_DARK);
        if (classList.contains(THEME_DARK)) {
            localStorage.setItem(THEME, THEME_DARK);
            return;
        }

        localStorage.removeItem(THEME);
    })
});