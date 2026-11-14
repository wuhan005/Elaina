<p align="center"><img src="web/public/logo.png" alt="Elaina logo" width="96" /></p>
<h1 align="center">Elaina</h1>
<p align="center">A WebAssembly-powered remote code runner for C, Go, PHP, Python, and JavaScript.</p>
<p align="center"><a href="README_zh.md">简体中文</a></p>

Create a sandbox and share its link. Anyone with the link can run code in the browser or call `POST /r/{uid}/execute`. See the [API spec](docs/swagger.yaml).

## Quick start

Send this to your agent.

```text
Set up and run Elaina on this machine (https://github.com/wuhan005/Elaina). If the repository is already open, use the current checkout.

Check the build dependencies and PostgreSQL. Run scripts/download-wasm.sh and scripts/build-wasm-go.sh to set up the runtimes, then build the frontend with pnpm and the server with Go. If there's no config yet, use config/elaina.example.yaml to set the database connection and the bcrypt hash for the initial admin password.

Start Elaina, check that the page loads, then tell me where to open it and how to sign in.
```

[MIT License](LICENSE)
