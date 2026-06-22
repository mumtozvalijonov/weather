# weather

Go weather API backed by Redis cache and Open-Meteo.

Idea source: [Weather API Wrapper Service](https://roadmap.sh/projects/weather-api-wrapper-service).

## Run

Start the web app, API, and Redis with Docker Compose:

```sh
make up
```

The web app listens on `http://localhost:3000`.
The API listens on `http://localhost:8000`.

Test the endpoint:

```sh
curl "http://localhost:8000/weather?lat=40.7128&lon=-74.0060"
```

Stop the stack:

```sh
make down
```

Follow logs:

```sh
make logs
```

## Rate Limit Check

The `/weather` route is rate limited per client IP. Send several requests quickly:

```sh
for i in 1 2 3 4 5 6 7; do
  curl -i "http://localhost:8000/weather?lat=40.7128&lon=-74.0060"
done
```

Expected result: initial requests return `200 OK`; once the burst is exhausted, responses return `429 Too Many Requests`.

## Tests

Run backend tests inside Docker:

```sh
make test-backend
```

Run frontend lint inside Docker:

```sh
make test-web
```
