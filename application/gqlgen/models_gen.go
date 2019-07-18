// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package gqlgen

import (
	"fmt"
	"io"
	"strconv"
)

type Message struct {
	ID        string  `json:"id"`
	Sender    *User   `json:"sender"`
	Content   string  `json:"content"`
	CreatedAt *string `json:"createdAt"`
	UpdatedAt *string `json:"updatedAt"`
}

type NewMessage struct {
	SenderID string  `json:"senderID"`
	Content  string  `json:"content"`
	PostID   string  `json:"postID"`
	ThreadID *string `json:"threadID"`
}

type NewPost struct {
	CreatedByID  string   `json:"createdByID"`
	OrgID        string   `json:"orgID"`
	Type         PostType `json:"type"`
	Title        string   `json:"title"`
	Description  *string  `json:"description"`
	Destination  *string  `json:"destination"`
	Origin       *string  `json:"origin"`
	Size         string   `json:"size"`
	NeededAfter  *string  `json:"neededAfter"`
	NeededBefore *string  `json:"neededBefore"`
	Category     *string  `json:"category"`
}

type Organization struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	URL       *string `json:"url"`
	CreatedAt *string `json:"createdAt"`
	UpdatedAt *string `json:"updatedAt"`
}

type Post struct {
	ID           string        `json:"id"`
	UUID         string        `json:"uuid"`
	Type         PostType      `json:"type"`
	CreatedBy    *User         `json:"createdBy"`
	Receiver     *User         `json:"receiver"`
	Provider     *User         `json:"provider"`
	Organization *Organization `json:"organization"`
	Title        string        `json:"title"`
	Description  *string       `json:"description"`
	Destination  *string       `json:"destination"`
	Origin       *string       `json:"origin"`
	Size         string        `json:"size"`
	NeededAfter  *string       `json:"neededAfter"`
	NeededBefore *string       `json:"neededBefore"`
	Category     string        `json:"category"`
	Status       string        `json:"status"`
	Thread       []*Thread     `json:"thread"`
	CreatedAt    *string       `json:"createdAt"`
	UpdatedAt    *string       `json:"updatedAt"`
}

type Thread struct {
	ID           string     `json:"id"`
	Participants []*User    `json:"participants"`
	Messages     []*Message `json:"messages"`
	PostID       string     `json:"postID"`
	CreatedAt    *string    `json:"createdAt"`
	UpdatedAt    *string    `json:"updatedAt"`
}

type User struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	FirstName   string  `json:"firstName"`
	LastName    string  `json:"lastName"`
	Nickname    string  `json:"nickname"`
	UUID        string  `json:"uuid"`
	AccessToken string  `json:"accessToken"`
	CreatedAt   *string `json:"createdAt"`
	UpdatedAt   *string `json:"updatedAt"`
	AdminRole   *Role   `json:"adminRole"`
}

type PostType string

const (
	PostTypeRequest PostType = "REQUEST"
	PostTypeOffer   PostType = "OFFER"
)

var AllPostType = []PostType{
	PostTypeRequest,
	PostTypeOffer,
}

func (e PostType) IsValid() bool {
	switch e {
	case PostTypeRequest, PostTypeOffer:
		return true
	}
	return false
}

func (e PostType) String() string {
	return string(e)
}

func (e *PostType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PostType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PostType", str)
	}
	return nil
}

func (e PostType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type Role string

const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
)

var AllRole = []Role{
	RoleAdmin,
	RoleUser,
}

func (e Role) IsValid() bool {
	switch e {
	case RoleAdmin, RoleUser:
		return true
	}
	return false
}

func (e Role) String() string {
	return string(e)
}

func (e *Role) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Role(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Role", str)
	}
	return nil
}

func (e Role) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
