export class Optional<T> {
    private value?: T
    private hasValue: boolean

    constructor(value?: T) {
        this.value = value
        this.hasValue = value !== undefined
    }

    isSet(): boolean {
        return this.hasValue
    }

    get(defaultVal?: T): T {
        if (!this.hasValue) {
            if (defaultVal !== undefined) {
                return defaultVal
            }

            throw new Error('No value present')
        }

        return this.value!
    }

    set(value: T): void {
        if (value === undefined) {
            this.clear()

            return
        }

        this.value = value
        this.hasValue = true
    }

    clear(): void {
        this.value = undefined
        this.hasValue = false
    }

    static some<T>(value: T): Optional<T> {
        return new Optional<T>(value)
    }

    static none<T>(): Optional<T> {
        return new Optional<T>()
    }
}
