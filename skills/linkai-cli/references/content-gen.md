# Image / Video / Audio generation

All three are in default scopes.

## Image — scope `image:gen`

```
linkai image gen "<prompt>" [--model <m>] [--size <s>] [--aspect-ratio <r>] [--quality <q>] [--image <url>]... [--json] [--dry-run]
```

- **Do not hardcode model names.** Omit `--model` to use the account default, or discover available image models with `linkai model list --type IMAGE` and pass a returned code to `--model`.
- `--size` / `--aspect-ratio` / `--quality` are model-specific (e.g. `2K`, `16:9`, `hd`); leave unset to use the model default. Unsupported fields are ignored by the server, not rejected.
- Image-to-image: pass one or more reference image URLs via `--image` (repeatable). Support is model-dependent.

JSON output: `{ "url": "...", ... }`

## Video — scope `video:gen`

```
linkai video gen "<prompt>" [--model <m>] [--duration <sec>] [--aspect-ratio <r>] [--size <s>] [--mode std|pro] [--image <url>]... [--image-mode <m>] [--json] [--dry-run]
```

- **Do not hardcode model names.** Omit `--model` to use the account default, or discover available video models with `linkai model list --type VIDEO` and pass a returned code to `--model`.
- All sizing options (`--duration`, `--aspect-ratio`, `--size`, `--mode`) are model-specific; leave unset to use the model default. `--mode` only applies to kling models.
- Image-to-video: pass reference image URL(s) via `--image` (repeatable); `--image-mode` chooses `reference` or `first_last_frame` (model-dependent).
- The CLI polls until the task completes (typically 30s–3min). **Do not add your own polling.** Just wait for the command to return.

JSON output: `{ "video_url": "...", "task_id": "...", ... }`

## Audio TTS — scope `audio:gen`

```
linkai audio speech "<text>" [--model tts-1|tts-1-hd] [--voice <id>] [--output <path.mp3>] [--json] [--dry-run]
```

- Without `--output`: returns `{ "url": "..." }` (CDN URL, time-limited).
- With `--output <path>`: downloads to disk; the local path is reported.
- `--model` defaults to `tts-1`; use `tts-1-hd` for higher quality.
