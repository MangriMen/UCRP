export const buildInfo = {
  version: import.meta.env.VITE_APP_VERSION ?? 'dev',
  buildTime: import.meta.env.VITE_BUILD_TIME ?? 'unknown',
}
