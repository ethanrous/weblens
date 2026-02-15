import { spawn, execSync } from 'child_process'
import fs from 'fs'
import path from 'path'
import http from 'http'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const REPO_ROOT = path.resolve(__dirname, '../../..')
const BUILD_DIR = path.join(REPO_ROOT, '_build')
const PW_DIR = path.join(BUILD_DIR, 'playwright')
const PID_FILE = path.join(PW_DIR, '.backend-pid')
const FS_DIR = path.join(PW_DIR, 'fs')
const PORT = process.env.WEBLENS_TEST_PORT ?? '14100'
const MONGO_URI = process.env.WEBLENS_MONGODB_URI ?? 'mongodb://127.0.0.1:27019/?replicaSet=rs0&directConnection=true'
const MONGO_CONTAINER = 'weblens-playwright-test-mongod'
const DB_NAME = 'playwright-test'

function pollHealth(url: string, timeoutMs: number): Promise<void> {
    const start = Date.now()
    return new Promise((resolve, reject) => {
        const check = () => {
            const req = http.get(url, (res) => {
                if (res.statusCode === 200) {
                    resolve()
                } else {
                    retry()
                }
            })
            req.on('error', retry)
            req.setTimeout(2000, () => {
                req.destroy()
                retry()
            })
        }

        const retry = () => {
            if (Date.now() - start > timeoutMs) {
                reject(new Error(`Backend did not become healthy within ${timeoutMs}ms`))
                return
            }
            setTimeout(check, 500)
        }

        check()
    })
}

export default async function globalSetup() {
    // Clean filesystem directory
    if (fs.existsSync(FS_DIR)) {
        fs.rmSync(FS_DIR, { recursive: true })
    }
    fs.mkdirSync(FS_DIR, { recursive: true })
    fs.mkdirSync(path.join(FS_DIR, 'data'), { recursive: true })
    fs.mkdirSync(path.join(FS_DIR, 'cache'), { recursive: true })

    // Ensure PID dir exists
    fs.mkdirSync(PW_DIR, { recursive: true })

    // Drop the test database
    try {
        execSync(
            `docker exec ${MONGO_CONTAINER} mongosh --quiet --eval "db.getSiblingDB('${DB_NAME}').dropDatabase()"`,
            { stdio: 'pipe' },
        )
        console.debug(`Dropped database '${DB_NAME}'`)
    } catch {
        console.debug(`Could not drop database '${DB_NAME}' (may not exist yet)`)
    }

    const binaryPath = path.join(BUILD_DIR, 'bin', 'weblens_debug')
    if (!fs.existsSync(binaryPath)) {
        throw new Error(`Backend binary not found at ${binaryPath}. Run ./scripts/test-playwright.bash to build it.`)
    }

    const uiPath = path.join(REPO_ROOT, 'weblens-vue', 'weblens-nuxt', '.output', 'public')
    if (!fs.existsSync(uiPath)) {
        throw new Error(`Frontend build not found at ${uiPath}. Run ./scripts/test-playwright.bash to build it.`)
    }

    // Spawn the backend
    const child = spawn(binaryPath, [], {
        env: {
            ...process.env,
            WEBLENS_PORT: PORT,
            WEBLENS_MONGODB_URI: MONGO_URI,
            WEBLENS_MONGODB_NAME: DB_NAME,
            WEBLENS_DATA_PATH: path.join(FS_DIR, 'data'),
            WEBLENS_CACHE_PATH: path.join(FS_DIR, 'cache'),
            WEBLENS_UI_PATH: uiPath,
            WEBLENS_LOG_LEVEL: 'debug',
            WEBLENS_DO_CACHE: 'false',
        },
        stdio: ['ignore', 'pipe', 'pipe'],
        detached: true,
    })

    // Log backend output to file for debugging
    const logPath = path.join(PW_DIR, 'backend.log')
    const logStream = fs.createWriteStream(logPath)
    child.stdout.pipe(logStream)
    child.stderr.pipe(logStream)

    child.unref()

    if (!child.pid) {
        throw new Error('Failed to spawn backend process')
    }

    fs.writeFileSync(PID_FILE, child.pid.toString())
    console.debug(`Backend spawned with PID ${child.pid}, logs at ${logPath}`)

    // Poll until healthy
    console.debug(`Waiting for backend at http://localhost:${PORT}/api/v1/info ...`)
    await pollHealth(`http://localhost:${PORT}/api/v1/info`, 30000)
    console.debug('Backend is healthy')
}
