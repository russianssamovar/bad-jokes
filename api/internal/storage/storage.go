package storage

import (
    "badJokes/internal/models"
    "badJokes/internal/storage/postgres"
    "badJokes/internal/storage/sqlite"
    "database/sql"
)

type UserRepository interface {
    Register(username, email, password string) (int64, error)
    Authenticate(email, password string) (*models.User, error)
}

type JokesRepository interface {
    Insert(body string, authorID int64) (int64, error)
    ListPage(page, pageSize int, sortField, order string, currentUserID int64) ([]models.Joke, error)
    GetJokeByID(jokeID, currentUserID int64) (models.Joke, error)
    DeleteJoke(jokeID int64) error
}

type CommentsRepository interface {
    AddComment(jokeID, userID int64, body string, parentID *int64) (int64, error)
    GetComments(jokeID int64) ([]models.Comment, error)
    GetCommentsByJokeID(jokeID, currentUserID int64) ([]models.Comment, error)
    DeleteComment(commentID int64) error
    GetCommentByID(commentID int64) (models.Comment, error)
}

type EntityRepository interface {
    AddVote(entityType string, entityID, userID int64, voteType string) error
    RemoveVote(entityType string, entityID, userID int64) error
    GetVote(entityType string, entityID, userID int64) (string, error)
    AddReaction(entityType string, entityID, userID int64, reactionType string) error
    RemoveReaction(entityType string, entityID, userID int64, reactionType string) error
    GetReaction(entityType string, entityID, userID int64, reactionType string) (bool, error)
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

func NewCommentsRepository(dbType string, dbConn *sql.DB) CommentsRepository {
    switch dbType {
    case "postgres":
        return postgres.NewCommentsRepository(dbConn)
    case "sqlite":
        return sqlite.NewCommentsRepository(dbConn)
    default:
        panic("unsupported database type")
    }
}

func NewEntityRepository(dbType string, dbConn *sql.DB) EntityRepository {
    switch dbType {
    case "postgres":
        return postgres.NewEntityRepository(dbConn)
    case "sqlite":
        return sqlite.NewEntityRepository(dbConn)
    default:
        panic("unsupported database type")
    }
}