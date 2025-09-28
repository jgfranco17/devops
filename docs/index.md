# Home

Welcome to **devops**, a developer-focused command-line tool built in Go.
`devops` helps you simplify repetitive workflows by defining and executing tasks in a single,
declarative definitionfile.

---

## What is `devops`

`devops` is a **task runner CLI** that executes commands defined in a YAML file.
It gives teams and individuals a consistent way to manage project automation without scattering
scripts across multiple files, ensuring consistency and reproducibility between local and CI/CD
environments.

Key features include:

- A unified interface for running project tasks
- Built-in config management under `devops-definition.yaml`
- Structured logging with adjustable verbosity (`-v`, `-vv`, `-vvv`)
- Context-aware execution with safe cancellation (`SIGINT`/`SIGTERM`)
- Extensible command set powered by [Cobra](https://github.com/spf13/cobra)

---

## Why use `devops`

Modern development and operations often require juggling multiple tools, scripts, and environments.
Traditional shell scripts can become brittle, hard to maintain, and inconsistent across platforms.

`devops` solves this problem by:

- **Centralizing workflows** in a single `devops-definition.yaml` file
- **Ensuring cross-platform compatibility** via Go’s standard libraries
- **Providing safer execution** with built-in context handling
- **Scaling with your needs**: extend or customize as projects evolve

If you’ve ever used `make`, `npm run`, or `just`, `devops` gives you a similar experience — but
with modern ergonomics, strong typing, and structured logging.

---

## How it works

At its core, `devops` reads a **task definition file** (`devops-definition.yaml`) and executes
tasks step by step.

### Example

Below is an example of a `devops-definition.yaml` file you would create for your project.

```yaml title="devops-definition.yaml"
name: my-python-project
description: My example project written in Python
version: 0.0.2
repo_url: https://github.com/myuser/my-python-project

codebase:
  language: python
  dependencies:
    - requirements.txt
  install:
    steps:
      - pip install -r requirements.txt
  test:
    fail_fast: true
    steps:
      - python -m pytest -vv -m unit
  build:
    fail_fast: true
    env:
      FOO: BAR
    steps:
      - echo "Build completed successfully with $FOO"
```
