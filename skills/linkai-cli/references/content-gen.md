# Image / Video / Audio generation

All three are in default scopes.

## Image — scope `image:gen`

```
linkai image gen "<prompt>" [--model <m>] [--size <WxH> | --aspect-ratio <r>] [--json] [--dry-run]
```

- Prefer `--aspect-ratio` (e.g. `16:9`, `1:1`) when the model supports it; otherwise use `--size 1024x1024`.
- Common models: `dall-e-3`, `doubao-seedream-4.5`. Omit `--model` to use the default.

JSON output: `{ "url": "...", ... }`

## Video — scope `video:gen`

```
linkai video gen "<prompt>" [--model <m>] [--duration <sec>] [--aspect-ratio <r>] [--mode std|pro] [--json] [--dry-run]
```

- Defaults: `duration=5`, `aspect-ratio=16:9`, `mode=std`.
- The CLI polls until the task completes (typically 30s–3min). **Do not add your own polling.** Just wait for the command to return.
- `mode=pro` is slower but higher quality.

JSON output: `{ "url": "...", "task_id": "...", ... }`

## Audio TTS — scope `audio:gen`

```
linkai audio speech "<text>" [--model tts-1|tts-1-hd] [--voice <id>] [--output <path.mp3>] [--json] [--dry-run]
```

- Without `--output`: returns `{ "url": "..." }` (CDN URL, time-limited).
- With `--output <path>`: downloads to disk; the local path is reported.
- `--model` defaults to `tts-1`; use `tts-1-hd` for higher quality.
