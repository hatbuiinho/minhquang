# OTA Storage

Local self-hosted Capacitor bundles are served from this directory at `/ota/`.

Generate a dev Android bundle from the frontend project:

```bash
cd ../fe
yarn ota:publish
```

That writes:

- `android/dev/<version>.zip`
- `android/dev/latest.json`
