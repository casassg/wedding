# Migrations

The SQLite DB is baked into the Docker image at build time (`RUN sqlite3` in Dockerfile).
No persistent volume, no LiteFS. The deploy uses `strategy = "bluegreen"` in `fly.toml`,
which creates fresh machines from the new image and destroys old ones, guaranteeing
the baked-in schema is always current.

`CREATE TABLE IF NOT EXISTS` is safe here because every deploy starts from a fresh image.
