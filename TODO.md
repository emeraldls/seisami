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

```
failed uploading board 77459bbc-3592-4415-97f4-136adcd934cf: unable to upsert board: sync api returned status 500: {"message":"unable to parse board created_at value \"2025-11-08 08:48:06 +0100 WAT\""}
failed uploading board 579fa3d8-a8c0-44ae-a46a-8c4c239047ed: unable to upsert board: sync api returned status 500: {"message":"unable to parse board created_at value \"2025-11-08 18:41:48 +0100 WAT\""}
failed uploading board 0a7ad9c9-952a-43b9-a10d-5e8f74efe098: unable to upsert board: sync api returned status 500: {"message":"unable to parse board created_at value \"2025-11-08 19:27:54 +0100 WAT\""}
failed uploading column 4b91bbb2-0489-4beb-8241-4cd86fd3332d: unable to upsert column: sync api returned status 500: {"message":"unable to parse column created_at value \"2025-11-08 20:53:55 +0100 WAT\""}
failed uploading card 0485312e-635f-49d7-87fd-8a40a5f7f827: unable to upsert card: sync api returned status 500: {"message":"unable to parse card created_at value \"2025-11-08 20:53:59 +0100 WAT\""}
  ```