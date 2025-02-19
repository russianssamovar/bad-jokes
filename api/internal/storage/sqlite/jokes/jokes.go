package jokes

type Joke struct {
	ID           int64              `json:"id"`
	Title        string             `json:"title"`
	Body         string             `json:"body"`
	AuthorID     int64              `json:"author_id"`
	CreatedAt    string             `json:"created_at"`
	ModifiedAt   string             `json:"modified_at"`
	Social       SocialInteractions `json:"social"`
	CommentCount int                `json:"comment_count"`
}

type Comment struct {
	ID              int64  `json:"id"`
	JokeID          int64  `json:"joke_id"`
	ParentCommentID *int64 `json:"parent_comment_id,omitempty"`
	UserID          int64  `json:"user_id"`
	Body            string `json:"body"`
	CreatedAt       string `json:"created_at"`
	ModifiedAt      string `json:"modified_at"`
	IsAuthor        bool   `json:"is_author"`
}

type SocialInteractions struct {
	Pluses    int              `json:"pluses"`
	Minuses   int              `json:"minuses"`
	Reactions map[string]int   `json:"reactions"`
	User      *UserInteraction `json:"user,omitempty"`
}

type UserInteraction struct {
	VoteType  string   `json:"vote_type,omitempty"`
	Reactions []string `json:"reactions,omitempty"`
}
