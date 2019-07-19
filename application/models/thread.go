package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type Thread struct {
	ID           int       `json:"id" db:"id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	Uuid         uuid.UUID `json:"uuid" db:"uuid"`
	PostID       int       `json:"post_id" db:"post_id"`
	Post         Post      `belongs_to:"posts"`
	Messages     Messages  `has_many:"messages"`
	Participants Users     `has_many:"users"`
}

// String is not required by pop and may be deleted
func (t Thread) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Threads is not required by pop and may be deleted
type Threads []Thread

// String is not required by pop and may be deleted
func (t Threads) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (t *Thread) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: t.ID, Name: "ID"},
		&validators.UUIDIsPresent{Field: t.Uuid, Name: "Uuid"},
		&validators.IntIsPresent{Field: t.PostID, Name: "PostID"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (t *Thread) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (t *Thread) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

func FindThreadByUUID(uuid string) (Thread, error) {

	if uuid == "" {
		return Thread{}, fmt.Errorf("error: thread uuid must not be blank")
	}

	thread := Thread{}
	queryString := fmt.Sprintf("uuid = '%s'", uuid)

	if err := DB.Where(queryString).First(&thread); err != nil {
		return Thread{}, fmt.Errorf("error finding thread by uuid: %s", err.Error())
	}

	return thread, nil
}

func FindThreadByPostIDAndUserID(postID int, userID int) (Thread, error) {

	if postID == 0 || userID == 0 {
		err := fmt.Errorf("error: post postID and userID must not be 0. Got: %v and %v", postID, userID)
		return Thread{}, err
	}

	threads := []Thread{}

	if err := DB.Q().LeftJoin("thread_participants tp", "threads.id = tp.thread_id").
		Where("tp.user_id = ?", userID).All(&threads); err != nil {
		fmt.Errorf("Error getting threads: %v", err.Error())
		return Thread{}, err
	}

	fmt.Printf("\nAAAAAAB threads %v ... userID: %v\n", len(threads), userID)

	// TODO Rewrite this to do it the proper way
	for _, t := range threads {
		if t.PostID == postID {
			return t, nil
		}
	}

	return Thread{}, nil

}
