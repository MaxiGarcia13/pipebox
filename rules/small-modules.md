# Small modules

- Keep components and functions small, with a single clear responsibility.
- Prefer splitting into focused files over growing one large file.
- Extract helpers, composables, and subcomponents when a unit does more than one job.
- Keep Electron **main**, **preload**, and **renderer** concerns in separate modules — do not mix process boundaries in one file.
