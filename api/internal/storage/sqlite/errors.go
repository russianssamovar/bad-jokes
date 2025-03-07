package sqlite

import "fmt"

var ErrJokeNotFound = fmt.Errorf("joke not found")
var ErrCommentNotFound = fmt.Errorf("comment not found")