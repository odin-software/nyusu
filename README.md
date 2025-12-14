# Nyusu

RSS feed reader built with Go.

## Local Setup

1. **Install dependencies**

   ```bash
   go mod download
   ```

2. **Set environment variables**

   ```bash
   export ENVIRONMENT=development
   export DB_URL=nyusu.db
   export PORT=8888
   ```

3. **Run database migrations**

   ```bash
   goose -dir sql/schema sqlite3 nyusu.db up
   ```

4. **Start the server**

   ```bash
   go run .
   ```

5. **Open in browser**
   ```
   http://localhost:8888
   ```

## Docker Setup

### Building the Docker Image

```bash
docker build -t nyusu .
```

### Running with Docker

```bash
# Create a volume for persistent database storage
docker volume create nyusu-data

# Run the container
docker run -d \
  --name nyusu \
  -p 8888:8888 \
  -v nyusu-data:/data \
  -e ENVIRONMENT=production \
  -e PORT=8888 \
  -e SCRAPPER_TICK=60 \
  ghcr.io/odin-software/nyusu:latest
```

**Note**: Database migrations run automatically on container startup, so no manual setup is required!

### Environment Variables

| Variable         | Description                                      | Default            |
| ---------------- | ------------------------------------------------ | ------------------ |
| `DB_URL`         | Path to SQLite database file                     | `/data/nyusu.db`   |
| `PORT`           | Port number for the server                       | `8888`             |
| `ENVIRONMENT`    | Environment mode (`development` or `production`) | `development`      |
| `SCRAPPER_TICK`  | Interval in seconds for RSS feed scraping        | `60`               |
| `PRODUCTION_URL` | Production URL for CORS (production only)        | `https://nyusu.do` |

### Automatic Deployment

This repository uses GitHub Actions to automatically build and push Docker images to GitHub Container Registry (ghcr.io) on every push to the `main` branch.

**How it works:**

1. Push code to `main` branch
2. GitHub Actions builds the Docker image
3. Image is pushed to `ghcr.io/odin-software/nyusu:latest`
4. Watchtower on your self-hosted server automatically pulls and updates the container

**Using Watchtower for auto-updates:**

```bash
docker run -d \
  --name watchtower \
  -v /var/run/docker.sock:/var/run/docker.sock \
  containrrr/watchtower \
  --interval 300 \
  nyusu
```

This will check for updates every 5 minutes and automatically restart the container with the new image.
