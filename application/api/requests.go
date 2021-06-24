package api

import "github.com/gofrs/uuid"

// swagger:model
type Requests []Request

// Request is a hand carry request
//
// swagger:model
type Request struct {
	// unique id (uuid) for thread
	//
	// swagger:strfmt uuid4
	// unique: true
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	ID uuid.UUID `json:"id"`

	// Description of this request
	Description string `json:"description"`

	// Profile of the user that created this request.
	CreatedBy *User `json:"created_by"`

	// Whether request is editable by current user
	IsEditable bool `json:"isEditable"`

	// Request status: OPEN, ACCEPTED, DELIVERED, RECEIVED, COMPLETED, REMOVED
	RequestStatus string `json:"status"`
}
