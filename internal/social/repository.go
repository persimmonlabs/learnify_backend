package social

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// Repository handles social data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new social repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// FollowUser creates follow relationship
func (r *Repository) FollowUser(followerID, followingID string) error {
	query := `
		INSERT INTO user_relationships (follower_id, following_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (follower_id, following_id) DO NOTHING
	`
	_, err := r.db.Exec(query, followerID, followingID)
	if err != nil {
		return fmt.Errorf("failed to create follow relationship: %w", err)
	}
	return nil
}

// UnfollowUser removes follow relationship
func (r *Repository) UnfollowUser(followerID, followingID string) error {
	query := `
		DELETE FROM user_relationships
		WHERE follower_id = $1 AND following_id = $2
	`
	result, err := r.db.Exec(query, followerID, followingID)
	if err != nil {
		return fmt.Errorf("failed to remove follow relationship: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("follow relationship not found")
	}

	return nil
}

// GetFollowers retrieves user's followers
func (r *Repository) GetFollowers(userID string) ([]string, error) {
	query := `
		SELECT follower_id
		FROM user_relationships
		WHERE following_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query followers: %w", err)
	}
	defer rows.Close()

	var followers []string
	for rows.Next() {
		var followerID string
		if err := rows.Scan(&followerID); err != nil {
			return nil, fmt.Errorf("failed to scan follower: %w", err)
		}
		followers = append(followers, followerID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating followers: %w", err)
	}

	return followers, nil
}

// GetFollowing retrieves users that user follows
func (r *Repository) GetFollowing(userID string) ([]string, error) {
	query := `
		SELECT following_id
		FROM user_relationships
		WHERE follower_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query following: %w", err)
	}
	defer rows.Close()

	var following []string
	for rows.Next() {
		var followingID string
		if err := rows.Scan(&followingID); err != nil {
			return nil, fmt.Errorf("failed to scan following: %w", err)
		}
		following = append(following, followingID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating following: %w", err)
	}

	return following, nil
}

// CreateActivity creates activity feed item
func (r *Repository) CreateActivity(activity *ActivityFeed) error {
	metadataJSON, err := json.Marshal(activity.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO activity_feed (
			id, user_id, activity_type, reference_type, reference_id,
			metadata, visibility, created_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7
		)
		RETURNING id, created_at
	`

	err = r.db.QueryRow(
		query,
		activity.UserID,
		activity.ActivityType,
		activity.ReferenceType,
		activity.ReferenceID,
		metadataJSON,
		activity.Visibility,
		time.Now(),
	).Scan(&activity.ID, &activity.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create activity: %w", err)
	}

	return nil
}

// GetActivityFeed retrieves activity feed for user
func (r *Repository) GetActivityFeed(userID string, limit int) ([]ActivityFeed, error) {
	query := `
		SELECT
			af.id,
			af.user_id,
			af.activity_type,
			af.reference_type,
			af.reference_id,
			af.metadata,
			af.visibility,
			af.created_at
		FROM activity_feed af
		INNER JOIN user_relationships ur ON af.user_id = ur.following_id
		WHERE ur.follower_id = $1
			AND (af.visibility = 'public' OR af.visibility = 'friends')
		ORDER BY af.created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query activity feed: %w", err)
	}
	defer rows.Close()

	var activities []ActivityFeed
	for rows.Next() {
		var activity ActivityFeed
		var metadataJSON []byte

		err := rows.Scan(
			&activity.ID,
			&activity.UserID,
			&activity.ActivityType,
			&activity.ReferenceType,
			&activity.ReferenceID,
			&metadataJSON,
			&activity.Visibility,
			&activity.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &activity.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating activities: %w", err)
	}

	return activities, nil
}

// GetRecommendations retrieves course recommendations
func (r *Repository) GetRecommendations(userID string, recType string) ([]Recommendation, error) {
	query := `
		SELECT
			id,
			user_id,
			course_id,
			recommendation_type,
			match_score,
			reason,
			metadata,
			created_at,
			expires_at
		FROM recommendations
		WHERE user_id = $1
			AND (expires_at IS NULL OR expires_at > NOW())
	`

	args := []interface{}{userID}
	if recType != "" && recType != "all" {
		query += " AND recommendation_type = $2"
		args = append(args, recType)
	}

	query += " ORDER BY match_score DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query recommendations: %w", err)
	}
	defer rows.Close()

	var recommendations []Recommendation
	for rows.Next() {
		var rec Recommendation
		var metadataJSON []byte

		err := rows.Scan(
			&rec.ID,
			&rec.UserID,
			&rec.CourseID,
			&rec.RecommendationType,
			&rec.MatchScore,
			&rec.Reason,
			&metadataJSON,
			&rec.CreatedAt,
			&rec.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recommendation: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &rec.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		recommendations = append(recommendations, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recommendations: %w", err)
	}

	return recommendations, nil
}

// CreateRecommendation creates recommendation (or updates if exists)
func (r *Repository) CreateRecommendation(rec *Recommendation) error {
	metadataJSON, err := json.Marshal(rec.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO recommendations (
			user_id, course_id, recommendation_type, match_score,
			reason, metadata, created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, course_id, recommendation_type)
		DO UPDATE SET
			match_score = EXCLUDED.match_score,
			reason = EXCLUDED.reason,
			metadata = EXCLUDED.metadata,
			created_at = EXCLUDED.created_at,
			expires_at = EXCLUDED.expires_at
		RETURNING id
	`

	err = r.db.QueryRow(
		query,
		rec.UserID,
		rec.CourseID,
		rec.RecommendationType,
		rec.MatchScore,
		rec.Reason,
		metadataJSON,
		time.Now(),
		rec.ExpiresAt,
	).Scan(&rec.ID)

	if err != nil {
		return fmt.Errorf("failed to create recommendation: %w", err)
	}

	return nil
}

// GetTrendingCourses retrieves trending courses
func (r *Repository) GetTrendingCourses(limit int) ([]TrendingCourse, error) {
	query := `
		SELECT
			id,
			course_id,
			velocity,
			signups_24h,
			signups_previous_24h,
			rank,
			meta_category,
			calculated_at
		FROM trending_courses
		ORDER BY rank ASC
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query trending courses: %w", err)
	}
	defer rows.Close()

	var courses []TrendingCourse
	for rows.Next() {
		var course TrendingCourse
		err := rows.Scan(
			&course.ID,
			&course.CourseID,
			&course.Velocity,
			&course.Signups24h,
			&course.SignupsPrevious24h,
			&course.Rank,
			&course.MetaCategory,
			&course.CalculatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trending course: %w", err)
		}
		courses = append(courses, course)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trending courses: %w", err)
	}

	return courses, nil
}

// UpdateTrendingCourses updates trending cache (batch operation)
func (r *Repository) UpdateTrendingCourses(courses []TrendingCourse) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete old trending data
	_, err = tx.Exec("DELETE FROM trending_courses")
	if err != nil {
		return fmt.Errorf("failed to delete old trending data: %w", err)
	}

	// Batch insert new trending data
	if len(courses) > 0 {
		stmt, err := tx.Prepare(pq.CopyIn(
			"trending_courses",
			"course_id",
			"velocity",
			"signups_24h",
			"signups_previous_24h",
			"rank",
			"meta_category",
			"calculated_at",
		))
		if err != nil {
			return fmt.Errorf("failed to prepare copy statement: %w", err)
		}

		for _, course := range courses {
			_, err = stmt.Exec(
				course.CourseID,
				course.Velocity,
				course.Signups24h,
				course.SignupsPrevious24h,
				course.Rank,
				course.MetaCategory,
				time.Now(),
			)
			if err != nil {
				return fmt.Errorf("failed to add course to batch: %w", err)
			}
		}

		_, err = stmt.Exec()
		if err != nil {
			return fmt.Errorf("failed to execute batch insert: %w", err)
		}

		err = stmt.Close()
		if err != nil {
			return fmt.Errorf("failed to close statement: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetUserAchievements retrieves earned achievements
func (r *Repository) GetUserAchievements(userID string) ([]Achievement, error) {
	query := `
		SELECT
			a.id,
			a.name,
			a.description,
			a.badge_icon,
			a.criteria,
			a.rarity,
			a.created_at,
			ua.unlocked_at
		FROM achievements a
		INNER JOIN user_achievements ua ON a.id = ua.achievement_id
		WHERE ua.user_id = $1
		ORDER BY ua.unlocked_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query achievements: %w", err)
	}
	defer rows.Close()

	var achievements []Achievement
	for rows.Next() {
		var achievement Achievement
		var criteriaJSON []byte
		var unlockedAt time.Time

		err := rows.Scan(
			&achievement.ID,
			&achievement.Name,
			&achievement.Description,
			&achievement.BadgeIcon,
			&criteriaJSON,
			&achievement.Rarity,
			&achievement.CreatedAt,
			&unlockedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan achievement: %w", err)
		}

		if len(criteriaJSON) > 0 {
			if err := json.Unmarshal(criteriaJSON, &achievement.Criteria); err != nil {
				return nil, fmt.Errorf("failed to unmarshal criteria: %w", err)
			}
		}

		achievements = append(achievements, achievement)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating achievements: %w", err)
	}

	return achievements, nil
}

// UnlockAchievement awards achievement to user
func (r *Repository) UnlockAchievement(userID, achievementID string) error {
	query := `
		INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, achievement_id) DO NOTHING
	`

	result, err := r.db.Exec(query, userID, achievementID)
	if err != nil {
		return fmt.Errorf("failed to unlock achievement: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("achievement already unlocked")
	}

	return nil
}

// GetCollaborativeFilteringCandidates finds users with similar course completions
func (r *Repository) GetCollaborativeFilteringCandidates(userID string, minOverlap float64) ([]string, error) {
	query := `
		WITH user_courses AS (
			SELECT
				user_id,
				array_agg(course_id) as courses,
				count(*) as course_count
			FROM user_progress
			WHERE completed_at IS NOT NULL
			GROUP BY user_id
		),
		current_user AS (
			SELECT courses, course_count
			FROM user_courses
			WHERE user_id = $1
		)
		SELECT uc.user_id
		FROM user_courses uc, current_user cu
		WHERE uc.user_id != $1
			AND uc.courses && cu.courses
		GROUP BY uc.user_id, cu.course_count
		HAVING count(*) >= $2 * cu.course_count
		LIMIT 50
	`

	rows, err := r.db.Query(query, userID, minOverlap)
	if err != nil {
		return nil, fmt.Errorf("failed to query similar users: %w", err)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, uid)
	}

	return userIDs, nil
}

// GetCoursesCompletedByUsers retrieves courses completed by list of users
func (r *Repository) GetCoursesCompletedByUsers(userIDs []string, excludeUserID string) ([]string, error) {
	query := `
		SELECT DISTINCT course_id
		FROM user_progress
		WHERE user_id = ANY($1)
			AND completed_at IS NOT NULL
			AND course_id NOT IN (
				SELECT course_id
				FROM user_progress
				WHERE user_id = $2
			)
	`

	rows, err := r.db.Query(query, pq.Array(userIDs), excludeUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to query courses: %w", err)
	}
	defer rows.Close()

	var courseIDs []string
	for rows.Next() {
		var cid string
		if err := rows.Scan(&cid); err != nil {
			return nil, fmt.Errorf("failed to scan course ID: %w", err)
		}
		courseIDs = append(courseIDs, cid)
	}

	return courseIDs, nil
}

// CalculateTrendingVelocity calculates velocity for all courses
func (r *Repository) CalculateTrendingVelocity() ([]TrendingCourse, error) {
	query := `
		SELECT
			gc.id as course_id,
			gc.meta_category,
			COUNT(*) FILTER (
				WHERE up.started_at > NOW() - INTERVAL '24 hours'
			) as signups_24h,
			COUNT(*) FILTER (
				WHERE up.started_at BETWEEN NOW() - INTERVAL '48 hours'
					AND NOW() - INTERVAL '24 hours'
			) as signups_prev_24h,
			CASE
				WHEN COUNT(*) FILTER (
					WHERE up.started_at BETWEEN NOW() - INTERVAL '48 hours'
						AND NOW() - INTERVAL '24 hours'
				) > 0
				THEN
					COUNT(*) FILTER (WHERE up.started_at > NOW() - INTERVAL '24 hours')::decimal /
					COUNT(*) FILTER (WHERE up.started_at BETWEEN NOW() - INTERVAL '48 hours'
						AND NOW() - INTERVAL '24 hours')
				ELSE
					CASE
						WHEN COUNT(*) FILTER (WHERE up.started_at > NOW() - INTERVAL '24 hours') > 0
						THEN 10.0
						ELSE 0.0
					END
			END as velocity
		FROM generated_courses gc
		LEFT JOIN user_progress up ON gc.id = up.course_id
		GROUP BY gc.id, gc.meta_category
		HAVING COUNT(*) FILTER (WHERE up.started_at > NOW() - INTERVAL '24 hours') > 0
		ORDER BY velocity DESC
		LIMIT 100
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate velocity: %w", err)
	}
	defer rows.Close()

	var courses []TrendingCourse
	rank := 1
	for rows.Next() {
		var course TrendingCourse
		err := rows.Scan(
			&course.CourseID,
			&course.MetaCategory,
			&course.Signups24h,
			&course.SignupsPrevious24h,
			&course.Velocity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trending course: %w", err)
		}
		course.Rank = rank
		courses = append(courses, course)
		rank++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trending courses: %w", err)
	}

	return courses, nil
}
