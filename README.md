# Nyusu

RSS feed reader built with Go, PostgreSQL, and Authentik OIDC.

## Local Setup

1. **Install dependencies**

   ```bash
   go mod download
   ```

2. **Set up PostgreSQL**

   Create a database and user for Nyusu:

   ```sql
   CREATE USER nyusu WITH PASSWORD 'password';
   CREATE DATABASE nyusu OWNER nyusu;
   ```

3. **Set environment variables**

   Copy `.env.example` to `.env` and fill in the values:

   ```bash
   cp .env.example .env
   ```

4. **Start the server**

   ```bash
   go run .
   ```

   Migrations run automatically on startup.

5. **Open in browser**
   ```
   http://localhost:8888
   ```

## Docker Setup

### Running with Docker

```bash
docker run -d \
  --name nyusu \
  -p 8888:8888 \
  -e DB_URL=postgres://nyusu:password@db-host:5432/nyusu \
  -e OIDC_ISSUER_URL=https://auth.odin.do/application/o/nyusu \
  -e OIDC_CLIENT_ID=your-client-id \
  -e OIDC_CLIENT_SECRET=your-client-secret \
  -e OIDC_REDIRECT_URL=https://nyusu.odin.do/auth/callback \
  -e ENVIRONMENT=production \
  git.odin.do/odin-software/nyusu:latest
```

Migrations run automatically on container startup.

### Environment Variables

| Variable            | Description                                      | Default                |
| ------------------- | ------------------------------------------------ | ---------------------- |
| `DB_URL`            | PostgreSQL connection string                     | —                      |
| `PORT`              | Port number for the server                       | `8888`                 |
| `ENVIRONMENT`       | Environment mode (`development` or `production`) | `development`          |
| `SCRAPPER_TICK`     | Interval in seconds for RSS feed fetching        | `300` (production)     |
| `PRODUCTION_URL`    | Production URL for CORS                          | `https://nyusu.odin.do`|
| `OIDC_ISSUER_URL`   | Authentik OIDC issuer URL                        | —                      |
| `OIDC_CLIENT_ID`    | OIDC client ID                                   | —                      |
| `OIDC_CLIENT_SECRET`| OIDC client secret                               | —                      |
| `OIDC_REDIRECT_URL` | OIDC callback URL                                | —                      |

### Deployment

Gitea Actions automatically builds and pushes Docker images to `git.odin.do/odin-software/nyusu` on every push to `main`.
