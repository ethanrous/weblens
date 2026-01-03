# Writing tests best practices for Weblens

When writing tests in Go, it is a best practice to use a separate `_test` package instead of the main package. This approach has several advantages:

1. **Isolation**: By using a separate `_test` package, you ensure that your tests only have access to the exported functions and types of the main package. This helps to simulate how users of your package will interact with it, leading to more realistic and reliable tests.
2. **Avoiding Circular Dependencies**: Using a `_test` package helps to avoid circular dependencies that can arise when tests are placed in the same package as the code being tested.
3. **Cleaner Code**: It encourages better organization of test code, making it easier to maintain and understand.
4. **Improved Test Coverage**: By testing only the exported API, you can focus on the public interface of your package, which is often more relevant for users of that package.

When it comes time to run your tests, ALWAYS use the script at `./scripts/test-weblens.bash`. This script is designed to set up the appropriate environment and run your tests consistently across different systems. If you wish to run tests specifically for a single package, you can use the command `./scripts/test-weblens.bash <path/to/package/...>`, replacing `<path/to/package...>` with the name of the package you want to test. For example:

```bash
./scripts/test-weblens.bash ./models/task/...
```

This command will run all tests in the specified package and its sub-packages, ensuring that your tests are executed in the correct context.

The test script will always output coverage information when run. To check code coverage of the most recently run test, you can use the command:

```bash
make cover
```
