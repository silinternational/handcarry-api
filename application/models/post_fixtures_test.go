package models

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
)

type PostFixtures struct {
	Users
	Posts
	PostHistories
	Files
	Locations
	PotentialProviders
}

func CreateFixturesValidateUpdate_RequestStatus(status PostStatus, ms *ModelSuite, t *testing.T) Post {
	uf := createUserFixtures(ms.DB, 1)
	org := uf.Organization
	user := uf.Users[0]

	location := Location{}
	createFixture(ms, &location)

	post := Post{
		CreatedByID:    user.ID,
		OrganizationID: org.ID,
		DestinationID:  location.ID,
		Type:           PostTypeRequest,
		Title:          "Test Request",
		Size:           PostSizeMedium,
		UUID:           domain.GetUUID(),
		Status:         status,
	}

	createFixture(ms, &post)

	return post
}

func createFixturesForTestPostCreate(ms *ModelSuite) PostFixtures {
	uf := createUserFixtures(ms.DB, 1)
	org := uf.Organization
	user := uf.Users[0]

	posts := Posts{
		{UUID: domain.GetUUID(), Title: "title0"},
		{Title: "title1"},
		{},
	}
	locations := make(Locations, len(posts))
	for i := range posts {
		locations[i].Description = "location " + strconv.Itoa(i)
		createFixture(ms, &locations[i])

		posts[i].Status = PostStatusOpen
		posts[i].Type = PostTypeRequest
		posts[i].Size = PostSizeTiny
		posts[i].CreatedByID = user.ID
		posts[i].OrganizationID = org.ID
		posts[i].DestinationID = locations[i].ID
	}
	createFixture(ms, &posts[2])

	return PostFixtures{
		Users: Users{user},
		Posts: posts,
	}
}

func createFixturesForTestPostUpdate(ms *ModelSuite) PostFixtures {
	posts := createPostFixtures(ms.DB, 2, 0, false)
	posts[0].Title = "new title"
	posts[1].Title = ""

	return PostFixtures{
		Posts: posts,
	}
}

func createFixturesForTestPost_manageStatusTransition_forwardProgression(ms *ModelSuite) PostFixtures {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	posts := createPostFixtures(ms.DB, 4, 0, false)
	posts[1].Status = PostStatusAccepted
	posts[1].CreatedByID = users[1].ID
	posts[1].ProviderID = nulls.NewInt(users[0].ID)
	ms.NoError(ms.DB.Save(&posts[1]))

	// Give the last Post an intermediate PostHistory
	pHistory := PostHistory{
		Status: PostStatusAccepted,
		PostID: posts[3].ID,
	}
	mustCreate(ms.DB, &pHistory)

	// Give these new statuses while by-passing the status transition validation
	for i, status := range [2]PostStatus{PostStatusAccepted, PostStatusDelivered} {
		index := i + 2

		posts[index].Status = status
		ms.NoError(ms.DB.Save(&posts[index]))
	}

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}

func createFixturesForTestPost_manageStatusTransition_backwardProgression(ms *ModelSuite) PostFixtures {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	posts := createPostFixtures(ms.DB, 4, 0, false)

	// Put the first two posts into ACCEPTED status (also give them matching PostHistory entries)
	posts[0].Status = PostStatusAccepted
	posts[0].CreatedByID = users[0].ID
	posts[0].ProviderID = nulls.NewInt(users[1].ID)
	posts[1].Status = PostStatusAccepted
	posts[1].CreatedByID = users[1].ID
	posts[1].ProviderID = nulls.NewInt(users[0].ID)
	ms.NoError(ms.DB.Save(&posts))

	// add in a PostHistory entry as if it had already happened
	pHistory := PostHistory{
		Status:     PostStatusAccepted,
		PostID:     posts[2].ID,
		ReceiverID: nulls.NewInt(users[0].ID),
		ProviderID: nulls.NewInt(users[1].ID),
	}
	mustCreate(ms.DB, &pHistory)

	pHistory = PostHistory{
		Status:     PostStatusDelivered,
		PostID:     posts[3].ID,
		ReceiverID: nulls.NewInt(users[0].ID),
		ProviderID: nulls.NewInt(users[1].ID),
	}
	mustCreate(ms.DB, &pHistory)

	for i := 2; i < 4; i++ {
		// AfterUpdate creates the new PostHistory for this status
		posts[i].Status = PostStatusCompleted
		posts[i].CompletedOn = nulls.NewTime(time.Now())
		posts[i].ProviderID = nulls.NewInt(users[1].ID)
		ms.NoError(ms.DB.Save(&posts[i]))
	}

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}

func CreateFixturesForPostsGetFiles(ms *ModelSuite) PostFixtures {
	uf := createUserFixtures(ms.DB, 1)
	organization := uf.Organization
	user := uf.Users[0]

	location := Location{}
	createFixture(ms, &location)

	post := Post{CreatedByID: user.ID, OrganizationID: organization.ID, DestinationID: location.ID}
	createFixture(ms, &post)

	files := make(Files, 3)

	for i := range files {
		var file File
		ms.Nil(file.Store(fmt.Sprintf("file_%d.gif", i), []byte("GIF87a")),
			"failed to create file fixture")
		files[i] = file
		_, err := post.AttachFile(files[i].UUID.String())
		ms.NoError(err, "failed to attach file to post fixture")
	}

	return PostFixtures{
		Users: Users{user},
		Posts: Posts{post},
		Files: files,
	}
}

func createFixturesForPostFindByUserAndUUID(ms *ModelSuite) PostFixtures {
	orgs := Organizations{{}, {}}
	for i := range orgs {
		orgs[i].UUID = domain.GetUUID()
		orgs[i].AuthConfig = "{}"
		createFixture(ms, &orgs[i])
	}

	users := createUserFixtures(ms.DB, 2).Users

	// both users are in org 0, but need user 0 to also be in org 1
	createFixture(ms, &UserOrganization{
		OrganizationID: orgs[1].ID,
		UserID:         users[0].ID,
		AuthID:         users[0].Email,
		AuthEmail:      users[0].Email,
	})

	posts := createPostFixtures(ms.DB, 3, 0, false)
	posts[1].OrganizationID = orgs[1].ID
	posts[2].Status = PostStatusRemoved
	ms.NoError(ms.DB.Save(&posts))

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}

//        Org0                Org1           Org2
//        |  |                | | |          | |
//        |  +----+-----------+ | +----+-----+ +
//        |       |             |      |       |
//       User1  User0        User3   Trust   User2
//
// Org0: Post0 (SAME), Post2 (SAME, COMPLETED), Post3 (SAME, REMOVED), Post4 (SAME)
// Org1: Post1 (SAME)
// Org2: Post5 (ALL), Post6 (TRUSTED), Post7 (SAME)
//
func CreateFixtures_Posts_FindByUser(ms *ModelSuite) PostFixtures {
	orgs := createOrganizationFixtures(ms.DB, 3)

	trusts := OrganizationTrusts{
		{PrimaryID: orgs[1].ID, SecondaryID: orgs[2].ID},
		{PrimaryID: orgs[2].ID, SecondaryID: orgs[1].ID},
	}
	createFixture(ms, &trusts)

	users := createUserFixtures(ms.DB, 4).Users

	createFixture(ms, &UserOrganization{
		OrganizationID: orgs[1].ID,
		UserID:         users[0].ID,
		AuthID:         users[0].Email,
		AuthEmail:      users[0].Email,
	})

	uo, err := users[2].FindUserOrganization(orgs[0])
	ms.NoError(err)
	uo.OrganizationID = orgs[2].ID
	ms.NoError(DB.UpdateColumns(&uo, "organization_id"))

	uo, err = users[3].FindUserOrganization(orgs[0])
	ms.NoError(err)
	uo.OrganizationID = orgs[1].ID
	ms.NoError(DB.UpdateColumns(&uo, "organization_id"))

	posts := createPostFixtures(ms.DB, 8, 0, false)
	posts[1].OrganizationID = orgs[1].ID
	posts[2].Status = PostStatusOpen
	posts[3].Status = PostStatusRemoved
	posts[4].CreatedByID = users[1].ID
	posts[5].OrganizationID = orgs[2].ID
	posts[5].Visibility = PostVisibilityAll
	posts[6].OrganizationID = orgs[2].ID
	posts[6].Visibility = PostVisibilityTrusted
	posts[7].OrganizationID = orgs[2].ID
	ms.NoError(ms.DB.Save(&posts))

	// can't go directly to "completed"
	posts[2].Status = PostStatusAccepted
	ms.NoError(ms.DB.Save(&posts[2]))
	posts[2].Status = PostStatusCompleted
	ms.NoError(ms.DB.Save(&posts[2]))

	ms.NoError(posts[0].SetDestination(Location{Description: "Australia", Country: "AU"}))
	ms.NoError(posts[1].SetOrigin(Location{Description: "Australia", Country: "AU"}))

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}

func createFixturesFor_Posts_GetPotentialProviders(ms *ModelSuite) PostFixtures {
	posts := createPostFixtures(ms.DB, 2, 0, false)
	pps := createPotentialProviderFixtures(ms.DB, 2, 2)

	return PostFixtures{
		Posts:              posts,
		PotentialProviders: pps,
	}
}

func createFixtures_Posts_FindByUser_SearchText(ms *ModelSuite) PostFixtures {
	orgs := Organizations{{}, {}}
	for i := range orgs {
		orgs[i].UUID = domain.GetUUID()
		orgs[i].AuthConfig = "{}"
		createFixture(ms, &orgs[i])
	}

	unique := domain.GetUUID().String()
	users := Users{
		{Email: unique + "_user0@example.com", Nickname: unique + "User0"},
		{Email: unique + "_user1@example.com", Nickname: unique + "User1"},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
		createFixture(ms, &users[i])
	}

	userOrgs := UserOrganizations{
		{OrganizationID: orgs[0].ID, UserID: users[0].ID, AuthID: users[0].Email, AuthEmail: users[0].Email},
		{OrganizationID: orgs[1].ID, UserID: users[0].ID, AuthID: users[0].Email, AuthEmail: users[0].Email},
		{OrganizationID: orgs[1].ID, UserID: users[1].ID, AuthID: users[1].Email, AuthEmail: users[1].Email},
	}
	for i := range userOrgs {
		createFixture(ms, &userOrgs[i])
	}

	locations := make([]Location, 6)
	for i := range locations {
		createFixture(ms, &locations[i])
	}

	posts := Posts{
		{CreatedByID: users[0].ID, OrganizationID: orgs[0].ID, Title: "With Match"},
		{CreatedByID: users[0].ID, OrganizationID: orgs[1].ID, Title: "MXtch In Description",
			Description: nulls.NewString("This has the lower case match in it.")},
		{CreatedByID: users[0].ID, OrganizationID: orgs[0].ID, Status: PostStatusCompleted,
			Title: "With Match But Completed"},
		{CreatedByID: users[0].ID, OrganizationID: orgs[0].ID, Status: PostStatusRemoved,
			Title: "With Match But Removed"},
		{CreatedByID: users[1].ID, OrganizationID: orgs[1].ID, Title: "User1 No MXtch"},
		{CreatedByID: users[1].ID, OrganizationID: orgs[1].ID, Title: "User1 With MATCH"},
	}

	for i := range posts {
		posts[i].UUID = domain.GetUUID()
		posts[i].DestinationID = locations[i].ID
		posts[i].Type = PostTypeRequest
		createFixture(ms, &posts[i])
	}

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}

func CreateFixtures_Post_IsEditable(ms *ModelSuite) PostFixtures {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	posts := createPostFixtures(ms.DB, 2, 0, false)
	posts[1].Status = PostStatusRemoved

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}

func createFixturesForPostGetAudience(ms *ModelSuite) PostFixtures {
	orgs := make(Organizations, 2)
	for i := range orgs {
		orgs[i] = Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
		createFixture(ms, &orgs[i])
	}

	users := createUserFixtures(ms.DB, 2).Users

	posts := createPostFixtures(ms.DB, 2, 0, false)
	posts[1].OrganizationID = orgs[1].ID
	ms.NoError(ms.DB.Save(&posts[1]))

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}

func createFixturesForGetLocationForNotifications(ms *ModelSuite) PostFixtures {
	uf := createUserFixtures(ms.DB, 1)
	users := uf.Users

	posts := createPostFixtures(ms.DB, 2, 1, false)
	posts[0].OriginID = nulls.Int{}
	ms.NoError(ms.DB.Save(&posts[0]))

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}
