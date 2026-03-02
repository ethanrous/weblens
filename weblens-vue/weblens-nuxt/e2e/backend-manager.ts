import { spawn, type ChildProcess } from 'child_process'
import fs from 'fs'
import path from 'path'
import http from 'http'
import { fileURLToPath } from 'url'
import { randomInt } from 'crypto'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const REPO_ROOT = path.resolve(__dirname, '../../..')
const BUILD_DIR = path.join(REPO_ROOT, '_build')
const PW_DIR = path.join(BUILD_DIR, 'playwright')
const LOG_DIR = path.join(BUILD_DIR, 'logs', 'playwright')

const VERBOSE = Boolean(process.env.WEBLENS_VERBOSE)

const WEBLENS_PORT_BASE = 10100
const MONGO_PORT = 27020

export function makeLogFile(
    workerIndex: number,
    logClass: string,
    name: string,
    opts?: { noCreate?: boolean },
): string {
    const filename = `${logClass}-${name.replaceAll('/', '_').replaceAll(' ', '_')}.log`

    const dateString = new Date(Number(process.env.WEBLENS_PLAYWRIGHT_TEST_START_TIME))
        .toISOString()
        .replace('T', '_')
        .replace(/:/g, '-')
        .split('.')[0]

    const logPath = path.join(LOG_DIR, `pw-test-${dateString}`, filename)
    fs.mkdirSync(path.dirname(logPath), { recursive: true, mode: 0o777 })
    // Clear existing log
    if (!opts?.noCreate) {
        if (VERBOSE) console.debug(`[worker-${workerIndex}] Creating log file at ${logPath}...`)
        fs.closeSync(fs.openSync(logPath, 'w', 0o777))
    }

    return logPath
}

export interface TestBackend {
    baseURL: string
    port: number
    dbName: string
    workerIndex: number
    process: ChildProcess
    logPath: string
}

function pollHealth(url: string, timeoutMs: number, logPath?: string): Promise<void> {
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
                reject(new Error(`Backend did not become healthy within ${timeoutMs}ms. Check logs at ${logPath}`))
                return
            }
            setTimeout(check, 500)
        }

        check()
    })
}

function sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms))
}

function isProcessRunning(pid: number): boolean {
    try {
        process.kill(pid, 0)
        return true
    } catch {
        return false
    }
}

export async function startTestBackend(workerIndex: number, testName: string): Promise<TestBackend> {
    const port = WEBLENS_PORT_BASE + workerIndex * 1000 + randomInt(999)
    const dbName = `pw-${testName.replaceAll('/', '_').replaceAll('.', '_')}`.slice(0, 63) // MongoDB database names have a max length of 64

    // Fresh filesystem per test
    const fsDir = path.join(PW_DIR, 'fs', `worker-${workerIndex}`)
    if (fs.existsSync(fsDir)) {
        fs.rmSync(fsDir, { recursive: true })
    }
    fs.mkdirSync(path.join(fsDir, 'data'), { recursive: true })
    fs.mkdirSync(path.join(fsDir, 'cache'), { recursive: true })

    // Validate build artifacts
    const binaryPath = path.join(BUILD_DIR, 'bin', 'weblens_debug')
    if (!fs.existsSync(binaryPath)) {
        throw new Error(`Backend binary not found at ${binaryPath}. Run ./scripts/test-playwright.bash to build it.`)
    }

    const uiPath = path.join(REPO_ROOT, 'weblens-vue', 'weblens-nuxt', '.output', 'public')
    if (!fs.existsSync(uiPath)) {
        throw new Error(`Frontend build not found at ${uiPath}. Run ./scripts/test-playwright.bash to build it.`)
    }

    const logPath = makeLogFile(workerIndex, 'backend', dbName)

    const logStream = fs.createWriteStream(logPath)

    // Spawn the binary directly (mongo is already running from worker fixture)
    const mongoUri = `mongodb://127.0.0.1:${MONGO_PORT}/?replicaSet=rs0&directConnection=true`
    const child = spawn(binaryPath, [], {
        cwd: REPO_ROOT,
        env: {
            ...process.env,
            WEBLENS_PORT: String(port),
            WEBLENS_MONGODB_URI: mongoUri,
            WEBLENS_MONGODB_NAME: dbName,
            WEBLENS_DATA_PATH: path.join(fsDir, 'data'),
            WEBLENS_CACHE_PATH: path.join(fsDir, 'cache'),
            WEBLENS_UI_PATH: uiPath,
            WEBLENS_LOG_LEVEL: 'trace',
            WEBLENS_LOG_FORMAT: 'dev',
            WEBLENS_DO_CACHE: 'false',
            WEBLENS_DO_PROFILING: 'false',
            // Auto-initialize with a default admin user to avoid manual setup in tests
            WEBLENS_INIT_ROLE: 'core',
            // Use low bcrypt cost for faster tests, not to be used in production!
            WEBLENS_USE_DANGEROUSLY_INSECURE_PASSWORD_HASHING: 'true',
        },
        stdio: ['ignore', 'pipe', 'pipe'],
        detached: true,
    })

    child.stdout?.pipe(logStream)
    child.stderr?.pipe(logStream)
    child.unref()

    if (!child.pid) {
        throw new Error(`[worker-${workerIndex}] Failed to spawn backend process`)
    }

    const start = Date.now()
    const baseURL = `http://localhost:${port}`
    await pollHealth(`${baseURL}/health`, 25_000, logPath)
    if (VERBOSE)
        console.debug(
            `Backend for test ${testName} with PID ${child.pid} on port :${port} is healthy after ${Date.now() - start}ms - logs at ${logPath}`,
        )

    return { baseURL, port, dbName, workerIndex, process: child, logPath }
}

export async function stopTestBackend(backend: TestBackend): Promise<void> {
    const { workerIndex } = backend
    const pid = backend.process.pid

    if (pid) {
        try {
            process.kill(pid, 'SIGTERM')
        } catch {
            if (VERBOSE) console.debug(`[worker-${workerIndex}] Process already exited`)
            return
        }

        // Wait up to 5s for graceful shutdown
        for (let i = 0; i < 50; i++) {
            await sleep(100)
            if (!isProcessRunning(pid)) {
                return
            }
        }

        // Force kill if still running
        if (VERBOSE) console.debug(`[worker-${workerIndex}] Backend did not exit in time, sending SIGKILL...`)
        try {
            process.kill(pid, 'SIGKILL')
        } catch {
            // Already exited
        }
    }
}
