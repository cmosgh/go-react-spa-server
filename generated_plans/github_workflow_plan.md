# GitHub Workflow Plan: Build, Test, and E2E Test Go Server (Step-by-Step)

This document details a step-by-step plan for creating a GitHub Actions workflow to automatically build, unit test, and run end-to-end (e2e) tests for your Go server application. Each step includes an explanation of its purpose.

## 1. Workflow File Location

**Explanation:** GitHub Actions workflows are defined in YAML files located in the `.github/workflows/` directory of your repository. This standard location allows GitHub to automatically discover and run your workflows.

**Action:** Create a new file named `go-ci.yml` inside the `.github/workflows/` directory.

## 2. Workflow Name

**Explanation:** The `name` field provides a human-readable title for your workflow, which will be displayed in the GitHub Actions UI. This helps in easily identifying the workflow among others.

**Action:** Add the following at the top of your `go-ci.yml` file:

```yaml
name: Go CI
```

## 3. Workflow Triggers

**Explanation:** The `on` field specifies when the workflow should run. We will configure it to trigger on two common events:
*   `push`: When changes are pushed to the `main` branch. This ensures that every new commit on the main development line is built and tested.
*   `pull_request`: When a pull request is opened, synchronized, or reopened targeting the `main` branch. This helps ensure that proposed changes are valid before they are merged.

**Action:** Add the following after the `name` field:

```yaml
on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main
```

## 4. Define Jobs

**Explanation:** Workflows are composed of one or more `jobs`. A job is a set of `steps` that execute on the same runner. We will define two jobs:
*   `build-and-test`: To build the Go server and run its unit tests.
*   `e2e-test`: To run the end-to-end tests, leveraging the `run-e2e-tests.sh` script which handles server startup and client build.

**Action:** Add the following after the `on` field:

```yaml
jobs:
```

## 5. Job: `build-and-test`

**Explanation:** This job will handle the compilation of your Go server and the execution of its unit tests.

**Action:** Add the following under `jobs:`:

```yaml
  build-and-test:
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.22.x' # Or your preferred Go version

    - name: Build Go Application
      run: go build -v ./...

    - name: Run Go Tests
      run: go test -v ./...
```

### Step 5.1: Checkout Code

**Explanation:** This step uses the `actions/checkout@v4` action, which is a standard GitHub Action that checks out your repository code onto the runner. This is essential for subsequent steps to access your project files.

### Step 5.2: Set up Go Environment

**Explanation:** This step uses the `actions/setup-go@v5` action to install a specific version of the Go SDK on the runner. It ensures that the correct Go tools are available for building and testing. Specifying a version like `1.22.x` (or your preferred stable version) is good practice to ensure consistent builds.

### Step 5.3: Build Go Application

**Explanation:** This step executes the `go build` command. The `-v` flag enables verbose output, and `./...` tells Go to build all packages within the current module. This compiles your Go source code into an executable.

### Step 5.4: Run Go Tests

**Explanation:** This step executes the `go test` command. Similar to `go build`, `-v` provides verbose output, and `./...` instructs Go to run all tests found in all packages within the current module. This verifies the correctness of your code.

## 6. Job: `e2e-test`

**Explanation:** This job will run the end-to-end tests. Since the `run-e2e-tests.sh` script handles building the client, starting the server, and running the tests, this job will primarily focus on setting up the necessary environments and executing that script.

**Action:** Add the following after the `build-and-test` job:

```yaml
  e2e-test:
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.22.x' # Ensure Go is available for 'go run' in the script

    - name: Set up Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '20' # Ensure Node.js is available for npm/npx in the script

    - name: Run E2E Tests Script
      run: chmod +x run-e2e-tests.sh && ./run-e2e-tests.sh
```

### Step 6.1: Checkout Code

**Explanation:** This checks out the repository code to access the `run-e2e-tests.sh` script and the client-side code.

### Step 6.2: Set up Go Environment

**Explanation:** The `run-e2e-tests.sh` script uses `go run main.go` to start the server, so the Go environment needs to be set up on the runner.

### Step 6.3: Set up Node.js

**Explanation:** The `run-e2e-tests.sh` script uses `npm install`, `npm run build`, and `npx playwright test`, so the Node.js environment needs to be set up on the runner.

### Step 6.4: Run E2E Tests Script

**Explanation:** This step first makes the `run-e2e-tests.sh` script executable (`chmod +x`) and then executes it. This script will handle the entire e2e test orchestration, including building the client, starting the server, and running the Playwright tests.

## 7. Complete `go-ci.yml` Example

**Explanation:** This is the complete YAML content for your GitHub Actions workflow file, combining all the steps and jobs outlined above.

**Action:** Ensure your `go-ci.yml` file contains the following exact content:

```yaml
name: Go CI

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main

jobs:
  build-and-test:
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.22.x' # Or your preferred Go version

    - name: Build Go Application
      run: go build -v ./...

    - name: Run Go Tests
      run: go test -v ./...

  e2e-test:
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.22.x' # Ensure Go is available for 'go run' in the script

    - name: Set up Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '20' # Ensure Node.js is available for npm/npx in the script

    - name: Run E2E Tests Script
      run: chmod +x run-e2e-tests.sh && ./run-e2e-tests.sh
```