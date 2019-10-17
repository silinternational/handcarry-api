package models

import (
	"reflect"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/validate"
)

func (ms *ModelSuite) TestPost_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		post     Post
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			post: Post{
				CreatedByID:    1,
				Type:           PostTypeRequest,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
				Uuid:           domain.GetUuid(),
			},
			wantErr: false,
		},
		{
			name: "missing created_by",
			post: Post{
				Type:           PostTypeRequest,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "created_by",
		},
		{
			name: "missing type",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "type",
		},
		{
			name: "missing organization_id",
			post: Post{
				CreatedByID: 1,
				Type:        PostTypeRequest,
				Title:       "A Request",
				Size:        PostSizeMedium,
				Status:      PostStatusOpen,
				Uuid:        domain.GetUuid(),
			},
			wantErr:  true,
			errField: "organization_id",
		},
		{
			name: "missing title",
			post: Post{
				CreatedByID:    1,
				Type:           PostTypeRequest,
				OrganizationID: 1,
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "title",
		},
		{
			name: "missing size",
			post: Post{
				CreatedByID:    1,
				Type:           PostTypeRequest,
				OrganizationID: 1,
				Title:          "A Request",
				Status:         PostStatusOpen,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "size",
		},
		{
			name: "missing status",
			post: Post{
				CreatedByID:    1,
				Type:           PostTypeRequest,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "missing uuid",
			post: Post{
				CreatedByID:    1,
				Type:           PostTypeRequest,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
			},
			wantErr:  true,
			errField: "uuid",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.post.Validate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(test.errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", test.errField, vErr.Errors)
				}
			} else if (test.wantErr == false) && (vErr.HasAny()) {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_ValidateCreate() {
	t := ms.T()

	tests := []struct {
		name     string
		post     Post
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "good - open",
			post: Post{
				CreatedByID:    1,
				Type:           PostTypeRequest,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
				Uuid:           domain.GetUuid(),
			},
			wantErr: false,
		},
		{
			name: "bad status - accepted",
			post: Post{
				CreatedByID:    1,
				Type:           PostTypeRequest,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusAccepted,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - committed",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Type:           PostTypeRequest,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusCommitted,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - received",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Type:           PostTypeRequest,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusReceived,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - completed",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Type:           PostTypeRequest,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusCompleted,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - removed",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Type:           PostTypeRequest,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusRemoved,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "create_status",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.post.ValidateCreate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(test.errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", test.errField, vErr.Errors)
				}
			} else if (test.wantErr == false) && (vErr.HasAny()) {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func CreateFixturesValidateUpdate(ms *ModelSuite, t *testing.T) []Post {
	ResetTables(t, ms.DB)

	// Create org
	org := &Organization{
		ID:         1,
		Name:       "TestOrg",
		Url:        nulls.String{},
		AuthType:   AuthTypeSaml,
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	err := ms.DB.Create(org)
	if err != nil {
		t.Errorf("Failed to create organization for test, error: %s", err)
		t.FailNow()
	}

	// Create User
	user := User{
		Email:     "user1@example.com",
		FirstName: "Existing",
		LastName:  "User",
		Nickname:  "Existing User ",
		Uuid:      domain.GetUuid(),
	}

	if err := ms.DB.Create(&user); err != nil {
		t.Errorf("could not create test user %v ... %v", user.Email, err)
		t.FailNow()
	}

	// Load Post test fixtures
	posts := []Post{
		{
			CreatedByID:    user.ID,
			OrganizationID: org.ID,
			Type:           PostTypeRequest,
			Title:          "Open Request 0",
			Size:           PostSizeMedium,
			Status:         PostStatusOpen,
			Uuid:           domain.GetUuid(),
		},
		{
			CreatedByID:    user.ID,
			OrganizationID: org.ID,
			Type:           PostTypeRequest,
			Title:          "Committed Request 1",
			Size:           PostSizeMedium,
			Status:         PostStatusCommitted,
			Uuid:           domain.GetUuid(),
		},
		{
			CreatedByID:    user.ID,
			OrganizationID: org.ID,
			Type:           PostTypeRequest,
			Title:          "Accepted Request 2",
			Size:           PostSizeMedium,
			Status:         PostStatusAccepted,
			Uuid:           domain.GetUuid(),
		},
		{
			CreatedByID:    user.ID,
			OrganizationID: org.ID,
			Type:           PostTypeRequest,
			Title:          "Received Request 3",
			Size:           PostSizeMedium,
			Status:         PostStatusReceived,
			Uuid:           domain.GetUuid(),
		},
		{
			CreatedByID:    user.ID,
			OrganizationID: org.ID,
			Type:           PostTypeRequest,
			Title:          "Completed Request 4",
			Size:           PostSizeMedium,
			Status:         PostStatusCompleted,
			Uuid:           domain.GetUuid(),
		},
	}

	if err := CreatePosts(posts); err != nil {
		t.Errorf("could not create test post ... %v", err)
		t.FailNow()
	}
	return posts
}

func (ms *ModelSuite) TestPost_ValidateUpdate() {
	t := ms.T()

	posts := CreateFixturesValidateUpdate(ms, t)

	tests := []struct {
		name     string
		post     Post
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "good status - from open to open",
			post: Post{
				Status: PostStatusOpen,
				Uuid:   posts[0].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from open to committed",
			post: Post{
				Title:  "New Title",
				Status: PostStatusCommitted,
				Uuid:   posts[0].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from open to removed",
			post: Post{
				Status: PostStatusRemoved,
				Uuid:   posts[0].Uuid,
			},
			wantErr: false,
		},
		{
			name: "bad status - from open to accepted",
			post: Post{
				Status: PostStatusAccepted,
				Uuid:   posts[0].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from open to received",
			post: Post{
				Status: PostStatusReceived,
				Uuid:   posts[0].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from open to completed",
			post: Post{
				Status: PostStatusCompleted,
				Uuid:   posts[0].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "good status - from committed to committed",
			post: Post{
				Title:  "New Title",
				Status: PostStatusCommitted,
				Uuid:   posts[1].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from committed to open",
			post: Post{
				Status: PostStatusOpen,
				Uuid:   posts[1].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from committed to accepted",
			post: Post{
				Status: PostStatusAccepted,
				Uuid:   posts[1].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from committed to removed",
			post: Post{
				Status: PostStatusRemoved,
				Uuid:   posts[1].Uuid,
			},
			wantErr: false,
		},
		{
			name: "bad status - from committed to received",
			post: Post{
				Status: PostStatusReceived,
				Uuid:   posts[1].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from committed to completed",
			post: Post{
				Status: PostStatusCompleted,
				Uuid:   posts[1].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "good status - from accepted to accepted",
			post: Post{
				Title:  "New Title",
				Status: PostStatusAccepted,
				Uuid:   posts[2].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to open",
			post: Post{
				Status: PostStatusOpen,
				Uuid:   posts[2].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to received",
			post: Post{
				Status: PostStatusReceived,
				Uuid:   posts[2].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to removed",
			post: Post{
				Status: PostStatusRemoved,
				Uuid:   posts[2].Uuid,
			},
			wantErr: false,
		},
		{
			name: "bad status - from accepted to committed",
			post: Post{
				Status: PostStatusCommitted,
				Uuid:   posts[2].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from accepted to completed",
			post: Post{
				Status: PostStatusCompleted,
				Uuid:   posts[2].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "good status - from received to received",
			post: Post{
				Title:  "New Title",
				Status: PostStatusReceived,
				Uuid:   posts[3].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from received to completed",
			post: Post{
				Status: PostStatusCompleted,
				Uuid:   posts[3].Uuid,
			},
			wantErr: false,
		},
		{
			name: "bad status - from received to open",
			post: Post{
				Status: PostStatusOpen,
				Uuid:   posts[3].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from received to committed",
			post: Post{
				Status: PostStatusCommitted,
				Uuid:   posts[3].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from received to accepted",
			post: Post{
				Status: PostStatusAccepted,
				Uuid:   posts[3].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "good status - from completed to completed",
			post: Post{
				Title:  "New Title",
				Status: PostStatusCompleted,
				Uuid:   posts[4].Uuid,
			},
			wantErr: false,
		},
		{
			name: "bad status - from completed to open",
			post: Post{
				Status: PostStatusOpen,
				Uuid:   posts[4].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from completed to committed",
			post: Post{
				Status: PostStatusCommitted,
				Uuid:   posts[4].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from completed to accepted",
			post: Post{
				Status: PostStatusAccepted,
				Uuid:   posts[4].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from completed to received",
			post: Post{
				Status: PostStatusReceived,
				Uuid:   posts[4].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.post.ValidateUpdate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(test.errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", test.errField, vErr.Errors)
				}
			} else if (test.wantErr == false) && (vErr.HasAny()) {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func CreatePostFixtures(ms *ModelSuite, t *testing.T, users Users) []Post {
	if err := DB.Load(&users[0], "Organizations"); err != nil {
		t.Errorf("failed to load organizations on users[0] fixture, %s", err)
	}

	// Load Post test fixtures
	posts := []Post{
		{
			CreatedByID:    users[0].ID,
			Type:           PostTypeRequest,
			OrganizationID: users[0].Organizations[0].ID,
			Title:          "A Request",
			Size:           PostSizeMedium,
			Status:         PostStatusOpen,
			Uuid:           domain.GetUuid(),
			ProviderID:     nulls.NewInt(users[1].ID),
		},
		{
			CreatedByID:    users[0].ID,
			Type:           PostTypeOffer,
			OrganizationID: users[0].Organizations[0].ID,
			Title:          "An Offer",
			Size:           PostSizeMedium,
			Status:         PostStatusOpen,
			Uuid:           domain.GetUuid(),
			ReceiverID:     nulls.NewInt(users[1].ID),
		},
	}
	for i := range posts {
		if err := ms.DB.Create(&posts[i]); err != nil {
			t.Errorf("could not create test post ... %v", err)
			t.FailNow()
		}
		if err := DB.Load(&posts[i], "CreatedBy", "Provider", "Receiver", "Organization"); err != nil {
			t.Errorf("Error loading post associations: %s", err)
			t.FailNow()
		}
	}
	return posts
}

func (ms *ModelSuite) TestPost_FindByUUID() {
	t := ms.T()
	ResetTables(t, ms.DB)

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)

	tests := []struct {
		name    string
		uuid    string
		want    Post
		wantErr bool
	}{
		{name: "good", uuid: posts[0].Uuid.String(), want: posts[0]},
		{name: "blank uuid", uuid: "", wantErr: true},
		{name: "wrong uuid", uuid: domain.GetUuid().String(), wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var post Post
			err := post.FindByUUID(test.uuid)
			if test.wantErr {
				if (err != nil) != test.wantErr {
					t.Errorf("FindByUUID() did not return expected error")
				}
			} else {
				if err != nil {
					t.Errorf("FindByUUID() error = %v", err)
				} else if post.Uuid != test.want.Uuid {
					t.Errorf("FindByUUID() got = %s, want %s", post.Uuid, test.want.Uuid)
				}
			}
		})
	}
}

func (ms *ModelSuite) TestPost_GetCreator() {
	t := ms.T()
	ResetTables(t, ms.DB)

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)

	tests := []struct {
		name string
		post Post
		want uuid.UUID
	}{
		{name: "good", post: posts[0], want: posts[0].CreatedBy.Uuid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user, err := test.post.GetCreator([]string{"uuid"})
			if err != nil {
				t.Errorf("GetCreator() error = %v", err)
			} else if user.Uuid != test.want {
				t.Errorf("GetCreator() got = %s, want %s", user.Uuid, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_GetProvider() {
	t := ms.T()
	ResetTables(t, ms.DB)

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)

	tests := []struct {
		name string
		post Post
		want *uuid.UUID
	}{
		{name: "good", post: posts[0], want: &posts[0].Provider.Uuid},
		{name: "nil", post: posts[1], want: nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user, err := test.post.GetProvider([]string{"uuid"})
			if err != nil {
				t.Errorf("GetProvider() error = %v", err)
			} else if test.want == nil {
				if user != nil {
					t.Errorf("expected nil, got %s", user.Uuid.String())
				}
			} else if user == nil {
				t.Errorf("received nil, expected %v", test.want.String())
			} else if user.Uuid != *test.want {
				t.Errorf("GetProvider() got = %s, want %s", user.Uuid, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_GetReceiver() {
	t := ms.T()
	ResetTables(t, ms.DB)

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)

	tests := []struct {
		name string
		post Post
		want *uuid.UUID
	}{
		{name: "good", post: posts[1], want: &posts[1].Receiver.Uuid},
		{name: "nil", post: posts[0], want: nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user, err := test.post.GetReceiver([]string{"uuid"})
			if err != nil {
				t.Errorf("GetReceiver() error = %v", err)
			} else if test.want == nil {
				if user != nil {
					t.Errorf("expected nil, got %s", user.Uuid.String())
				}
			} else if user == nil {
				t.Errorf("received nil, expected %v", test.want.String())
			} else if user.Uuid != *test.want {
				t.Errorf("GetProvider() got = %s, want %s", user.Uuid, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_GetOrganization() {
	t := ms.T()
	ResetTables(t, ms.DB)

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)

	tests := []struct {
		name string
		post Post
		want uuid.UUID
	}{
		{name: "good", post: posts[0], want: posts[0].Organization.Uuid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			org, err := test.post.GetOrganization([]string{"uuid"})
			if err != nil {
				t.Errorf("GetOrganization() error = %v", err)
			} else if org.Uuid != test.want {
				t.Errorf("GetOrganization() got = %s, want %s", org.Uuid, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_GetThreads() {
	t := ms.T()
	ResetTables(t, ms.DB)

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	threadFixtures := CreateThreadFixtures(ms, t, posts[0])
	threads := threadFixtures.Threads

	tests := []struct {
		name string
		post Post
		want []uuid.UUID
	}{
		{name: "no threads", post: posts[1], want: []uuid.UUID{}},
		{name: "two threads", post: posts[0], want: []uuid.UUID{threads[0].Uuid, threads[1].Uuid}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.post.GetThreads([]string{"uuid"}, users[0])
			if err != nil {
				t.Errorf("GetThreads() error: %v", err)
			} else {
				ids := make([]uuid.UUID, len(got))
				for i := range got {
					ids[i] = got[i].Uuid
				}
				if !reflect.DeepEqual(ids, test.want) {
					t.Errorf("GetThreads() got = %s, want %s", ids, test.want)
				}
			}
		})
	}
}

func (ms *ModelSuite) TestPost_GetThreadIdForUser() {
	t := ms.T()
	ResetTables(t, ms.DB)

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	threadFixtures := CreateThreadFixtures(ms, t, posts[0])
	thread0UUID := threadFixtures.Threads[0].Uuid.String()

	tests := []struct {
		name string
		post Post
		user User
		want *string
	}{
		{name: "no threads", post: posts[1], user: users[0], want: nil},
		{name: "good", post: posts[0], user: users[0], want: &thread0UUID},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.post.GetThreadIdForUser(test.user)
			if err != nil {
				t.Errorf("GetThreadIdForUser() error: %v", err)
			} else if test.want == nil {
				if got != nil {
					t.Errorf("GetThreadIdForUser() returned %v, expected nil", *got)
				}
			} else if got == nil {
				t.Errorf("GetThreadIdForUser() returned nil, expected %v", *test.want)
			} else if *test.want != *got {
				t.Errorf("GetThreadIdForUser() got = %s, want %s", *got, *test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_AttachFile() {
	t := ms.T()
	ResetTables(t, ms.DB)

	user := User{}
	if err := ms.DB.Create(&user); err != nil {
		t.Errorf("failed to create user fixture, %s", err)
	}

	organization := Organization{AuthConfig: "{}"}
	if err := ms.DB.Create(&organization); err != nil {
		t.Errorf("failed to create organization fixture, %s", err)
	}

	post := Post{
		CreatedByID:    user.ID,
		OrganizationID: organization.ID,
	}
	if err := ms.DB.Create(&post); err != nil {
		t.Errorf("failed to create post fixture, %s", err)
	}

	var fileFixture File
	const filename = "photo.gif"
	if err := fileFixture.Store(filename, []byte("GIF89a")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
	}

	if attachedFile, err := post.AttachFile(fileFixture.UUID.String()); err != nil {
		t.Errorf("failed to attach file to post, %s", err)
	} else {
		ms.Equal(filename, attachedFile.Name)
		ms.NotEqual(0, attachedFile.ID)
		ms.NotEqual(domain.EmptyUUID, attachedFile.UUID.String())
	}

	if err := DB.Load(&post); err != nil {
		t.Errorf("failed to load relations for test post, %s", err)
	}

	ms.Equal(1, len(post.Files))

	if err := DB.Load(&(post.Files[0])); err != nil {
		t.Errorf("failed to load files relations for test post, %s", err)
	}

	ms.Equal(filename, post.Files[0].File.Name)
}

func (ms *ModelSuite) TestPost_GetFiles() {
	t := ms.T()
	ResetTables(t, ms.DB)

	user := User{}
	if err := ms.DB.Create(&user); err != nil {
		t.Errorf("failed to create user fixture, %s", err)
	}

	organization := Organization{AuthConfig: "{}"}
	if err := ms.DB.Create(&organization); err != nil {
		t.Errorf("failed to create organization fixture, %s", err)
	}

	post := Post{
		CreatedByID:    user.ID,
		OrganizationID: organization.ID,
	}
	if err := ms.DB.Create(&post); err != nil {
		t.Errorf("failed to create post fixture, %s", err)
	}

	var f File
	const filename = "photo.gif"
	if err := f.Store(filename, []byte("GIF89a")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
	}

	if _, err := post.AttachFile(f.UUID.String()); err != nil {
		t.Errorf("failed to attach file to post, %s", err)
	}

	files, err := post.GetFiles()
	if err != nil {
		t.Errorf("failed to get files list for post, %s", err)
	}

	ms.Equal(1, len(files))
	ms.Equal(filename, files[0].Name)
}

// TestPost_AttachPhoto_GetPhoto tests the AttachPhoto and GetPhoto methods of models.Post
func (ms *ModelSuite) TestPost_AttachPhoto_GetPhoto() {
	t := ms.T()
	ResetTables(t, ms.DB)

	user := User{}
	if err := ms.DB.Create(&user); err != nil {
		t.Errorf("failed to create user fixture, %s", err)
	}

	organization := Organization{AuthConfig: "{}"}
	if err := ms.DB.Create(&organization); err != nil {
		t.Errorf("failed to create organization fixture, %s", err)
	}

	post := Post{
		CreatedByID:    user.ID,
		OrganizationID: organization.ID,
	}
	if err := ms.DB.Create(&post); err != nil {
		t.Errorf("failed to create post fixture, %s", err)
	}

	var photoFixture File
	const filename = "photo.gif"
	if err := photoFixture.Store(filename, []byte("GIF89a")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
	}

	attachedFile, err := post.AttachPhoto(photoFixture.UUID.String())
	if err != nil {
		t.Errorf("failed to attach photo to post, %s", err)
	} else {
		ms.Equal(filename, attachedFile.Name)
		ms.NotEqual(0, attachedFile.ID)
		ms.NotEqual(domain.EmptyUUID, attachedFile.UUID.String())
	}

	if err := DB.Load(&post); err != nil {
		t.Errorf("failed to load photo relation for test post, %s", err)
	}

	ms.Equal(filename, post.PhotoFile.Name)

	if got, err := post.GetPhoto(); err == nil {
		ms.Equal(attachedFile.UUID.String(), got.UUID.String())
		ms.True(got.URLExpiration.After(time.Now().Add(time.Minute)))
		ms.Equal(filename, got.Name)
	} else {
		ms.Fail("post.GetPhoto failed, %s", err)
	}
}
