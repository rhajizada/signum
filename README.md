# signum

signum is a badge generator that produces clean SVG status badges for READMEs, documentation, and CI/CD pipelines. It includes a CLI for local generation, a Go renderer package (pkg/renderer) for embedding in your own tools, and a self-hosted API server for creating and serving hosted badges you can update from CI.

- **CLI**: generate badges locally (stdout or file output), with configurable subject, status, color, style, and font.
- **Library**: use `pkg/renderer` as a small, reusable rendering package inside your own applications.
- **API**: run the HTTP server to create stored badges (returns an id + token), render them as SVG, and update or delete them securely—plus a “live” endpoint for one-off badge rendering without storage.

signum is built on top of [`narqo/go-badge`](https://github.com/narqo/go-badge).
