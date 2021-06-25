package actions

import (
	"fmt"
	"time"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type watchQueryFixtures struct {
	models.Users
	models.Locations
	models.Watches
	models.Meetings
}

func createFixturesForWatches(as *ActionSuite) watchQueryFixtures {
	// make 2 users, 1 that has Watches, and another that will try to mess with those Watches
	uf := test.CreateUserFixtures(as.DB, 2)
	locations := test.CreateLocationFixtures(as.DB, 3)
	watches := make(models.Watches, 2)
	for i := range watches {
		watches[i].OwnerID = uf.Users[0].ID
		watches[i].DestinationID = nulls.NewInt(locations[i].ID)
		test.MustCreate(as.DB, &watches[i])
	}
	meetings := models.Meetings{
		{
			CreatedByID: uf.Users[0].ID,
			Name:        "Mtg",
			LocationID:  locations[2].ID,

			StartDate: time.Now().Add(domain.DurationWeek * 8),
			EndDate:   time.Now().Add(domain.DurationWeek * 10),
		},
	}

	for i := range meetings {
		meetings[i].UUID = domain.GetUUID()
		createFixture(as, &meetings[i])
	}

	return watchQueryFixtures{
		Users:     uf.Users,
		Locations: locations,
		Watches:   watches,
		Meetings:  meetings,
	}
}

func (as *ActionSuite) Test_MyWatches() {
	f := createFixturesForWatches(as)
	watches := f.Watches

	owner := f.Users[0]
	destinations := f.Locations

	req := as.JSON("/watches")
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", owner.Nickname)
	req.Headers["content-type"] = "application/json"
	res := req.Get()

	body := res.Body.String()
	as.Equal(200, res.Code, "incorrect status code returned, body: %s", body)

	wantContains := []string{
		fmt.Sprintf(`"owner":{"id":"%s"`, owner.UUID),
		fmt.Sprintf(`"nickname":"%s"`, owner.Nickname),
		fmt.Sprintf(`"avatar_url":"%s"`, owner.AuthPhotoURL.String),
		fmt.Sprintf(`"id":"%s"`, watches[0].UUID.String()),
		fmt.Sprintf(`"id":"%s"`, watches[1].UUID.String()),
		fmt.Sprintf(`"destination":{"description":"%s"`, destinations[0].Description),
		fmt.Sprintf(`"destination":{"description":"%s"`, destinations[1].Description),
	}
	for _, w := range wantContains {
		as.Contains(body, w)
	}

	as.NotContains(body, `"origin":`)

	// Try with no watches
	nonOwner := f.Users[1]
	req = as.JSON("/watches")
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", nonOwner.Nickname)
	req.Headers["content-type"] = "application/json"
	res = req.Get()

	body = res.Body.String()
	as.Equal(200, res.Code, "incorrect status code returned, body: %s", body)
	as.Equal("[]\n", body, "expected an empty list in the response")
}
