import type { TaskInfo } from '@ethanrous/weblens-api'

export type TaskTreeNode = {
    task: TaskInfo
    children: TaskTreeNode[]
    completedCount: number
    totalCount: number
    percent: number
}

export type VisibleTaskRow = {
    task: TaskInfo
    depth: number
    hasChildren: boolean
    completedCount: number
    totalCount: number
    percent: number
}

// buildTaskTree groups a flat task list into a forest by parentTaskID. Roots are
// tasks whose parent is not present in the list (true roots plus orphans whose
// parent has already exited). Sibling order follows the input order. Each node's
// completedCount/totalCount/percent come from the task's own completedChildTasks/
// totalChildTasks (API response counts), so they hold even when child rows are absent.
export function buildTaskTree(tasks: TaskInfo[]): TaskTreeNode[] {
    const nodes = new Map<string, TaskTreeNode>()
    for (const task of tasks) {
        nodes.set(task.taskID, { task, children: [], completedCount: 0, totalCount: 0, percent: 0 })
    }

    const roots: TaskTreeNode[] = []
    for (const node of nodes.values()) {
        const parent = node.task.parentTaskID ? nodes.get(node.task.parentTaskID) : undefined
        if (parent && parent !== node) {
            parent.children.push(node)
        } else {
            roots.push(node)
        }
    }

    for (const node of nodes.values()) {
        node.totalCount = node.task.totalChildTasks ?? 0
        node.completedCount = node.task.completedChildTasks ?? 0
        node.percent = node.totalCount === 0 ? 0 : Math.round((node.completedCount / node.totalCount) * 100)
    }

    return roots
}

// flattenVisible walks the forest depth-first into the rows currently visible,
// descending into a node's children only when it is expanded. A visited guard
// keeps a malformed cycle from looping forever.
export function flattenVisible(roots: TaskTreeNode[], expanded: Set<string>): VisibleTaskRow[] {
    const rows: VisibleTaskRow[] = []
    const seen = new Set<string>()

    const walk = (nodes: TaskTreeNode[], depth: number) => {
        for (const node of nodes) {
            if (seen.has(node.task.taskID)) {
                continue
            }

            // Skip queued descendant nodes (and with them their subtrees) entirely. This prevents the
            // page from trying to render possibly hundreds of child rows that don't provide much value.
            if (depth > 0 && node.task.State === 'InQueue') {
                continue
            }

            seen.add(node.task.taskID)

            rows.push({
                task: node.task,
                depth,
                hasChildren: node.children.length > 0,
                completedCount: node.completedCount,
                totalCount: node.totalCount,
                percent: node.percent,
            })

            if (node.children.length > 0 && expanded.has(node.task.taskID)) {
                walk(node.children, depth + 1)
            }
        }
    }

    walk(roots, 0)
    return rows
}
