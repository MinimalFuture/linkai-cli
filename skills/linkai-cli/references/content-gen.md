# Content Generation (Image / Video / Audio)

Generate multimedia content using AI models on the LinkAI platform.

## Image generation

**Required scope**: `image:write`

```bash
linkai image gen "<prompt>"
```

### Flags

| Flag | Type | Description |
|---|---|---|
| `--model` | string | Image model (e.g., `dall-e-3`, `doubao-seedream-4.5`) |
| `--size` | string | Image size (e.g., `1024x1024`) |
| `--aspect-ratio` | string | Aspect ratio (e.g., `1:1`, `16:9`) — use instead of `--size` when the model supports it |
| `--json` | bool | JSON output |
| `--dry-run` | bool | Print request without executing |

### Examples

```bash
linkai image gen "a cat sitting on a cloud"
linkai image gen "logo design for a coffee shop" --model dall-e-3 --size 1024x1024
linkai image gen "landscape painting" --aspect-ratio 16:9
```

The response contains the generated image URL.

---

## Video generation

**Required scope**: `video:write`

```bash
linkai video gen "<prompt>"
```

The CLI automatically polls for completion and prints the video URL when ready. Generation typically takes 30 seconds to 3 minutes.

### Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--model` | string | | Video model (e.g., `jimeng_t2v_v30`) |
| `--duration` | int | 5 | Video duration in seconds |
| `--aspect-ratio` | string | `16:9` | Aspect ratio (`16:9`, `9:16`, `1:1`) |
| `--mode` | string | `std` | Generation mode: `std` (standard) or `pro` (higher quality) |
| `--json` | bool | | JSON output |
| `--dry-run` | bool | | Print request without executing |

### Examples

```bash
linkai video gen "a dolphin jumping out of the ocean"
linkai video gen "product showcase" --duration 10 --mode pro --aspect-ratio 9:16
```

---

## Audio / Text-to-Speech (TTS)

**Required scope**: `audio:write`

```bash
linkai audio speech "<text>"
```

### Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--model` | string | `tts-1` | TTS model (`tts-1` or `tts-1-hd`) |
| `--voice` | string | | Voice type ID |
| `--output` | string | | Save audio to local file (e.g., `speech.mp3`) |
| `--json` | bool | | JSON output |
| `--dry-run` | bool | | Print request without executing |

### Examples

```bash
linkai audio speech "Hello, welcome to LinkAI"
linkai audio speech "今天天气真好" --model tts-1-hd --output greeting.mp3
linkai audio speech "Product introduction" --voice <voice_id> --output intro.mp3
```

Without `--output`, the response contains the audio URL. With `--output`, the audio file is downloaded and saved locally.
