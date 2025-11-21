package identity

import (
	"database/sql"
)

// Repository handles identity data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new identity repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser inserts a new user
func (r *Repository) CreateUser(user *User) error {
	query := `
		INSERT INTO users (id, email, password_hash, name, avatar_url, created_at, updated_at, last_login)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.AvatarURL,
		user.CreatedAt,
		user.UpdatedAt,
		user.LastLogin,
	)
	return err
}

// GetUserByEmail retrieves user by email
func (r *Repository) GetUserByEmail(email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, name, avatar_url, created_at, updated_at, last_login
		FROM users
		WHERE email = $1
	`
	user := &User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByID retrieves user by ID
func (r *Repository) GetUserByID(id string) (*User, error) {
	query := `
		SELECT id, email, password_hash, name, avatar_url, created_at, updated_at, last_login
		FROM users
		WHERE id = $1
	`
	user := &User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Set default privacy settings
	user.PrivacySettings = &PrivacySettings{
		ProfileVisibility:    "friends",
		ActivityVisibility:   "friends",
		ProgressVisibility:   "friends",
		AllowFollowers:       true,
		ShowInLeaderboards:   true,
		ShowCompletedCourses: true,
	}

	return user, nil
}

// UpdateUser updates user information
func (r *Repository) UpdateUser(user *User) error {
	query := `
		UPDATE users
		SET name = $1, avatar_url = $2, updated_at = $3, last_login = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(
		query,
		user.Name,
		user.AvatarURL,
		user.UpdatedAt,
		user.LastLogin,
		user.ID,
	)
	return err
}

// CreateArchetype creates user archetype
func (r *Repository) CreateArchetype(archetype *UserArchetype) error {
	query := `
		INSERT INTO user_archetypes (id, user_id, meta_category, domain, skill_level, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(
		query,
		archetype.ID,
		archetype.UserID,
		archetype.MetaCategory,
		archetype.Domain,
		archetype.SkillLevel,
		archetype.CreatedAt,
		archetype.UpdatedAt,
	)
	return err
}

// GetArchetypeByUserID retrieves user's archetype
func (r *Repository) GetArchetypeByUserID(userID string) (*UserArchetype, error) {
	query := `
		SELECT id, user_id, meta_category, domain, skill_level, created_at, updated_at
		FROM user_archetypes
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	archetype := &UserArchetype{}
	err := r.db.QueryRow(query, userID).Scan(
		&archetype.ID,
		&archetype.UserID,
		&archetype.MetaCategory,
		&archetype.Domain,
		&archetype.SkillLevel,
		&archetype.CreatedAt,
		&archetype.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return archetype, nil
}

// CreateVariables creates user variables
func (r *Repository) CreateVariables(variables []UserVariable) error {
	if len(variables) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO user_variables (id, user_id, variable_key, variable_value, archetype_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, v := range variables {
		_, err := stmt.Exec(
			v.ID,
			v.UserID,
			v.VariableKey,
			v.VariableValue,
			v.ArchetypeID,
			v.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetVariablesByUserID retrieves user's variables
func (r *Repository) GetVariablesByUserID(userID string) ([]UserVariable, error) {
	query := `
		SELECT id, user_id, variable_key, variable_value, archetype_id, created_at
		FROM user_variables
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variables []UserVariable
	for rows.Next() {
		var v UserVariable
		err := rows.Scan(
			&v.ID,
			&v.UserID,
			&v.VariableKey,
			&v.VariableValue,
			&v.ArchetypeID,
			&v.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		variables = append(variables, v)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return variables, nil
}
