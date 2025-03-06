package models

type Joke struct {
	ID             int64              `json:"id"`
	Title          string             `json:"title"`
	Body           string             `json:"body"`
	AuthorID       int64              `json:"author_id"`
	AuthorUsername string             `json:"author_username"`
	CreatedAt      string             `json:"created_at"`
	ModifiedAt     string             `json:"modified_at"`
	Social         SocialInteractions `json:"social"`
	CommentCount   int                `json:"comment_count"`
}

type Comment struct {
	ID             int64              `json:"id"`
	JokeID         int64              `json:"joke_id"`
	ParentID       int64              `json:"parent_id,omitempty"`
	UserID         int64              `json:"user_id"`
	Body           string             `json:"body"`
	CreatedAt      string             `json:"created_at"`
	ModifiedAt     string             `json:"modified_at"`
	IsAuthor       bool               `json:"is_author"`
	Social         SocialInteractions `json:"social"`
	AuthorID       int64              `json:"author_id"`
	AuthorUsername string             `json:"author_username"`
    IsDeleted      bool               `json:"is_deleted"`
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

type JokeWithComments struct {
	Joke     Joke      `json:"joke"`
	Comments []Comment `json:"comments"`
}
