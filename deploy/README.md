# Docker Compose Deploy

Create production env from the template:

```bash
cp .env.example .env
```

Required files on the Ubuntu server:

```txt
.env
docker-compose.yml
be/secrets/firebase-service-account.json
ota-storage/
```

Create the shared Nginx Docker network once if it does not exist:

```bash
docker network create nginx_network
```

Start the API:

```bash
make compose-up
```

Run database migrations:

```bash
make compose-migrate
```

Publish a production OTA bundle on the server:

```bash
make compose-ota
```

Override OTA version when needed:

```bash
make compose-ota OTA_VERSION=20260803T231500Z
```

The API serves OTA files from the mounted `./ota-storage` directory, so the API
container does not need to restart after publishing.

Build and upload Android release APK to DeployGate:

```bash
make android-deploygate
```

Create a release keystore before the first release build:

```bash
make android-keystore
```

The command creates `fe/android/app/release.keystore` by default and asks for
the keystore/key passwords in the terminal. Keep this file and the passwords
backed up outside git; losing them prevents signing future updates with the same
key.

Required `.env` values:

```txt
VITE_API_BASE_URL
ANDROID_KEYSTORE_FILE
ANDROID_KEYSTORE_PASSWORD
ANDROID_KEY_ALIAS
ANDROID_KEY_PASSWORD
DEPLOYGATE_API_TOKEN
DEPLOYGATE_OWNER_NAME
MINIO_ENDPOINT
MINIO_ACCESS_KEY
MINIO_SECRET_KEY
MINIO_BUCKET
MINIO_PUBLIC_BASE_URL
```

Avatar uploads use the existing S3-compatible MinIO server. `MINIO_ENDPOINT`
must be reachable by the user's browser because it is embedded in the presigned
PUT URL. Configure the bucket for public `GET` and allow `PUT` CORS from the web
app origin. Use `MINIO_USE_SSL=true` in production.

`VITE_API_BASE_URL` is baked into the Android web bundle during release build,
so it must point to the reachable production/staging API URL. If it is omitted,
`make android-release` fails instead of producing an APK that falls back to a
local development backend.

`ANDROID_KEYSTORE_FILE` is resolved from `fe/android/app`, because Gradle runs
inside that directory. For example:

```txt
ANDROID_KEYSTORE_FILE=release.keystore
```

Validate compose with the sample env without creating `.env`:

```bash
make compose-config COMPOSE_ENV_FILE=.env.example
```
