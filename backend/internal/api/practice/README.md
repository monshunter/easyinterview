# Practice API

Practice is a continuous text conversation. A session owns an ordered `messages` list; clients append ordinary user messages through `POST /practice/sessions/{sessionId}/messages`, and the server returns the persisted user message, assistant reply, and refreshed session snapshot.

`clientMessageId` is the message replay key. The service reserves the user message before the AI call, performs `practice.session.chat` outside the repository transaction, then commits the assistant reply. Replaying the same key and text returns the original result; reusing the key with different text fails with a typed conflict.

Completion remains idempotent through `Idempotency-Key` and starts asynchronous `report_generate` processing. The voice endpoint stays mounted only as an explicit fail-closed boundary and always returns `422 AI_UNSUPPORTED_CAPABILITY`; no voice provider call or persistence side effect is allowed.
