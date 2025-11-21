package social

import (
	"time"
)

// UserRelationship represents follow relationship
type UserRelationship struct {
	ID          string
	FollowerID  string
	FollowingID string
	CreatedAt   time.Time
}

// ActivityFeed represents activity ticker item
type ActivityFeed struct {
	ID            string
	UserID        string
	ActivityType  string
	ReferenceType string
	ReferenceID   string
	Metadata      interface{}
	Visibility    string
	CreatedAt     time.Time
}

// Achievement represents achievement definition
type Achievement struct {
	ID          string
	Name        string
	Description string
	BadgeIcon   string
	Criteria    interface{}
	Rarity      string
	CreatedAt   time.Time
}

// UserAchievement represents earned achievement
type UserAchievement struct {
	ID            string
	UserID        string
	AchievementID string
	UnlockedAt    time.Time
}

// Recommendation represents course recommendation
type Recommendation struct {
	ID                 string
	UserID             string
	CourseID           string
	RecommendationType string
	MatchScore         int
	Reason             string
	Metadata           interface{}
	CreatedAt          time.Time
	ExpiresAt          *time.Time
}

// TrendingCourse represents trending course data
type TrendingCourse struct {
	ID                   string
	CourseID             string
	Velocity             float64
	Signups24h           int
	SignupsPrevious24h   int
	Rank                 int
	MetaCategory         string
	CalculatedAt         time.Time
}
