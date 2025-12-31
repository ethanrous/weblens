export class PortablePath {
    private rootAlias: string
    private relativePath: string[]
    private isDirectory: boolean = false

    constructor(path: string) {
        const parts = path.split(':')
        if (parts.length !== 2) {
            throw new Error(`Invalid portable path format: ${path}. Expected format is 'ROOT_ALIAS:path/to/file'`)
        }

        this.rootAlias = parts[0]!

        if (parts[1]!.endsWith('/')) {
            this.isDirectory = true
            parts[1] = parts[1]!.slice(0, -1) // Remove trailing slash
        }

        this.relativePath = parts[1]!.split('/')
    }

    public hasParent(parent: PortablePath): boolean {
        if (this.rootAlias !== parent.rootAlias) {
            return false
        }

        if (this.relativePath.length <= parent.relativePath.length) {
            return false
        }

        for (const [index, part] of parent.relativePath.entries()) {
            if (this.relativePath[index] !== part) {
                return false
            }
        }

        return true
    }

    public equals(other: PortablePath): boolean {
        if (this.rootAlias !== other.rootAlias || this.relativePath.length !== other.relativePath.length) {
            return false
        }

        return this.relativePath.join('/') === other.relativePath.join('/')
    }

    public isInTrash(): boolean {
        if (this.relativePath.length < 2) {
            return false
        }

        return this.relativePath[1] === '.user_trash'
    }

    public toString(opts?: { noHome: boolean }): string {
        if (opts?.noHome) {
            const path = [...this.relativePath]
            path.splice(0, 1)

            return `${path.join('/')}`
        }

        return `${this.rootAlias}:${this.relativePath.join('/')}`
    }

    public get filename(): string {
        const filename = this.relativePath[this.relativePath.length - 1]!
        if (filename === '.user_trash') {
            return 'Trash'
        }

        if (this.relativePath.length === 1 && filename === useUserStore().user.username) {
            return 'Home'
        }

        return filename
    }

    public get parts(): string[] {
        return [...this.relativePath]
    }

    static fromParts(rootAlias: string, relativePath: string[]): PortablePath {
        if (!rootAlias || !relativePath || relativePath.length === 0) {
            throw new Error('Root alias and relative path must be provided')
        }
        return new PortablePath(`${rootAlias}:${relativePath.join('/')}`)
    }

    static empty(): PortablePath {
        return new PortablePath(':')
    }
}
