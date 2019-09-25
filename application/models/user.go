package models

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/silinternational/wecarry-api/domain"

	"github.com/silinternational/wecarry-api/auth"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"

	"github.com/gofrs/uuid"
)

const (
	PostRoleCreatedby string = "PostsCreated"
	PostRoleReceiving string = "PostsReceiving"
	PostRoleProviding string = "PostsProviding"
)

type User struct {
	ID                int                `json:"id" db:"id"`
	CreatedAt         time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" db:"updated_at"`
	Email             string             `json:"email" db:"email"`
	FirstName         string             `json:"first_name" db:"first_name"`
	LastName          string             `json:"last_name" db:"last_name"`
	Nickname          string             `json:"nickname" db:"nickname"`
	AdminRole         nulls.String       `json:"admin_role" db:"admin_role"`
	Uuid              uuid.UUID          `json:"uuid" db:"uuid"`
	AccessTokens      []UserAccessToken  `has_many:"user_access_tokens" json:"-"`
	Organizations     Organizations      `many_to_many:"user_organizations" json:"-"`
	UserOrganizations []UserOrganization `has_many:"user_organizations" json:"-"`
	PostsCreated      Posts              `has_many:"posts" fk_id:"created_by_id"`
	PostsProviding    Posts              `has_many:"posts" fk_id:"provider_id"`
	PostsReceiving    Posts              `has_many:"posts" fk_id:"receiver_id"`
}

// String is not required by pop and may be deleted
func (u User) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Users is not required by pop and may be deleted
type Users []User

// String is not required by pop and may be deleted
func (u Users) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *User) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: u.Email, Name: "Email"},
		&validators.StringIsPresent{Field: u.FirstName, Name: "FirstName"},
		&validators.StringIsPresent{Field: u.LastName, Name: "LastName"},
		&validators.StringIsPresent{Field: u.Nickname, Name: "Nickname"},
		&validators.UUIDIsPresent{Field: u.Uuid, Name: "Uuid"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// CreateAccessToken - Create and store new UserAccessToken
func (u *User) CreateAccessToken(org Organization, clientID string) (string, int64, error) {
	if clientID == "" {
		return "", 0, fmt.Errorf("cannot create token with empty clientID for user %s", u.Nickname)
	}

	token := createAccessTokenPart()
	hash := HashClientIdAccessToken(clientID + token)
	expireAt := createAccessTokenExpiry()

	userOrg, err := u.FindUserOrganization(org)
	if err != nil {
		return "", 0, err
	}

	userAccessToken := &UserAccessToken{
		UserID:             u.ID,
		UserOrganizationID: userOrg.ID,
		AccessToken:        hash,
		ExpiresAt:          expireAt,
	}

	if err := DB.Save(userAccessToken); err != nil {
		return "", 0, err
	}

	return token, expireAt.UTC().Unix(), nil
}

func (u *User) GetOrgIDs() []int {
	// ignore the error and allow the user's Organizations to be an empty slice.
	_ = DB.Load(u, "Organizations")

	s := make([]int, len(u.Organizations))
	for i, v := range u.Organizations {
		s[i] = v.ID
	}

	return s
}

func (u *User) FindOrCreateFromAuthUser(orgID int, authUser *auth.User) error {
	var userOrgs UserOrganizations
	err := userOrgs.FindByAuthEmail(authUser.Email, orgID)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(userOrgs) > 1 {
		return fmt.Errorf("too many user organizations found (%v), data integrity problem", len(userOrgs))
	}

	if len(userOrgs) == 1 {
		if userOrgs[0].AuthID != authUser.UserID {
			return fmt.Errorf("a user in this organization with this email address already exists with different user id")
		}
		err = DB.Where("uuid = ?", userOrgs[0].User.Uuid).First(u)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	newUser := true
	if u.ID != 0 {
		newUser = false
	}

	// update attributes from authUser
	u.FirstName = authUser.FirstName
	u.LastName = authUser.LastName
	u.Email = authUser.Email
	u.Nickname = fmt.Sprintf("%s %s", authUser.FirstName, authUser.LastName[:1])

	// if new user they will need a uuid
	if newUser {
		u.Uuid = domain.GetUuid()
	}

	err = DB.Save(u)
	if err != nil {
		return fmt.Errorf("unable to create new user record: %s", err.Error())
	}

	if len(userOrgs) == 0 {
		userOrg := &UserOrganization{
			OrganizationID: orgID,
			UserID:         u.ID,
			Role:           UserOrganizationRoleMember,
			AuthID:         authUser.UserID,
			AuthEmail:      u.Email,
			LastLogin:      time.Now(),
		}
		err = DB.Save(userOrg)
		if err != nil {
			return fmt.Errorf("unable to create new user_organization record: %s", err.Error())
		}
	}

	// reload user
	// err = DB.Eager().Where("id = ?", u.ID).First(u)
	// if err != nil {
	// 	return fmt.Errorf("unable to reload user after update: %s", err)
	// }

	return nil
}

// CanCreateOrganization returns true if the given user is allowed to create organizations
func (u *User) CanCreateOrganization() bool {
	return u.AdminRole.String == domain.AdminRoleSuperDuperAdmin || u.AdminRole.String == domain.AdminRoleSalesAdmin
}

func (u *User) CanEditOrganization(orgId int) bool {
	// make sure we're checking current user orgs
	err := DB.Load(u, "UserOrganizations")
	if err != nil {
		return false
	}

	for _, uo := range u.UserOrganizations {
		if uo.OrganizationID == orgId && uo.Role == UserOrganizationRoleAdmin {
			return true
		}
	}

	return false
}

func (u *User) FindByAccessToken(accessToken string) error {
	if accessToken == "" {
		return fmt.Errorf("error: access token must not be blank")
	}

	var userAccessToken UserAccessToken
	err := userAccessToken.FindByBearerToken(accessToken)
	if err != nil {
		return fmt.Errorf("error finding user by access token: %s", err.Error())
	}

	if userAccessToken.ID == 0 {
		return fmt.Errorf("error finding user by access token")
	}

	if userAccessToken.ExpiresAt.Before(time.Now()) {
		err := DB.Destroy(&userAccessToken)
		if err != nil {
			log.Printf("Unable to delete expired userAccessToken, id: %v", userAccessToken.ID)
		}
		return fmt.Errorf("access token has expired")
	}

	*u = userAccessToken.User
	return nil
}

func (u *User) FindByUUID(uuid string) error {

	if uuid == "" {
		return fmt.Errorf("error: uuid must not be blank")
	}

	if err := DB.Where("uuid = ?", uuid).First(u); err != nil {
		return fmt.Errorf("error finding user by uuid: %s", err.Error())
	}

	return nil
}

func createAccessTokenExpiry() time.Time {
	lifetime := envy.Get("ACCESS_TOKEN_LIFETIME", "28800")

	lifetimeSeconds, err := strconv.Atoi(lifetime)
	if err != nil {
		lifetimeSeconds = 28800
	}

	dtNow := time.Now()
	futureTime := dtNow.Add(time.Second * time.Duration(lifetimeSeconds))

	return futureTime
}

func createAccessTokenPart() string {
	var alphanumerics = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	tokenLength := 32
	b := make([]rune, tokenLength)
	for i := range b {
		b[i] = alphanumerics[rand.Intn(len(alphanumerics))]
	}

	accessToken := string(b)

	return accessToken
}

// HashClientIdAccessToken just returns a sha256.Sum256 of the input value
func HashClientIdAccessToken(accessToken string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(accessToken)))
}

func (u *User) GetOrganizations() ([]*Organization, error) {
	if err := DB.Load(u, "Organizations"); err != nil {
		return []*Organization{}, fmt.Errorf("error getting organizations for user id %v ... %v", u.ID, err)
	}

	orgs := make([]*Organization, len(u.Organizations))
	for i := range u.Organizations {
		orgs[i] = &u.Organizations[i]
	}

	return orgs, nil
}

func (u *User) FindUserOrganization(org Organization) (UserOrganization, error) {
	var userOrg UserOrganization
	if err := DB.Where("user_id = ? AND organization_id = ?", u.ID, org.ID).First(&userOrg); err != nil {
		return UserOrganization{}, fmt.Errorf("association not found for user '%v' and org '%v' (%s)", u.Nickname, org.Name, err.Error())
	}

	return userOrg, nil
}

func (u *User) GetPosts(postRole string) ([]*Post, error) {
	var postPtrs []*Post
	if err := DB.Load(u, postRole); err != nil {
		return postPtrs, fmt.Errorf("error getting posts for user id %v ... %v", u.ID, err)
	}

	var posts Posts
	switch postRole {
	case PostRoleCreatedby:
		posts = u.PostsCreated

	case PostRoleReceiving:
		posts = u.PostsReceiving

	case PostRoleProviding:
		posts = u.PostsProviding
	}

	for _, p := range posts {
		p1 := p
		postPtrs = append(postPtrs, &p1)
	}

	return postPtrs, nil
}
