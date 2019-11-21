// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package gqlgen

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/silinternational/wecarry-api/models"
)

type CreateMessageInput struct {
	Content  string  `json:"content"`
	PostID   string  `json:"postID"`
	ThreadID *string `json:"threadID"`
}

type CreateOrganizationDomainInput struct {
	Domain         string `json:"domain"`
	OrganizationID string `json:"organizationID"`
}

type CreateOrganizationInput struct {
	Name       string  `json:"name"`
	URL        *string `json:"url"`
	AuthType   string  `json:"authType"`
	AuthConfig string  `json:"authConfig"`
}

type LocationInput struct {
	Description string   `json:"description"`
	Country     string   `json:"country"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
}

type RemoveOrganizationDomainInput struct {
	Domain         string `json:"domain"`
	OrganizationID string `json:"organizationID"`
}

type SetThreadLastViewedAtInput struct {
	ThreadID string    `json:"threadID"`
	Time     time.Time `json:"time"`
}

type UpdateOrganizationInput struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	URL        *string `json:"url"`
	AuthType   string  `json:"authType"`
	AuthConfig string  `json:"authConfig"`
}

type UpdatePostStatusInput struct {
	ID     string            `json:"id"`
	Status models.PostStatus `json:"status"`
}

type UpdateUserInput struct {
	ID       *string        `json:"id"`
	PhotoID  *string        `json:"photoID"`
	Location *LocationInput `json:"location"`
}

type PostRole string

const (
	PostRoleCreatedby PostRole = "CREATEDBY"
	PostRoleReceiving PostRole = "RECEIVING"
	PostRoleProviding PostRole = "PROVIDING"
)

var AllPostRole = []PostRole{
	PostRoleCreatedby,
	PostRoleReceiving,
	PostRoleProviding,
}

func (e PostRole) IsValid() bool {
	switch e {
	case PostRoleCreatedby, PostRoleReceiving, PostRoleProviding:
		return true
	}
	return false
}

func (e PostRole) String() string {
	return string(e)
}

func (e *PostRole) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PostRole(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PostRole", str)
	}
	return nil
}

func (e PostRole) MarshalGQL(w io.Writer) {
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
