# Contributing to KraftKit 🚀🐒🧰

Thank you for your interest in contributing to KraftKit! KraftKit is an open-source project and we welcome contributors of all skill levels.

## How to Contribute

1.  **Fork the Repository**: Create a fork of the `unikraft/kraftkit` repository.
2.  **Clone the Fork**: `git clone https://github.com/YOUR_USERNAME/kraftkit.git`
3.  **Create a Feature Branch**: `git checkout -b feature/your-feature-name`
4.  **Make Changes**: Implement your changes and ensure they follow the project's coding standards.
5.  **Commit Changes**: `git commit -m "feat: your feature description"`
6.  **Push to Fork**: `git push origin feature/your-feature-name`
7.  **Open a Pull Request**: Submit your PR to the `staging` branch of the `unikraft/kraftkit` repository.

## Developing Build-time Environment Variables

We recently introduced a layered environment configuration system for Unikraft builds. If you are working on features related to this, please ensure:
*   Changes maintain the layering priority: CLI > ENV FILE > PROFILE > CONFIG FILE > DEFAULT.
*   New environment variables are validated appropriately in `internal/cli/kraft/build/env.go`.
*   Debug logging is included to verify propagation to underlying toolchains.

## Code Standards

*   Use `go fmt` to format your code.
*   Write clear, concise commit messages.
*   Include tests for new features where possible.

## Community

Join us on [Discord](https://bit.ly/UnikraftDiscord) in the `#kraftkit` channel to discuss development and get help from the maintainers.

---
KraftKit is part of the [Unikraft OSS Project](https://unikraft.org) and licensed under `BSD-3-Clause`.
