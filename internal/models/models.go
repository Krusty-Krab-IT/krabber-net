package models

import (
	"errors"
)

// Define a custom ErrRecordNotFound error. We'll return this from our Get() method when
// looking up a movie that doesn't exist in our database.
var (
	ErrRecordNotFound = errors.New("record not found")
)

// Create a Models struct which wraps the MovieModel. We'll add other models to this,
// like a UserModel and PermissionModel, as our build progresses.
type Models struct {
	Molts    MoltModel
	Crabs    CrabModel
	Tokens   TokenModel
	Follows  FollowModel
	Comments CommentModel
	Likes    LikesModel
	Trench   TrenchModel
}

// For ease of use, we also add a New() method which returns a Models struct containing
// the initialized MovieModel.
func NewModels(db ItemService) Models {
	return Models{
		Crabs:    CrabModel{SVC: db},
		Molts:    MoltModel{SVC: db},
		Tokens:   TokenModel{SVC: db},
		Follows:  FollowModel{SVC: db},
		Comments: CommentModel{SVC: db},
		Likes:    LikesModel{SVC: db},
		Trench:   TrenchModel{SVC: db},
	}
}
