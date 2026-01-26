# Writing tests best practices for Weblens

When writing tests in Go, it is a best practice to use a separate `_test` package instead of the main package. This approach has several advantages:

1. **Isolation**: By using a separate `_test` package, you ensure that your tests only have access to the exported functions and types of the main package. This helps to simulate how users of your package will interact with it, leading to more realistic and reliable tests.
2. **Avoiding Circular Dependencies**: Using a `_test` package helps to avoid circular dependencies that can arise when tests are placed in the same package as the code being tested.
3. **Cleaner Code**: It encourages better organization of test code, making it easier to maintain and understand.
4. **Improved Test Coverage**: By testing only the exported API, you can focus on the public interface of your package, which is often more relevant for users of that package.

Write tests that cover large swaths of functionality, focusing on edge cases and error handling. Use table-driven tests to cover multiple scenarios in a concise manner. Writing fewer and smaller tests is encouraged, as long as they effectively cover the functionality being tested. This approach helps to keep the test suite manageable and easier to maintain in the future. Do not use custom types or helper functions in your tests unless absolutely necessary, as this can add unnecessary complexity and make the tests harder to read, understand, and refactor.

## What NOT to test

DO NOT write tests for the following, as they provide no value and cannot catch bugs that the compiler wouldn't already detect:

- **Constant definitions**: Do not test that constants are defined or have specific values (e.g., `assert.Equal(t, "GLOBAL", GlobalTaskPoolID)`)
- **Default struct values**: Do not test default zero values of struct fields (e.g., testing that a boolean field defaults to `false`)
- **Simple struct assignments**: Do not test that setting a field to a value results in that field having that value (e.g., `opts.Persistent = true; assert.True(t, opts.Persistent)`)
- **Error variable existence**: Do not test that error variables are defined or contain specific strings (e.g., `assert.NotNil(t, ErrTaskError)`)
- Private functions or unexported methods: Since these are not part of the public API, testing them directly is discouraged. Focus on testing the public interface instead.

Instead, focus on testing **behavior** and **logic** that can have bugs:

- Functions that transform data
- Error handling paths and edge cases
- Concurrent access and race conditions
- Integration between components
- State transitions and side effects

When it comes time to run your tests, ALWAYS use the script at `./scripts/test-weblens.bash`. This script is designed to set up the appropriate environment and run your tests consistently across different systems. If you wish to run tests specifically for a single package, you can use the command `./scripts/test-weblens.bash <path/to/package/...>`, replacing `<path/to/package...>` with the name of the package you want to test. For example:

```bash
./scripts/test-weblens.bash ./models/task/...
```

This command will run all tests in the specified package and its sub-packages, ensuring that your tests are executed in the correct context.

The test script will always output coverage information when run. To check code coverage of the most recently run test, you can use the command:

```bash
make cover
```

# Code style and linting

Weblens follows standard Go code style conventions. To ensure your code adheres to these conventions, please run the following commands after making changes to go code:

```bash
golangci-lint run ./...
```

And run this commend for changes made to vue or typescript code:

```bash
pnpm run lint
```
