import { spawnSync } from 'node:child_process'
import {
  existsSync,
  mkdtempSync,
  readdirSync,
  rmSync,
  rmdirSync,
  symlinkSync,
  unlinkSync,
} from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join } from 'node:path'

const [source, ...command] = process.argv.slice(2)

if ((source !== 'index' && source !== 'head') || command.length === 0) {
  console.error('usage: bun scripts/run-on-tree.ts {index|head} <command> [args...]')
  process.exit(2)
}

const git = (args: string[], env: NodeJS.ProcessEnv = process.env) => {
  const result = spawnSync('git', args, { encoding: 'utf8', env })
  if (result.status !== 0) {
    throw new Error(`git ${args.join(' ')} failed: ${result.stderr}`)
  }
  return result.stdout.trim()
}

const repoRoot = git(['rev-parse', '--show-toplevel'])
const snapshot = mkdtempSync(join(tmpdir(), 'ucrp-tree-'))
const links: string[] = []

const unlink = (link: string) => {
  try {
    unlinkSync(link)
  } catch {
    rmdirSync(link)
  }
}

const cleanup = () => {
  for (const link of links) {
    try {
      unlink(link)
    } catch {
      console.error(`run-on-tree: could not remove link ${link}, leaving ${snapshot} in place`)
      return
    }
  }
  rmSync(snapshot, { recursive: true, force: true })
}

const prefix = `${snapshot}/`

try {
  if (source === 'index') {
    git(['-C', repoRoot, 'checkout-index', '--all', `--prefix=${prefix}`])
  } else {
    const indexFile = join(snapshot, '.run-on-tree-index')
    const env = { ...process.env, GIT_INDEX_FILE: indexFile }
    git(['-C', repoRoot, 'read-tree', 'HEAD'], env)
    git(['-C', repoRoot, 'checkout-index', '--all', `--prefix=${prefix}`], env)
    rmSync(indexFile, { force: true })
  }

  const moduleDirs = ['node_modules']
  for (const group of ['apps', 'packages']) {
    const groupDir = join(repoRoot, group)
    if (!existsSync(groupDir)) continue
    for (const entry of readdirSync(groupDir)) moduleDirs.push(join(group, entry, 'node_modules'))
  }
  for (const relative of moduleDirs) {
    const target = join(repoRoot, relative)
    if (!existsSync(target)) continue
    const link = join(snapshot, relative)
    if (!existsSync(dirname(link))) continue
    symlinkSync(target, link, 'junction')
    links.push(link)
  }

  const result = spawnSync(command[0], command.slice(1), { cwd: snapshot, stdio: 'inherit' })
  if (result.error) console.error(result.error.message)
  process.exitCode = result.status ?? 1
} finally {
  cleanup()
}
