# Expected Outcome

- `createPracticePlan` maps optional client `roundId` and returns server-derived paired `roundId/roundSequence`.
- Real PostgreSQL persists the first incomplete canonical round and reads the same pair back.
- Equal-duration adjacent rounds remain distinct; duration is only an integrity check.
- Non-contiguous sequence `1,2,4` advances from `2` to the existing immediate successor `4`.
- Session start resolves the persisted pair to the real round name/type/focus; persona does not replace round context.
- Audit metadata contains the derived pair and excludes question / answer / hint / prompt / response text.
