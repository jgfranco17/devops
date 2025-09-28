# Devops CLI

_DevOps, simplified._

![STATUS](https://img.shields.io/badge/status-active-brightgreen?style=for-the-badge)
![LICENSE](https://img.shields.io/badge/license-BSD3-blue?style=for-the-badge)

---

## Introduction

## Installation

To download the CLI, an install script has been provided.

```bash
go install github.com/jgfranco17/devops@latest
```

> [!NOTE]
> This CLI is still a beta prototype.

## Testing

```bash
# Run standard assertions with go-test
just test
```

### Automation

#### GitHub Actions Integration

Tests are automatically run on:

- Every pull request
- Every push to main branch
- Scheduled nightly runs

#### Quality Gates

- All tests must pass before merging
- Minimum code coverage requirements
- Performance benchmarks must be met

## License

This project is licensed under the BSD-3 License. See the LICENSE file for more details.
