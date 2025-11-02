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

The project still doesnt have a very good structure yet... there are copies of alot of things between server & app... There are also some stuffs that shouldnt be where they are....

--- Some rando thoughts

You can think of a board like a workspace. Then within the board you have columns (or lists) that contain cards (or tasks). Each card can have multiple features (or subtasks). 

A board also has transcriptions associated with it.