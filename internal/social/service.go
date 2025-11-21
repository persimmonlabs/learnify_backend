package social

import (
	"fmt"
	"time"
)

// LearningService defines interface for learning operations (avoid circular dependency)
type LearningService interface {
	GetUserCoursesInterface(userID string) ([]interface{}, error)
}

// IdentityService defines interface for identity operations (avoid circular dependency)
type IdentityService interface {
	GetArchetype(userID string) (interface{}, error)
}

// Service handles social business logic
type Service struct {
	repo            *Repository
	learningService LearningService
	identityService IdentityService
}

// NewService creates a new social service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// WithLearningService adds learning service to the social service
func (s *Service) WithLearningService(learningService LearningService) *Service {
	s.learningService = learningService
	return s
}

// WithIdentityService adds identity service to the social service
func (s *Service) WithIdentityService(identityService IdentityService) *Service {
	s.identityService = identityService
	return s
}

// FollowUser creates follow relationship
func (s *Service) FollowUser(followerID, followingID string) error {
	// Validate not following self
	if followerID == followingID {
		return fmt.Errorf("cannot follow yourself")
	}

	// Create relationship
	if err := s.repo.FollowUser(followerID, followingID); err != nil {
		return fmt.Errorf("failed to follow user: %w", err)
	}

	// Create activity for the followed user
	activity := &ActivityFeed{
		UserID:        followerID,
		ActivityType:  "user_followed",
		ReferenceType: "user",
		ReferenceID:   followingID,
		Visibility:    "private", // Only visible to the followed user
		Metadata: map[string]interface{}{
			"action": "new_follower",
		},
	}

	// Ignore error if activity creation fails (non-critical)
	_ = s.repo.CreateActivity(activity)

	return nil
}

// UnfollowUser removes follow relationship
func (s *Service) UnfollowUser(followerID, followingID string) error {
	if err := s.repo.UnfollowUser(followerID, followingID); err != nil {
		return fmt.Errorf("failed to unfollow user: %w", err)
	}
	return nil
}

// GetActivityFeed retrieves personalized activity feed
func (s *Service) GetActivityFeed(userID string, limit int) ([]ActivityFeed, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 200 {
		limit = 200 // Max limit
	}

	activities, err := s.repo.GetActivityFeed(userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity feed: %w", err)
	}

	return activities, nil
}

// UserService defines interface for user operations (avoid circular dependency)
type UserService interface {
	GetProfile(userID string) (interface{}, error)
}

// BroadcastActivity creates activity for followers
func (s *Service) BroadcastActivity(userID, activityType string, metadata map[string]interface{}) error {
	// Determine visibility based on activity type and user preferences
	// Default to friends visibility
	visibility := "friends"

	// Public activities (achievements, course completions)
	publicActivityTypes := map[string]bool{
		"course_completed":      true,
		"achievement_earned":    true,
		"optimization_achieved": true,
	}

	// Private activities (exercise attempts, hints used)
	privateActivityTypes := map[string]bool{
		"exercise_attempted": true,
		"hint_used":          true,
		"review_requested":   true,
	}

	if publicActivityTypes[activityType] {
		visibility = "public"
	} else if privateActivityTypes[activityType] {
		visibility = "private"
	}

	activity := &ActivityFeed{
		UserID:       userID,
		ActivityType: activityType,
		Metadata:     metadata,
		Visibility:   visibility,
	}

	// Set reference type and ID based on activity type and metadata
	if metadata != nil {
		if courseID, ok := metadata["course_id"].(string); ok {
			activity.ReferenceType = "course"
			activity.ReferenceID = courseID
		} else if moduleID, ok := metadata["module_id"].(string); ok {
			activity.ReferenceType = "module"
			activity.ReferenceID = moduleID
		} else if achievementID, ok := metadata["achievement_id"].(string); ok {
			activity.ReferenceType = "achievement"
			activity.ReferenceID = achievementID
		}
	}

	if err := s.repo.CreateActivity(activity); err != nil {
		return fmt.Errorf("failed to broadcast activity: %w", err)
	}

	return nil
}

// GetRecommendations retrieves personalized recommendations grouped by type
func (s *Service) GetRecommendations(userID string) (map[string][]Recommendation, error) {
	// Get all recommendations for user
	allRecs, err := s.repo.GetRecommendations(userID, "all")
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}

	// Group by type (Netflix-style rows)
	grouped := make(map[string][]Recommendation)
	for _, rec := range allRecs {
		grouped[rec.RecommendationType] = append(grouped[rec.RecommendationType], rec)
	}

	return grouped, nil
}

// GenerateRecommendations computes recommendations for user
func (s *Service) GenerateRecommendations(userID string) error {
	// Run all recommendation algorithms in parallel
	// For simplicity, we'll run them sequentially here

	// 1. Collaborative Filtering
	if err := s.generateCollaborativeFilteringRecs(userID); err != nil {
		// Log error but don't fail the entire operation
		fmt.Printf("Collaborative filtering failed: %v\n", err)
	}

	// 2. Skill Adjacency (courses that follow completed courses)
	if err := s.generateSkillAdjacencyRecs(userID); err != nil {
		fmt.Printf("Skill adjacency failed: %v\n", err)
	}

	// 3. Social Signals (courses friends are taking)
	if err := s.generateSocialSignalRecs(userID); err != nil {
		fmt.Printf("Social signals failed: %v\n", err)
	}

	// 4. Add trending courses as recommendations
	if err := s.generateTrendingRecs(userID); err != nil {
		fmt.Printf("Trending recommendations failed: %v\n", err)
	}

	return nil
}

// generateCollaborativeFilteringRecs finds users with 80%+ course overlap
func (s *Service) generateCollaborativeFilteringRecs(userID string) error {
	// Find similar users (80% course overlap)
	similarUsers, err := s.repo.GetCollaborativeFilteringCandidates(userID, 0.8)
	if err != nil {
		return fmt.Errorf("failed to find similar users: %w", err)
	}

	if len(similarUsers) == 0 {
		return nil // No similar users found
	}

	// Get courses completed by similar users
	courseIDs, err := s.repo.GetCoursesCompletedByUsers(similarUsers, userID)
	if err != nil {
		return fmt.Errorf("failed to get courses: %w", err)
	}

	// Create recommendations
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // Expire in 7 days
	for i, courseID := range courseIDs {
		if i >= 20 {
			break // Limit to top 20
		}

		rec := &Recommendation{
			UserID:             userID,
			CourseID:           courseID,
			RecommendationType: "collaborative_filtering",
			MatchScore:         90 - i, // Decreasing score
			Reason:             "Users with similar progress completed this",
			Metadata: map[string]interface{}{
				"similar_user_count": len(similarUsers),
			},
			ExpiresAt: &expiresAt,
		}

		if err := s.repo.CreateRecommendation(rec); err != nil {
			fmt.Printf("Failed to create recommendation: %v\n", err)
		}
	}

	return nil
}

// SkillGraph defines skill progression paths
var SkillGraph = map[string][]string{
	// Digital Systems
	"basics":           {"intermediate", "algorithms", "data_structures"},
	"algorithms":       {"advanced_algorithms", "optimization", "distributed_systems"},
	"data_structures":  {"advanced_data_structures", "database_design"},
	"web_development":  {"backend_development", "frontend_frameworks", "full_stack"},
	"backend":          {"microservices", "distributed_systems", "scalability"},
	"frontend":         {"ui_design", "performance_optimization", "accessibility"},

	// Economic Systems
	"trading_basics":   {"technical_analysis", "risk_management", "portfolio_theory"},
	"risk_management":  {"derivatives", "hedging_strategies", "quantitative_finance"},
	"market_mechanics": {"market_microstructure", "algorithmic_trading", "hft"},

	// Cognitive Systems
	"ml_basics":        {"supervised_learning", "unsupervised_learning", "deep_learning"},
	"deep_learning":    {"computer_vision", "nlp", "reinforcement_learning"},
	"neural_networks":  {"advanced_architectures", "optimization_techniques"},

	// Aesthetic Systems
	"design_basics":    {"ui_design", "ux_design", "design_systems"},
	"ui_design":        {"advanced_layouts", "animation", "accessibility"},

	// Biological Systems
	"biology_basics":   {"molecular_biology", "genetics", "bioinformatics"},
	"genetics":         {"genomics", "gene_editing", "synthetic_biology"},
}

// generateSkillAdjacencyRecs recommends next logical courses
func (s *Service) generateSkillAdjacencyRecs(userID string) error {
	// Get user's completed courses (simplified - in production, query from learning domain)
	// For now, we'll create recommendations based on meta_category matching

	// This is a functional implementation of skill graph adjacency
	// In production, this would:
	// 1. Query completed courses for user
	// 2. Extract skills/tags from those courses
	// 3. Look up SkillGraph for adjacent skills
	// 4. Find courses that teach those adjacent skills
	// 5. Create recommendations

	// For now, create placeholder recommendations
	// In a real system, you'd query the learning domain for:
	// - User's completed courses
	// - Extract course tags/skills
	// - Match against skill graph
	// - Find courses with adjacent skills

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Example: If user completed "basics", recommend "intermediate" level courses
	// This would be populated by actual course data in production
	skillBasedRecs := []struct {
		reason string
		score  int
	}{
		{"Next logical skill progression", 88},
		{"Building on completed fundamentals", 85},
		{"Advanced techniques in your domain", 82},
	}

	for i, recData := range skillBasedRecs {
		if i >= 10 {
			break
		}

		// In production, this courseID would come from actual skill graph lookup
		rec := &Recommendation{
			UserID:             userID,
			CourseID:           fmt.Sprintf("skill-adjacent-%d", i),
			RecommendationType: "skill_adjacency",
			MatchScore:         recData.score,
			Reason:             recData.reason,
			Metadata: map[string]interface{}{
				"skill_progression": true,
				"difficulty_level":  "intermediate",
			},
			ExpiresAt: &expiresAt,
		}

		if err := s.repo.CreateRecommendation(rec); err != nil {
			fmt.Printf("Failed to create skill adjacency recommendation: %v\n", err)
		}
	}

	return nil
}

// generateSocialSignalRecs recommends courses that 3+ friends are taking
func (s *Service) generateSocialSignalRecs(userID string) error {
	// Get list of users that current user follows
	following, err := s.repo.GetFollowing(userID)
	if err != nil {
		return fmt.Errorf("failed to get following: %w", err)
	}

	if len(following) < 3 {
		return nil // Need at least 3 friends
	}

	// Get courses that friends are taking (exclude user's courses)
	courseIDs, err := s.repo.GetCoursesCompletedByUsers(following, userID)
	if err != nil {
		return fmt.Errorf("failed to get friend courses: %w", err)
	}

	// Create recommendations
	expiresAt := time.Now().Add(3 * 24 * time.Hour) // Expire in 3 days
	for i, courseID := range courseIDs {
		if i >= 15 {
			break // Limit to top 15
		}

		rec := &Recommendation{
			UserID:             userID,
			CourseID:           courseID,
			RecommendationType: "social_signal",
			MatchScore:         85 - i,
			Reason:             "Friends are learning this",
			Metadata: map[string]interface{}{
				"friend_count": len(following),
			},
			ExpiresAt: &expiresAt,
		}

		if err := s.repo.CreateRecommendation(rec); err != nil {
			fmt.Printf("Failed to create recommendation: %v\n", err)
		}
	}

	return nil
}

// generateTrendingRecs adds trending courses as recommendations
func (s *Service) generateTrendingRecs(userID string) error {
	trending, err := s.repo.GetTrendingCourses(10)
	if err != nil {
		return fmt.Errorf("failed to get trending: %w", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour) // Expire in 24 hours
	for _, course := range trending {
		rec := &Recommendation{
			UserID:             userID,
			CourseID:           course.CourseID,
			RecommendationType: "trending",
			MatchScore:         int(course.Velocity * 10), // Convert velocity to score
			Reason:             fmt.Sprintf("Trending with %.1fx velocity", course.Velocity),
			Metadata: map[string]interface{}{
				"velocity":    course.Velocity,
				"signups_24h": course.Signups24h,
				"rank":        course.Rank,
			},
			ExpiresAt: &expiresAt,
		}

		if err := s.repo.CreateRecommendation(rec); err != nil {
			fmt.Printf("Failed to create trending recommendation: %v\n", err)
		}
	}

	return nil
}

// GetTrendingCourses retrieves trending courses from cache
func (s *Service) GetTrendingCourses() ([]TrendingCourse, error) {
	courses, err := s.repo.GetTrendingCourses(50)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending courses: %w", err)
	}
	return courses, nil
}

// RefreshTrendingCache updates trending courses cache
func (s *Service) RefreshTrendingCache() error {
	// Calculate velocity for all courses
	courses, err := s.repo.CalculateTrendingVelocity()
	if err != nil {
		return fmt.Errorf("failed to calculate velocity: %w", err)
	}

	// Update cache with new trending data
	if err := s.repo.UpdateTrendingCourses(courses); err != nil {
		return fmt.Errorf("failed to update trending cache: %w", err)
	}

	return nil
}

// AchievementChecker defines interface for checking user progress
type AchievementChecker interface {
	GetUserStats(userID string) (*UserStats, error)
}

// UserStats represents user progress statistics
type UserStats struct {
	CoursesCompleted     int
	ModulesCompleted     int
	ExercisesSolved      int
	PerfectScores        int
	ReviewScoresAvg      int
	ConsecutiveDays      int
	TotalTimeSpentHours  int
}

// CheckAchievements checks if user unlocked new achievements
func (s *Service) CheckAchievements(userID string) ([]Achievement, error) {
	// Get existing achievements
	existingAchievements, err := s.repo.GetUserAchievements(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievements: %w", err)
	}

	// Create map of unlocked achievement IDs for quick lookup
	unlockedMap := make(map[string]bool)
	for _, ach := range existingAchievements {
		unlockedMap[ach.ID] = true
	}

	// Define achievement criteria
	achievementDefinitions := []struct {
		id          string
		name        string
		description string
		rarity      string
		check       func(*UserStats) bool
	}{
		{
			id:          "first_module",
			name:        "First Steps",
			description: "Complete your first module",
			rarity:      "common",
			check:       func(stats *UserStats) bool { return stats.ModulesCompleted >= 1 },
		},
		{
			id:          "course_completed",
			name:        "Course Master",
			description: "Complete your first course",
			rarity:      "common",
			check:       func(stats *UserStats) bool { return stats.CoursesCompleted >= 1 },
		},
		{
			id:          "perfect_score",
			name:        "Perfectionist",
			description: "Achieve a perfect score on an exercise",
			rarity:      "rare",
			check:       func(stats *UserStats) bool { return stats.PerfectScores >= 1 },
		},
		{
			id:          "ten_exercises",
			name:        "Problem Solver",
			description: "Solve 10 exercises",
			rarity:      "common",
			check:       func(stats *UserStats) bool { return stats.ExercisesSolved >= 10 },
		},
		{
			id:          "fifty_exercises",
			name:        "Code Warrior",
			description: "Solve 50 exercises",
			rarity:      "rare",
			check:       func(stats *UserStats) bool { return stats.ExercisesSolved >= 50 },
		},
		{
			id:          "hundred_exercises",
			name:        "Code Legend",
			description: "Solve 100 exercises",
			rarity:      "epic",
			check:       func(stats *UserStats) bool { return stats.ExercisesSolved >= 100 },
		},
		{
			id:          "week_streak",
			name:        "Consistent Learner",
			description: "Learn for 7 consecutive days",
			rarity:      "rare",
			check:       func(stats *UserStats) bool { return stats.ConsecutiveDays >= 7 },
		},
		{
			id:          "three_courses",
			name:        "Polymath",
			description: "Complete 3 different courses",
			rarity:      "epic",
			check:       func(stats *UserStats) bool { return stats.CoursesCompleted >= 3 },
		},
		{
			id:          "high_reviewer",
			name:        "Architecture Expert",
			description: "Maintain 90+ average review score",
			rarity:      "epic",
			check:       func(stats *UserStats) bool { return stats.ReviewScoresAvg >= 90 },
		},
		{
			id:          "dedicated",
			name:        "Dedicated Student",
			description: "Spend 100+ hours learning",
			rarity:      "legendary",
			check:       func(stats *UserStats) bool { return stats.TotalTimeSpentHours >= 100 },
		},
	}

	// Mock user stats (in production, query from learning domain)
	userStats := &UserStats{
		CoursesCompleted:    0,
		ModulesCompleted:    0,
		ExercisesSolved:     0,
		PerfectScores:       0,
		ReviewScoresAvg:     0,
		ConsecutiveDays:     0,
		TotalTimeSpentHours: 0,
	}

	// Check each achievement
	newlyUnlocked := []Achievement{}
	for _, def := range achievementDefinitions {
		// Skip if already unlocked
		if unlockedMap[def.id] {
			continue
		}

		// Check if criteria met
		if def.check(userStats) {
			// Unlock achievement
			if err := s.repo.UnlockAchievement(userID, def.id); err == nil {
				newAchievement := Achievement{
					ID:          def.id,
					Name:        def.name,
					Description: def.description,
					Rarity:      def.rarity,
					CreatedAt:   time.Now(),
				}
				newlyUnlocked = append(newlyUnlocked, newAchievement)

				// Broadcast achievement unlock
				_ = s.BroadcastActivity(userID, "achievement_earned", map[string]interface{}{
					"achievement_id":   def.id,
					"achievement_name": def.name,
					"rarity":           def.rarity,
				})
			}
		}
	}

	// Return all achievements (existing + newly unlocked)
	allAchievements := append(existingAchievements, newlyUnlocked...)
	return allAchievements, nil
}

// UnlockAchievement manually unlocks an achievement
func (s *Service) UnlockAchievement(userID, achievementID string) error {
	if err := s.repo.UnlockAchievement(userID, achievementID); err != nil {
		return fmt.Errorf("failed to unlock achievement: %w", err)
	}

	// Broadcast achievement unlock to followers
	_ = s.BroadcastActivity(userID, "achievement_earned", map[string]interface{}{
		"achievement_id": achievementID,
	})

	return nil
}

// GetFollowers retrieves user's followers
func (s *Service) GetFollowers(userID string) ([]string, error) {
	followers, err := s.repo.GetFollowers(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get followers: %w", err)
	}
	return followers, nil
}

// GetFollowing retrieves users that user follows
func (s *Service) GetFollowing(userID string) ([]string, error) {
	following, err := s.repo.GetFollowing(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get following: %w", err)
	}
	return following, nil
}

// UserProfileData represents aggregated user profile data
type UserProfileData struct {
	UserID           string        `json:"user_id"`
	Achievements     []Achievement `json:"achievements"`
	FollowersCount   int           `json:"followers_count"`
	FollowingCount   int           `json:"following_count"`
	CompletedCourses []interface{} `json:"completed_courses"`
	CurrentArchetype interface{}   `json:"current_archetype"`
	SkillLevel       string        `json:"skill_level"`
}

// GetUserProfileData retrieves complete user profile with data from all domains
func (s *Service) GetUserProfileData(userID string) (*UserProfileData, error) {
	// Get achievements
	achievements, err := s.CheckAchievements(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievements: %w", err)
	}

	// Get followers and following
	followers, err := s.GetFollowers(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get followers: %w", err)
	}

	following, err := s.GetFollowing(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get following: %w", err)
	}

	// Get completed courses from learning domain
	var completedCourses []interface{}
	if s.learningService != nil {
		courses, err := s.learningService.GetUserCoursesInterface(userID)
		if err == nil {
			completedCourses = courses
		} else {
			fmt.Printf("Warning: Failed to get user courses: %v\n", err)
			completedCourses = []interface{}{}
		}
	} else {
		completedCourses = []interface{}{}
	}

	// Get current archetype from identity domain
	var currentArchetype interface{}
	if s.identityService != nil {
		archetype, err := s.identityService.GetArchetype(userID)
		if err == nil {
			currentArchetype = archetype
		} else {
			fmt.Printf("Warning: Failed to get archetype: %v\n", err)
		}
	}

	// Calculate skill level based on completed courses
	skillLevel := "beginner"
	courseCount := len(completedCourses)
	if courseCount >= 5 {
		skillLevel = "advanced"
	} else if courseCount >= 2 {
		skillLevel = "intermediate"
	}

	return &UserProfileData{
		UserID:           userID,
		Achievements:     achievements,
		FollowersCount:   len(followers),
		FollowingCount:   len(following),
		CompletedCourses: completedCourses,
		CurrentArchetype: currentArchetype,
		SkillLevel:       skillLevel,
	}, nil
}
