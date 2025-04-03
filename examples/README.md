# Nebula Go SDK Examples

This directory contains example programs demonstrating how to use the `nebula-sdk-go`.

## Running Examples

1.  **Set Environment Variables:**
    The examples typically look for the `NEBULA_BASE_URL` environment variable to know where your Nebula BaaS instance is running. You might also need `NEBULA_TEST_EMAIL` and `NEBULA_TEST_PASSWORD` for the `basic_usage` example.

    ```bash
    # Example for Linux/macOS
    export NEBULA_BASE_URL="http://localhost:8080"
    export NEBULA_TEST_EMAIL="[email address removed]"
    export NEBULA_TEST_PASSWORD="your-test-password"
    ```

2.  **Navigate to Example Directory:**

    ```bash
    cd examples/basic_usage
    ```

3.  **Run the Go Program:**

    ```bash
    go run main.go
    ```

    _(Ensure the Nebula BaaS backend server specified by `NEBULA_BASE_URL` is running)_
