export type Task<R = void, A = R> = (args: A) => Promise<R>

export default class TaskQueue {
    queue: Task[] = []
    groupQueue: Map<string, Task[]> = new Map()
    noCollide: (string | undefined)[] = []
    activeRunnerCount: number = 0
    maxConcurrentRunners: number = 1
    waitingGroups: string[] = []

    constructor(maxConcurrentRunners?: number) {
        this.maxConcurrentRunners = maxConcurrentRunners ?? 1
    }

    async addTask(task: Task, taskGroupID?: string) {
        if (taskGroupID !== undefined) {
            let groupQueue = this.groupQueue.get(taskGroupID)
            if (!groupQueue) {
                groupQueue = []
            }

            groupQueue.push(task)
            this.groupQueue.set(taskGroupID, groupQueue)
        } else {
            this.queue.push(task)
        }

        if (taskGroupID && this.noCollide.includes(taskGroupID)) {
            return
        }

        if (this.activeRunnerCount < this.maxConcurrentRunners) {
            this.noCollide[this.activeRunnerCount] = taskGroupId
            this.activeRunnerCount++
            await this.run(this.activeRunnerCount - 1, taskGroupID)
        } else if (taskGroupID) {
            this.waitingGroups.push(taskGroupID)
        }
    }

    async run(runnerID: number, taskGroupID?: string) {
        const task = this.getTask(taskGroupID)
        if (!task) {
            if (this.waitingGroups.length !== 0) {
                taskGroupID = this.waitingGroups.shift()
                await this.run(runnerID, taskGroupID)
                return
            } else {
                this.activeRunnerCount--
                this.noCollide[runnerID] = undefined
                return
            }
        }

        try {
            await task()
        } catch (err) {
            console.error('Task failed:', err)
        } finally {
            await this.run(runnerID, taskGroupID)
        }
    }

    private async runWithNextRunner<T>({
        getNext,
        onFailure,
        initCtx,
    }: {
        getNext: (ctx: T) => Task<T, void> | undefined
        onFailure?: (err: Error) => void
        initCtx: T
    }) {
        let ctx: T = initCtx
        while (true) {
            const task = getNext(ctx)
            if (!task) {
                break // Exit if no more tasks are available
            }

            try {
                ctx = await task()
            } catch (err) {
                console.error('Task failed:', err)
                onFailure?.(err as Error)
                break // Exit on error
            }
        }
    }

    public async runWithNext<T>({
        getNext,
        initCtx,
        onFailure,
        groupID,
    }: {
        getNext: (ctx: T) => Task<T, void> | undefined
        initCtx: T
        onFailure?: (err: Error) => void
        groupID?: string
    }) {
        await this.addTask(() => this.runWithNextRunner({ getNext, onFailure, initCtx }), groupID)
    }

    private getTask(taskGroupID?: string): Task | undefined {
        if (taskGroupID) {
            return this.groupQueue.get(taskGroupID)?.shift()
        }

        return this.queue.shift()
    }
}
