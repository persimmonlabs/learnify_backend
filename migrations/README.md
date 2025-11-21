# Learnify Database Migrations

Organized PostgreSQL migrations for the Universal Blueprint learning platform.

## Migration Files

| File | Description | Tables Created |
|------|-------------|----------------|
| `000_migration_tracker.sql` | Migration tracking system | `schema_migrations` |
| `001_create_identity_tables.sql` | User identity & variables | `users`, `user_archetypes`, `user_variables` |
| `002_create_curriculum_tables.sql` | Blueprint & courses | `blueprint_modules`, `generated_courses`, `generated_modules`, `exercises`, `course_tags` |
| `003_create_progress_tables.sql` | Learning progress | `user_progress`, `module_completions`, `architecture_reviews` |
| `004_create_social_tables.sql` | Social network | `user_relationships`, `activity_feed`, `achievements`, `user_achievements` |
| `005_create_discovery_tables.sql` | Recommendations | `recommendations`, `trending_courses`, `user_course_interactions` |

## Running Migrations

### Method 1: Sequential Execution (Recommended)

```bash
# Run in order with PostgreSQL
psql -U postgres -d learnify -f backend/migrations/000_migration_tracker.sql
psql -U postgres -d learnify -f backend/migrations/001_create_identity_tables.sql
psql -U postgres -d learnify -f backend/migrations/002_create_curriculum_tables.sql
psql -U postgres -d learnify -f backend/migrations/003_create_progress_tables.sql
psql -U postgres -d learnify -f backend/migrations/004_create_social_tables.sql
psql -U postgres -d learnify -f backend/migrations/005_create_discovery_tables.sql
```

### Method 2: Batch Execution

```bash
# Run all migrations at once
cat backend/migrations/*.sql | psql -U postgres -d learnify
```

### Method 3: Node.js Migration Runner (Future)

```javascript
// backend/scripts/migrate.js
const { Pool } = require('pg');
const fs = require('fs');
const path = require('path');

async function runMigrations() {
  const pool = new Pool({ connectionString: process.env.DATABASE_URL });
  const files = fs.readdirSync('./backend/migrations').sort();

  for (const file of files) {
    if (!file.endsWith('.sql')) continue;

    const version = file.split('_')[0];
    const exists = await pool.query(
      'SELECT 1 FROM schema_migrations WHERE version = $1',
      [version]
    );

    if (exists.rows.length > 0) {
      console.log(`✓ Migration ${version} already applied`);
      continue;
    }

    const sql = fs.readFileSync(path.join('./backend/migrations', file), 'utf-8');
    await pool.query(sql);
    console.log(`✓ Applied migration ${version}`);
  }

  await pool.end();
}

runMigrations().catch(console.error);
```

## Database Schema Layers

### Layer 1: Identity
- User accounts and authentication
- Archetype selection (meta-category + domain)
- Runtime variable injection (`{ENTITY}`, `{STATE}`, etc.)

### Layer 2: Curriculum
- Universal Blueprint templates (7 modules)
- Generated course instances (variables injected)
- Exercises with test cases

### Layer 3: Progress
- Course progress tracking
- Exercise submissions
- AI Senior Reviews (4-category scoring)

### Layer 4: Social
- Follow graph
- Activity feed (Strava for Brains)
- Achievement system

### Layer 5: Discovery
- Collaborative filtering
- Skill adjacency recommendations
- Trending velocity calculations
- Social signal recommendations

## Key Features

### Template vs Instance
- **Blueprint Modules**: Universal templates with `{VARIABLE}` placeholders
- **Generated Courses**: User-specific instances with injected variables
- **Variable Injection**: Stored in `user_variables` table

### Social Architecture
- **No Fluff**: Data-driven activity only (no likes, comments, status updates)
- **Passive Social**: Automatic progress broadcasts
- **Netflix Rows**: Algorithmic recommendations

### Performance Optimizations
- Denormalized activity feed for fast reads
- Cached trending courses (hourly refresh)
- Pre-computed recommendations
- Comprehensive indexing strategy

## Schema Validation

Check applied migrations:
```sql
SELECT * FROM schema_migrations ORDER BY version;
```

Count tables:
```sql
SELECT COUNT(*) FROM information_schema.tables
WHERE table_schema = 'public';
-- Expected: 18 tables + 1 migration tracker = 19 total
```

## Rollback Strategy

Migrations do not include automatic rollback. To reverse:

```sql
-- Example: Rollback migration 005
DROP TABLE IF EXISTS user_course_interactions CASCADE;
DROP TABLE IF EXISTS trending_courses CASCADE;
DROP TABLE IF EXISTS recommendations CASCADE;
DELETE FROM schema_migrations WHERE version = '005';
```

## Next Steps

1. Set up PostgreSQL database: `createdb learnify`
2. Run migrations in order
3. Verify with: `psql -d learnify -c "\dt"`
4. Build backend API for:
   - LLM domain validation
   - Variable injection engine
   - Curriculum generator
   - Recommendation algorithms

---

**Philosophy**: "Strava for Brains, not Facebook for Learning"
