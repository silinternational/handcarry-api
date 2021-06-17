package gqlgen

import (
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type publicProfileFixtures struct {
	models.User
}

func createFixturesForGetPublicProfile() publicProfileFixtures {
	unique := domain.GetUUID().String()
	user := models.User{
		UUID:         domain.GetUUID(),
		Nickname:     "user0" + unique,
		AuthPhotoURL: nulls.NewString("https://example.com/userphoto/1"),
	}
	return publicProfileFixtures{user}
}

func (gs *GqlgenSuite) Test_getPublicProfile() {
	t := gs.T()
	f := createFixturesForGetPublicProfile()
	tests := []struct {
		name string
		user *models.User
		want *PublicProfile
	}{
		{
			name: "fully-specified User",
			user: &f.User,
			want: &PublicProfile{
				ID:        f.User.UUID.String(),
				Nickname:  f.User.Nickname,
				AvatarURL: &f.User.AuthPhotoURL.String,
			},
		},
		{
			name: "nil user",
			user: nil,
			want: &PublicProfile{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := getPublicProfile(test.Ctx(), tt.user)

			gs.NotNil(profile)
			gs.Equal(tt.want, profile, "incorrect profile")
		})
	}
}
