package actions

import (
	"time"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

//  meeting       creator    invitees            participants          organizers
//  0 Mtg Past    user 0
//  1 Mtg Recent  user 0     invitee2            user1
//  2 Mtg Now     user 0     invitee0, invitee1  user0, user1, user2   user1
//  3 Mtg Future  user 0
//
//  Inviter for all invites is user 0
func createFixturesForMeetings(as *ActionSuite) meetingQueryFixtures {
	uf := test.CreateUserFixtures(as.DB, 3)
	user := uf.Users[0]
	locations := test.CreateLocationFixtures(as.DB, 4)

	err := aws.CreateS3Bucket()
	as.NoError(err, "failed to create S3 bucket, %s", err)

	var fileFixture models.File
	fErr := fileFixture.Store("new_photo.webp", []byte("RIFFxxxxWEBPVP"))
	as.Nil(fErr, "failed to create ImageFile fixture")

	meetings := models.Meetings{
		{
			CreatedByID: user.ID,
			Name:        "Mtg Past",
			LocationID:  locations[0].ID,

			StartDate: time.Now().Add(time.Duration(-domain.DurationWeek * 10)),
			EndDate:   time.Now().Add(time.Duration(-domain.DurationWeek * 8)),
		},
		{
			CreatedByID: user.ID,
			Name:        "Mtg Recent",
			LocationID:  locations[1].ID,

			StartDate: time.Now().Add(time.Duration(-domain.DurationWeek * 4)),
			EndDate:   time.Now().Add(time.Duration(-domain.DurationWeek * 2)),
		},
		{
			CreatedByID: user.ID,
			Name:        "Mtg Now",
			LocationID:  locations[2].ID,
			StartDate:   time.Now().Add(time.Duration(-domain.DurationWeek * 2)),
			EndDate:     time.Now().Add(time.Duration(domain.DurationWeek * 2)),
			ImageFileID: nulls.NewInt(fileFixture.ID),
		},
		{
			CreatedByID: user.ID,
			Name:        "Mtg Future",
			LocationID:  locations[3].ID,
			StartDate:   time.Now().Add(time.Duration(domain.DurationWeek * 2)),
			EndDate:     time.Now().Add(time.Duration(domain.DurationWeek * 4)),
		},
	}

	for i := range meetings {
		meetings[i].UUID = domain.GetUUID()
		createFixture(as, &meetings[i])
	}

	posts := test.CreatePostFixtures(as.DB, 3, false)
	posts[0].MeetingID = nulls.NewInt(meetings[2].ID)
	posts[1].MeetingID = nulls.NewInt(meetings[2].ID)
	as.NoError(as.DB.Update(&posts))

	invites := models.MeetingInvites{
		{
			MeetingID: meetings[2].ID,
			InviterID: user.ID,
			Email:     "invitee0@example.com",
		},
		{
			MeetingID: meetings[2].ID,
			InviterID: user.ID,
			Email:     "invitee1@example.com",
		},
		{
			MeetingID: meetings[1].ID,
			InviterID: user.ID,
			Email:     "invitee2@example.com",
		},
	}
	for i := range invites {
		as.NoError(invites[i].Create())
	}

	participants := models.MeetingParticipants{
		{
			MeetingID:   meetings[2].ID,
			UserID:      uf.Users[0].ID,
			InviteID:    nulls.NewInt(invites[0].ID),
			IsOrganizer: false,
		},
		{
			MeetingID:   meetings[2].ID,
			UserID:      uf.Users[1].ID,
			InviteID:    nulls.NewInt(invites[1].ID),
			IsOrganizer: true,
		},
		{
			MeetingID:   meetings[1].ID,
			UserID:      uf.Users[1].ID,
			InviteID:    nulls.NewInt(invites[2].ID),
			IsOrganizer: false,
		},
		{
			MeetingID:   meetings[2].ID,
			UserID:      uf.Users[2].ID,
			IsOrganizer: false,
		},
	}
	createFixture(as, &participants)

	return meetingQueryFixtures{
		Locations:           locations,
		Meetings:            meetings,
		Users:               uf.Users,
		File:                fileFixture,
		Posts:               posts,
		MeetingInvites:      invites,
		MeetingParticipants: participants,
	}
}
