// Package profile loads Model Profile YAML files from
// `AI_MODEL_PROFILE_PATH` and exposes a name-keyed resolver to the
// AIClient. Profiles are reloaded periodically so a YAML edit takes effect
// in the running process within ≤30 s (spec §6 C-4).
//
// The loader uses RW mutex + atomic snapshot replacement so an in-flight
// AIClient call keeps the *ModelProfile pointer it captured at dispatch
// time; subsequent calls observe the new snapshot. Plan 001 ships with a
// polling reloader (plan permits fsnotify fallback under §2.3); fsnotify
// can be wired in by future plans without changing the public API.
package profile
