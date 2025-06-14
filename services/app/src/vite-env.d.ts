/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_CLERK_PUBLISHABLE_KEY: string
  readonly VITE_API_BASE_URL: string
  readonly VITE_USE_MOCK_INTEGRATIONS: string
  readonly VITE_INTEGRATION_API_PREFIX: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}