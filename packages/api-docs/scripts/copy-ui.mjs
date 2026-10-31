import { copyFileSync, cpSync } from 'node:fs'
import { createRequire } from 'node:module'
import { join } from 'node:path'

const require = createRequire(import.meta.url)
const uiDirectory = require('swagger-ui-dist').getAbsoluteFSPath()
const outputDirectory = 'dist'

for (const file of [
  'swagger-ui.css',
  'swagger-ui-bundle.js',
  'swagger-ui-standalone-preset.js',
  'favicon-32x32.png',
]) {
  copyFileSync(join(uiDirectory, file), join(outputDirectory, file))
}

cpSync('public', outputDirectory, { recursive: true })
