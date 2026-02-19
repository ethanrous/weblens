import { spawn, execSync, type ChildProcess } from 'child_process'
import fs from 'fs'
import path from 'path'
import http from 'http'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const REPO_ROOT = path.resolve(__dirname, '../../..')
const BUILD_DIR = path.join(REPO_ROOT, '_build')
const PW_DIR = path.join(BUILD_DIR, 'playwright')
const LOG_DIR = path.join(BUILD_DIR, 'logs', 'playwright')

const VERBOSE = Boolean(process.env.WEBLENS_VERBOSE)

const WEBLENS_PORT_BASE = 14100
const MONGO_PORT_BASE = 27020
const MONGOT_GRPC_PORT_BASE = 27128
const MONGOT_METRICS_PORT_BASE = 10046

// ── Worker-scoped: MongoDB lifecycle ──

function makeLogFile(
    workerIndex: number,
    type: string,
    opts?: { extraIdentifier?: string; noExt?: boolean; noCreate?: boolean },
): string {
    let filename = opts?.extraIdentifier
        ? `worker-${workerIndex}-${type}-${opts.extraIdentifier}`
        : `worker-${workerIndex}-${type}`

    if (!opts?.noExt) {
        filename += '.log'
    }

    const logPath = path.join(LOG_DIR, type, filename)
    fs.mkdir(path.dirname(logPath), { recursive: true, mode: 0o777 }, (err) => {
        if (err) {
            console.error(`[worker-${workerIndex}] Failed to create log directory: ${err}`)
        }
    })
    // Clear existing log
    if (!opts?.noCreate) {
        console.debug(`[worker-${workerIndex}] Creating log file at ${logPath}...`)
        fs.closeSync(fs.openSync(logPath, 'w', 0o777))
    }

    return logPath
}

export interface WorkerMongo {
    port: number
    stackName: string
    containerName: string
    workerIndex: number
}

export async function startWorkerMongo(workerIndex: number): Promise<WorkerMongo> {
    const mongoPort = MONGO_PORT_BASE + workerIndex
    const mongotGrpcPort = MONGOT_GRPC_PORT_BASE + workerIndex
    const mongotMetricsPort = MONGOT_METRICS_PORT_BASE + workerIndex
    const stackName = `weblens-core-pw-worker-${workerIndex}`
    const containerName = `weblens-${stackName}-mongod`

    // Clean mongo data directory so MongoDB starts fresh.
    // Data dir may contain root-owned files from the container, so use
    // Docker to remove subdirs before cleaning up the parent with Node.
    const mongoDataDir = path.join(BUILD_DIR, 'db', stackName)
    if (fs.existsSync(mongoDataDir)) {
        try {
            fs.rmSync(mongoDataDir, { recursive: true })
        } catch {
            execSync(`docker run --rm -v "${mongoDataDir}:/cleanup" alpine rm -rf /cleanup/mongod /cleanup/mongot`, {
                cwd: REPO_ROOT,
                stdio: 'pipe',
            })
            fs.rmSync(mongoDataDir, { recursive: true })
        }
    }

    if (VERBOSE) console.debug(`[worker-${workerIndex}] Starting MongoDB on port ${mongoPort}...`)

    // Launch mongo via the bash helper (handles docker compose, replica set, etc.)
    execSync(
        `source scripts/lib/all.bash && ` +
            `MONGOT_HOST_PORT_GRPC=${mongotGrpcPort} ` +
            `MONGOT_HOST_PORT_METRICS=${mongotMetricsPort} ` +
            `launch_mongo --stack-name "${stackName}" --mongo-port ${mongoPort}`,
        { cwd: REPO_ROOT, stdio: 'pipe', shell: '/bin/bash' },
    )

    // Wait for mongo to be healthy (replica set init can take a while)
    if (VERBOSE) console.debug(`[worker-${workerIndex}] Waiting for MongoDB to become healthy...`)
    const healthStart = Date.now()
    const healthTimeout = 60_000
    while (Date.now() - healthStart < healthTimeout) {
        try {
            const result = execSync(`docker exec ${containerName} mongosh --quiet --eval "rs.status().ok"`, {
                stdio: 'pipe',
            })
            if (result.toString().trim() === '1') {
                break
            }
        } catch {
            // Not ready yet
        }
        await new Promise((r) => setTimeout(r, 1000))
    }

    if (VERBOSE) console.debug(`[worker-${workerIndex}] MongoDB is healthy`)

    return { port: mongoPort, stackName, containerName, workerIndex }
}

export async function stopWorkerMongo(mongo: WorkerMongo): Promise<void> {
    const { workerIndex, stackName } = mongo
    if (VERBOSE) console.debug(`[worker-${workerIndex}] Cleaning up mongo stack ${stackName}...`)

    const logPath = makeLogFile(workerIndex, 'db', { noExt: true, noCreate: true })

    if (VERBOSE) console.debug(`[worker-${workerIndex}] Dumping MongoDB logs to ${logPath}-*.log...`)

    try {
        const out = execSync(
            `source scripts/lib/all.bash && dump_mongo_logs --stack-name "${stackName}" --logfile "${logPath}" && cleanup_mongo --stack-name "${stackName}"`,
            {
                cwd: REPO_ROOT,
                stdio: 'pipe',
                shell: '/bin/bash',
            },
        )

        if (VERBOSE) console.debug(`[worker-${workerIndex}] Mongo cleanup log: ${out.toString()}`)
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (e: any) {
        if (VERBOSE)
            console.debug(
                `[worker-${workerIndex}] Mongo cleanup failed: ${e.stderr.toString()} \n------------\n ${e.stdout.toString()}`,
            )
    }
    if (VERBOSE) console.debug(`[worker-${workerIndex}] Mongo teardown complete`)
}

// ── Test-scoped: Backend binary lifecycle ──

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

let testCounter = 0

export async function startTestBackend(workerIndex: number, mongo: WorkerMongo): Promise<TestBackend> {
    const port = WEBLENS_PORT_BASE + workerIndex
    const dbName = `pw-test-${workerIndex}-${testCounter++}`

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

    const logPath = makeLogFile(workerIndex, 'backend', { extraIdentifier: dbName })

    if (VERBOSE) console.debug(`[worker-${workerIndex}] Starting backend (db: ${dbName}) on port ${port}...`)

    const logStream = fs.createWriteStream(logPath)

    // Spawn the binary directly (mongo is already running from worker fixture)
    const mongoUri = `mongodb://127.0.0.1:${mongo.port}/?replicaSet=rs0&directConnection=true`
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

    if (VERBOSE) console.debug(`[worker-${workerIndex}] Backend PID ${child.pid}, logs at ${logPath}`)

    const baseURL = `http://localhost:${port}`
    await pollHealth(`${baseURL}/api/v1/info`, 25_000, logPath)
    if (VERBOSE) console.debug(`[worker-${workerIndex}] Backend is healthy`)

    return { baseURL, port, dbName, workerIndex, process: child, logPath }
}

export async function stopTestBackend(backend: TestBackend): Promise<void> {
    const { workerIndex } = backend
    const pid = backend.process.pid

    if (pid) {
        if (VERBOSE) console.debug(`[worker-${workerIndex}] Stopping backend PID ${pid}...`)
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
                if (VERBOSE) console.debug(`[worker-${workerIndex}] Backend exited gracefully`)
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
