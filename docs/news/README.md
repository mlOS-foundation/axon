# Axon News & Announcements

This directory contains news entries that are automatically synced to the [mlosfoundation.org](https://mlosfoundation.org) website.

## Adding a News Entry

1. Create a JSON file in `docs/news/entries/` with the naming convention:
   ```
   YYYY-MM-DD-descriptive-slug.json
   ```

2. Follow the schema in `news-schema.json`. Required fields:
   - `id`: Unique identifier (e.g., `axon-v3.1.0-release`)
   - `title`: News title
   - `date`: Publication date (YYYY-MM-DD)
   - `type`: One of `release`, `announcement`, `security`, `milestone`, `event`, `blog`
   - `summary`: Short summary (1-3 sentences)

3. Commit and push to `main`. The sync workflow will automatically trigger.

## Example

```json
{
  "id": "axon-v3.1.0-release",
  "title": "Axon v3.1.0 - New Features",
  "date": "2026-01-15",
  "type": "release",
  "featured": true,
  "badge": "NEW RELEASE",
  "summary": "New Axon release with improved model management.",
  "version": "v3.1.0",
  "links": [
    {
      "text": "View Release",
      "url": "https://github.com/mlOS-foundation/axon/releases/tag/v3.1.0",
      "primary": true
    }
  ]
}
```

## Workflow

```
Push to main → Validate → Build Feed → Trigger Website → Auto-deploy
```
