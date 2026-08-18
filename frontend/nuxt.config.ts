export default defineNuxtConfig({
  css: ['~/assets/css/main.css'],
  compatibilityDate: '2026-08-14',
  devtools: { enabled: true },

  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080',
    },
  },
})
