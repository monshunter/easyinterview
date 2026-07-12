# E2E.P0.044 continuous practice conversation

Given an authenticated running session, the page renders ordered assistant/user messages in one full-width chat window. Sending text calls `sendPracticeMessage({clientMessageId,text})`, persists the user message and assistant reply, and refreshes the same session.
