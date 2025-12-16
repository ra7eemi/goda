import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Goda',
  description: 'Discord Wrapper API for Zwafriya for Go.',

  locales: {
    root: {
      label: 'English',
      lang: 'en',
      link: '/',
      themeConfig: {
        nav: [
          { text: 'Home', link: '/' },
          { text: 'Introduction', link: 'introduction/what-is-goda' }
        ],
        sidebar: [
          {
            text: 'Introduction',
            items: [
              { text: 'What is Goda?', link: 'introduction/what-is-goda' },
              { text: 'Getting Started', link: 'introduction/getting-started' }
            ]
          }
        ]
      }
    },
    fr: {
      label: 'Français',
      lang: 'fr',
      link: '/fr/',
      themeConfig: {
        nav: [
          { text: 'Accueil', link: '/fr/' },
          { text: 'Introduction', link: '/fr/introduction/what-is-goda' }
        ],
        sidebar: [
          {
            text: 'Introduction',
            items: [
              { text: 'Qu\'est-ce que Goda ?', link: '/fr/introduction/what-is-goda' },
              { text: 'Guide de démarrage', link: '/fr/introduction/getting-started' }
            ]
          }
        ]
      }
    }
  }
})
