## Pending Fixes & Improvements

Aside from the items below, there are several inline TODOs scattered throughout the codebase. Some may already be resolved but not yet cleaned up; others are still valid and should be reviewed.

---

### App

* Many errors are currently silent and not surfaced to the frontend.
  → Add proper error propagation and UI logging.

* Implement centralized error logging and display errors directly in the UI for visibility.

* When a returning user logs in on a new device, automatically download all their existing cloud data to ensure full sync. (✅ Fixed)

* Prevent users who haven’t completed onboarding from recording or creating new content.

* Fix the `sync_state` validation logic - currently, it keeps pushing the `last_synced_at` value from local storage to the cloud even when no new data exists. (✅ Fixed)

* Improve overall codebase structure and organization for better readability and maintainability.

* Reduce duplication between the app and cloud server (e.g., shared types).
  → Extract and unify common modules or type definitions.

* Improve sync frequency: newly created boards should sync to the cloud instantly rather than waiting for the 30-second sync interval.

* Implement soft delete functionality for recoverable item deletion.

* Fix import logic: board import currently fails due to unexpected date format.

* Eliminate duplicate websocket handling — currently, board updates are sent both via the app backend and websocket, causing redundant writes.
  → Directly write changes to the websocket connection.

* Tie sync state management to a dedicated board for better consistency tracking.

* During board import, ensure that both board data and its associated sync state are imported together.
* Fix timestamp issue between local & cloud:
* Wrote a couple test cases for `/app`, need to write more covering all the components and edge cases. Then write test cases for `/server` as well.
* Fix, recording & syncing works when no board is selected.
* Fix performance issue when large number of items are present in a board. (fixed ✅)
* add more tools, add full calendar support


* When there's no column on the board & you're using cloud for processing, create the column locally doesnt work, the tool is not being called in the cloud, fix later.