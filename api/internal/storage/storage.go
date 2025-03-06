package storage

import (
	"badJokes/internal/models"
	"badJokes/internal/storage/postgres"
	"badJokes/internal/storage/sqlite"
	"database/sql"
	"errors"
)

var (
	ErrObjectNotExist = errors.New("object does not exist")
)

type UserRepository interface {
	Register(username, email, password string) (int64, error)
	Authenticate(email, password string) (*models.User, error)
}

type JokesRepository interface {
	Insert(body string, authorID int64) (int64, error)
	ListPage(page, pageSize int, sortField, order string, currentUserID int64) ([]models.Joke, error)
	AddVote(entityType string, entityID, userID int64, voteType string) error
	RemoveVote(entityType string, entityID, userID int64) error
	GetVote(entityType string, entityID, userID int64) (string, error)
	AddReaction(entityType string, entityID, userID int64, reactionType string) error
	RemoveReaction(entityType string, entityID, userID int64, reactionType string) error
	GetReaction(entityType string, entityID, userID int64, reactionType string) (bool, error)
	DeleteJoke(jokeID int64) error
	GetJokeByID(jokeID int64) (*models.Joke, error)
	AddComment(jokeID, userID int64, body string) (int64, error)
	GetComments(jokeID int64) ([]models.Comment, error)
}

func NewUserRepository(dbType string, dbConn *sql.DB) UserRepository {
	switch dbType {
	case "postgres":
		return postgres.NewUserRepository(dbConn)
	case "sqlite":
		return sqlite.NewUserRepository(dbConn)
	default:
		panic("unsupported database type")
	}
}

func NewJokesRepository(dbType string, dbConn *sql.DB) JokesRepository {
	switch dbType {
	case "postgres":
		return postgres.NewJokesRepository(dbConn)
	case "sqlite":
		return sqlite.NewJokesRepository(dbConn)
	default:
		panic("unsupported database type")
	}
}
