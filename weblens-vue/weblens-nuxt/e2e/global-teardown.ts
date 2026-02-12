import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const PID_FILE = path.resolve(__dirname, '../../../_build/playwright/.backend-pid')

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

export default async function globalTeardown() {
    if (!fs.existsSync(PID_FILE)) {
        console.debug('No backend PID file found, skipping teardown')
        return
    }

    const pid = parseInt(fs.readFileSync(PID_FILE, 'utf-8').trim(), 10)
    if (isNaN(pid)) {
        console.debug('Invalid PID in file, skipping teardown')
        fs.unlinkSync(PID_FILE)
        return
    }

    console.debug(`Sending SIGTERM to backend (PID ${pid})...`)
    try {
        process.kill(pid, 'SIGTERM')
    } catch {
        console.debug('Process already exited')
        fs.unlinkSync(PID_FILE)
        return
    }

    // Wait up to 2 seconds for graceful shutdown
    for (let i = 0; i < 20; i++) {
        await sleep(100)
        if (!isProcessRunning(pid)) {
            console.debug('Backend exited gracefully')
            fs.unlinkSync(PID_FILE)
            return
        }
    }

    // Force kill
    console.debug('Backend did not exit in time, sending SIGKILL...')
    try {
        process.kill(pid, 'SIGKILL')
    } catch {
        // Already exited
    }

    fs.unlinkSync(PID_FILE)
    console.debug('Backend killed')
}
