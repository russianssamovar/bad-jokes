package models

type User struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	IsAdmin    bool   `json:"is_admin"`
	CreatedAt  string `json:"created_at"`
	ModifiedAt string `json:"modified_at"`
}

type ModerationLog struct {
	ID           int64  `json:"id"`
	Action       string `json:"action"`
	TargetID     int64  `json:"target_id"`
	TargetType   string `json:"target_type"`
	PerformedBy  int64  `json:"performed_by"`
	AdminUsername string `json:"admin_username"`
	Details      string `json:"details"`
	CreatedAt    string `json:"created_at"`
}

type UserStats struct {
	TotalUsers        int            `json:"total_users"`
	AdminCount        int            `json:"admin_count"`
	NewUsersToday     int            `json:"new_users_today"`
	NewUsersThisWeek  int            `json:"new_users_this_week"`
	NewUsersThisMonth int            `json:"new_users_this_month"`
	MostActiveUsers   []*ActiveUser  `json:"most_active_users"`
}

type ActiveUser struct {
	ID            int64  `json:"id"`
	Username      string `json:"username"`
	JokesCount    int    `json:"jokes_count"`
	CommentsCount int    `json:"comments_count"`
}