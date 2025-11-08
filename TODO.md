## Here are some fixes that I temporarily ignored:

Aside from the listed todos here, there are alot of inline TODOs throughout the codebase. Some might have been fixed & the comment wasnt removed sha... But alot might still be valid.

### App 
- Alot of errors are silent & arent being sent to the frontend
- Add error logging to capture and display errors in the UI
- When a user that already has an account, logins on a new device, add support to download all data from cloud. 
- If a user hasnt completed onboarding, dont allow them to be able to record. 
- Local sync_state valid with cloud has an error. it keeps pushing the last_synced_at local data to cloud even when there is no new data.
- Codebase structure is poorMake the codebase have a very good structure
- there's alot of duplicates between app & cloud server.. eg types. Unify things would be better, they can share one module
- Because sync data only runs every 30s, when a new board is created locally, it's gonna take 30s before it's on cloud, which doesnt really make sense, we should be able to create new boards in cloud instantly.
- implement soft delete
