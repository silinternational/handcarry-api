package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/silinternational/wecarry-api/domain"
)

// Meeting represents an event where people gather together from different locations
type Meeting struct {
	ID          int          `json:"-" db:"id"`
	UUID        uuid.UUID    `json:"uuid" db:"uuid"`
	Name        string       `json:"name" db:"name"`
	Description nulls.String `json:"description" db:"description"`
	MoreInfoURL nulls.String `json:"more_info_url" db:"more_info_url"`
	StartDate   time.Time    `json:"start_date" db:"start_date"`
	EndDate     time.Time    `json:"end_date" db:"end_date"`
	InviteCode  nulls.UUID   `json:"invite_code" db:"invite_code"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	CreatedByID int          `json:"created_by_id" db:"created_by_id"`
	FileID      nulls.Int    `json:"file_id" db:"file_id"`
	LocationID  int          `json:"location_id" db:"location_id"`

	CreatedBy User     `json:"-" belongs_to:"users" fk_id:"CreatedByID"`
	ImgFile   *File    `json:"-" belongs_to:"files" fk_id:"FileID"`
	Location  Location `json:"-" belongs_to:"locations"`
}

// String is not required by pop and may be deleted
func (m Meeting) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Meetings is not required by pop and may be deleted
type Meetings []Meeting

// String is not required by pop and may be deleted
func (m Meetings) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (m *Meeting) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Field: m.UUID, Name: "UUID"},
		&validators.StringIsPresent{Field: m.Name, Name: "Name"},
		&validators.TimeIsPresent{Field: m.StartDate, Name: "StartDate"},
		&validators.TimeIsPresent{Field: m.EndDate, Name: "EndDate"},
		&validators.IntIsPresent{Field: m.CreatedByID, Name: "CreatedByID"},
		&validators.IntIsPresent{Field: m.LocationID, Name: "LocationID"},
		&dateValidator{StartDate: m.StartDate, EndDate: m.EndDate, Name: "Dates"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (m *Meeting) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (m *Meeting) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

type dateValidator struct {
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Message   string
}

func (v *dateValidator) IsValid(errors *validate.Errors) {
	if v.StartDate.Before(v.EndDate) {
		return
	}

	if v.StartDate.Format(domain.DateFormat) == v.EndDate.Format(domain.DateFormat) {
		return
	}

	v.Message = fmt.Sprintf("Start date must come no later than end date chronologically. Got %s and %s.",
		v.StartDate.Format(domain.DateFormat), v.EndDate.Format(domain.DateFormat))
	errors.Add(validators.GenerateKey(v.Name), v.Message)
}

// FindByUUID finds a meeting by the UUID field and loads its CreatedBy field
func (m *Meeting) FindByUUID(tx *pop.Connection, uuid string) error {
	if uuid == "" {
		return errors.New("error finding meeting: uuid must not be blank")
	}

	if err := tx.Where("uuid = ?", uuid).First(m); err != nil {
		return fmt.Errorf("error finding meeting by uuid: %s", err.Error())
	}

	return nil
}

func getOrdered(m *Meetings, q *pop.Query) error {
	return q.Order("start_date asc").All(m)
}

// FindOnDate finds the meetings that have StartDate before timeInFocus-date and an EndDate after it
// (inclusive on both)
func (m *Meetings) FindOnDate(tx *pop.Connection, timeInFocus time.Time) error {
	date := timeInFocus.Format(domain.DateTimeFormat)
	where := "start_date <= ? and end_date >= ?"

	if err := getOrdered(m, tx.Where(where, date, date)); err != nil {
		return fmt.Errorf("error finding meeting with start_date and end_date straddling %s ... %s",
			date, err.Error())
	}

	return nil
}

// FindOnOrAfterDate finds the meetings that have an EndDate on or after the timeInFocus-date
func (m *Meetings) FindOnOrAfterDate(tx *pop.Connection, timeInFocus time.Time) error {
	date := timeInFocus.Format(domain.DateTimeFormat)

	if err := getOrdered(m, tx.Where("end_date >= ?", date)); err != nil {
		return fmt.Errorf("error finding meeting with end_date before %s ... %s", date, err.Error())
	}

	return nil
}

// FindAfterDate finds the meetings that have a StartDate after the timeInFocus-date
func (m *Meetings) FindAfterDate(tx *pop.Connection, timeInFocus time.Time) error {
	date := timeInFocus.Format(domain.DateTimeFormat)

	if err := getOrdered(m, tx.Where("start_date > ?", date)); err != nil {
		return fmt.Errorf("error finding meeting with start_date after %s ... %s", date, err.Error())
	}

	return nil
}

// FindRecent finds the meetings that have an EndDate within the past <domain.RecentMeetingDelay> days
// before timeInFocus-date (not inclusive)
func (m *Meetings) FindRecent(tx *pop.Connection, timeInFocus time.Time) error {
	yesterday := timeInFocus.Add(-domain.DurationDay).Format(domain.DateTimeFormat)
	recentDate := timeInFocus.Add(-domain.RecentMeetingDelay)
	where := "end_date between ? and ?"

	if err := getOrdered(m, tx.Where(where, recentDate, yesterday)); err != nil {
		return fmt.Errorf("error finding meeting with end_date between %s and %s ... %s",
			recentDate, yesterday, err.Error())
	}

	return nil
}

func (m *Meeting) FindByInviteCode(tx *pop.Connection, code string) error {
	if code == "" {
		return errors.New("error finding meeting: invite_code must not be blank")
	}

	if err := tx.Where("invite_code = ?", code).First(m); err != nil {
		return fmt.Errorf("error finding meeting by invite_code: %s", err.Error())
	}

	return nil
}

// SetImageFile assigns a previously-stored File to this Meeting as its image. Parameter `fileID` is the UUID
// of the image to attach.
func (m *Meeting) SetImageFile(tx *pop.Connection, fileID string) (File, error) {
	return addFile(tx, m, fileID)
}

// ImageFile retrieves the file attached as the Meeting Image
func (m *Meeting) ImageFile(tx *pop.Connection) (*File, error) {
	if !m.FileID.Valid {
		return nil, nil
	}
	if m.ImgFile == nil {
		if err := tx.Load(m, "ImgFile"); err != nil {
			return nil, err
		}
	}
	if err := (*m.ImgFile).RefreshURL(tx); err != nil {
		return nil, err
	}
	f := *m.ImgFile
	return &f, nil
}

// RemoveFile removes an attached file from the Meeting
func (m *Meeting) RemoveFile(tx *pop.Connection) error {
	return removeFile(tx, m)
}

func (m *Meeting) GetCreator(tx *pop.Connection) (*User, error) {
	creator := User{}
	if err := tx.Find(&creator, m.CreatedByID); err != nil {
		return nil, err
	}
	return &creator, nil
}

// GetLocation returns the related Location object.
func (m *Meeting) GetLocation(tx *pop.Connection) (Location, error) {
	location := Location{}
	if err := tx.Find(&location, m.LocationID); err != nil {
		return location, err
	}

	return location, nil
}

// Create stores the Meeting data as a new record in the database.
func (m *Meeting) Create(tx *pop.Connection) error {
	return create(tx, m)
}

// Update writes the Meeting data to an existing database record.
func (m *Meeting) Update(tx *pop.Connection) error {
	return update(tx, m)
}

// SetLocation sets the location field, creating a new record in the database if necessary.
func (m *Meeting) SetLocation(tx *pop.Connection, location Location) error {
	location.ID = m.LocationID
	m.Location = location
	return m.Location.Update(tx)
}

// CanCreate returns a bool based on whether the current user is allowed to create a meeting
func (m *Meeting) CanCreate(user User) bool {
	return true
}

// CanUpdate returns a bool based on whether the current user is allowed to update a meeting
func (m *Meeting) CanUpdate(user User) bool {
	switch user.AdminRole {
	case UserAdminRoleSuperAdmin, UserAdminRoleSalesAdmin, UserAdminRoleAdmin:
		return true
	}

	return user.ID == m.CreatedByID
}

// Requests return all associated Requests
func (m *Meeting) Requests(tx *pop.Connection) (Requests, error) {
	var requests Requests
	if err := tx.Where("meeting_id = ?", m.ID).Order("updated_at desc").All(&requests); err != nil {
		return nil, fmt.Errorf("error getting requests for meeting id %v ... %v", m.ID, err)
	}

	return requests, nil
}

// Invites returns all of the MeetingInvites for this Meeting. Only the meeting creator and organizers are authorized.
func (m *Meeting) Invites(tx *pop.Connection, user User) (MeetingInvites, error) {
	i := MeetingInvites{}
	if m == nil {
		return i, nil
	}

	isOrganizer, err := m.isOrganizer(tx, user.ID)
	if err != nil {
		return i, err
	}

	if user.ID != m.CreatedByID && !isOrganizer && !user.isSuperAdmin() {
		return i, nil
	}
	if err := tx.Where("meeting_id = ?", m.ID).All(&i); err != nil {
		return i, err
	}
	return i, nil
}

// Participants returns all of the MeetingParticipants for this Meeting. Only the meeting creator and organizers are
// authorized.
func (m *Meeting) Participants(tx *pop.Connection, user User) (MeetingParticipants, error) {
	p := MeetingParticipants{}
	if m == nil {
		return p, nil
	}

	isOrganizer, err := m.isOrganizer(tx, user.ID)
	if err != nil {
		return p, err
	}

	if user.ID != m.CreatedByID && !isOrganizer && !user.isSuperAdmin() {
		return p, tx.Where("user_id = ? AND meeting_id = ?", user.ID, m.ID).All(&p)
	}
	if err := tx.Where("meeting_id = ?", m.ID).All(&p); err != nil {
		return p, err
	}
	return p, nil
}

// Organizers returns all of the users who are organizers for this Meeting. No authorization is checked, so
// any queries should render this as a PublicProfile to limit field visibility.
func (m *Meeting) Organizers(tx *pop.Connection) (Users, error) {
	u := Users{}
	if m == nil {
		return u, nil
	}
	if err := tx.
		Select("users.id", "users.uuid", "nickname", "file_id", "auth_photo_url").
		Where("meeting_participants.is_organizer=true").
		Where("meeting_participants.meeting_id=?", m.ID).
		Join("meeting_participants", "meeting_participants.user_id=users.id").
		All(&u); err != nil {
		return u, err
	}
	return u, nil
}

func (m *Meeting) RemoveInvite(tx *pop.Connection, email string) error {
	var invite MeetingInvite
	if err := invite.FindByMeetingIDAndEmail(tx, m.ID, email); err != nil {
		return err
	}
	return invite.Destroy(tx)
}

func (m *Meeting) RemoveParticipant(tx *pop.Connection, userUUID string) error {
	var user User
	if err := user.FindByUUID(tx, userUUID); err != nil {
		return fmt.Errorf("invalid user ID %s in Meeting.RemoveParticipant, %s", userUUID, err)
	}
	var participant MeetingParticipant
	if err := participant.FindByMeetingIDAndUserID(tx, m.ID, user.ID); err != nil {
		return fmt.Errorf("failed to load MeetingParticipant in Meeting.RemoveParticipant, %s", err)
	}
	return participant.Destroy(tx)
}

func (m *Meeting) IsCodeValid(tx *pop.Connection, code string) bool {
	if m.InviteCode.Valid && m.InviteCode.UUID.String() == code {
		return true
	}
	return false
}

func (m *Meeting) isOrganizer(tx *pop.Connection, userID int) (bool, error) {
	organizers, err := m.Organizers(tx)
	if err != nil {
		return false, errors.New("isOrganizer() error reading list of meeting organizers, " + err.Error())
	}
	for _, o := range organizers {
		if o.ID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (m *Meeting) isVisible(tx *pop.Connection, userID int) bool {
	return true
}

// FindByIDs finds all Meetings associated with the given IDs and loads them from the database
func (m *Meetings) FindByIDs(tx *pop.Connection, ids []int) error {
	ids = domain.UniquifyIntSlice(ids)
	return tx.Where("id in (?)", ids).All(m)
}
