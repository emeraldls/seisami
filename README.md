# **Seisami**

**Voice-driven task management for people who think faster than they type.**

* Capture tasks instantly by tapping the system `Fn` key
* AI groups everything into shared Kanban boards in real time
* Works offline by default with optional cloud sync
* Invite teammates, share boards, and assign tasks, all with real-time sync when connected (in progress)

---

 **Status:** *Actively in development*. Expect frequent updates, rough edges, and missing features.
If something breaks, please [open an issue](https://github.com/emeraldls/seisami/issues) or reach out. Your feedback helps shape the product.

---

## **Try It Locally**

Clone the repository, install Wails locally on your machine, then to start the app in development mode, run:

```bash
cd app && wails dev
```

To start the cloud server locally, run:

```bash
    cd server && go run main.go
```

To start the website locally, run:

```bash
    cd web && pnpm dev
```

Ensure you've installed necessary dependencies for the server and web projects.

Desktop binaries live in `build/bin/` when you run `wails build`.

---

Contributions welcome.
