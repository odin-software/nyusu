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
   export DB_ENGINE=sqlite3
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
