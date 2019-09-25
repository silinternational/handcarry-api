package models

import (
	"testing"
	"time"
)

func (ms *ModelSuite) TestUserAccessToken_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		token    UserAccessToken
		wantErr  bool
		errField string
	}{
		{
			name:    "minimum",
			token:   UserAccessToken{UserID: 1, AccessToken: "abc123", ExpiresAt: time.Now()},
			wantErr: false,
		},
		{
			name:     "missing user_id",
			token:    UserAccessToken{AccessToken: "abc123", ExpiresAt: time.Now()},
			wantErr:  true,
			errField: "user_id",
		},
		{
			name:     "missing access_token",
			token:    UserAccessToken{UserID: 1, ExpiresAt: time.Now()},
			wantErr:  true,
			errField: "access_token",
		},
		{
			name:     "missing expires_at",
			token:    UserAccessToken{UserID: 1, AccessToken: "abc123"},
			wantErr:  true,
			errField: "expires_at",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.token.Validate(DB)
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

func (ms *ModelSuite) TestUserAccessToken_DeleteByBearerToken() {
	t := ms.T()
	ResetTables(t, ms.DB)

	_, users, userOrgs := CreateUserFixtures(ms, t)
	tokens := CreateUserAccessTokenFixtures(t, users[0], userOrgs)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{name: "success1", token: tokens[0], wantErr: false},
		{name: "success2", token: tokens[1], wantErr: false},
		{name: "failure", token: "------", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var uat UserAccessToken
			err := uat.DeleteByBearerToken(test.token)
			if err != nil && !test.wantErr {
				t.Errorf("DeleteAccessToken() returned an unexpected error: %s", err)
			} else if err == nil && test.wantErr {
				t.Errorf("expected an error, but DeleteAccessToken() did not return an error")
			}
		})
	}
}

func (ms *ModelSuite) TestUserAccessToken_FindByBearerToken() {
	t := ms.T()
	ResetTables(t, ms.DB)

	_, users, userOrgs := CreateUserFixtures(ms, t)
	tokens := CreateUserAccessTokenFixtures(t, users[0], userOrgs)

	tests := []struct {
		name    string
		token   string
		want    User
		wantErr bool
	}{
		{name: "valid0", token: tokens[0], want: users[0]},
		{name: "valid1", token: tokens[1], want: users[0]},
		{name: "invalid", token: "000000", wantErr: true},
		{name: "empty", token: "", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var u UserAccessToken
			err := u.FindByBearerToken(test.token)
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected an error, but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("FindByAccessToken() returned an error: %v", err)
				} else if u.User.Uuid != test.want.Uuid {
					t.Errorf("found %v, expected %v", u, test.want)
				}
			}
		})
	}
}

func CreateUserAccessTokenFixtures(t *testing.T, user User, userOrgs UserOrganizations) []string {
	rawTokens := []string{"abc123", "xyz789"}
	// Load access token test fixtures
	tokens := UserAccessTokens{
		{
			UserID:             user.ID,
			UserOrganizationID: userOrgs[0].ID,
			AccessToken:        HashClientIdAccessToken(rawTokens[0]),
			ExpiresAt:          time.Unix(0, 0),
		},
		{
			UserID:             user.ID,
			UserOrganizationID: userOrgs[0].ID,
			AccessToken:        HashClientIdAccessToken(rawTokens[1]),
			ExpiresAt:          time.Date(2099, time.December, 31, 0, 0, 0, 0, time.UTC),
		},
	}

	if err := CreateUserAccessTokens(tokens); err != nil {
		t.Errorf("could not create access tokens ... %v", err)
		t.FailNow()
	}

	return rawTokens
}
