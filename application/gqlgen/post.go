package gqlgen

import (
	"context"
	"errors"
	"fmt"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// Post returns the post resolver. It is required by GraphQL
func (r *Resolver) Post() PostResolver {
	return &postResolver{r}
}

type postResolver struct{ *Resolver }

// ID resolves the `ID` property of the post query. It provides the UUID instead of the autoincrement ID.
func (r *postResolver) ID(ctx context.Context, obj *models.Post) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.UUID.String(), nil
}

// CreatedBy resolves the `createdBy` property of the post query. It retrieves the related record from the database.
func (r *postResolver) CreatedBy(ctx context.Context, obj *models.Post) (*PublicProfile, error) {
	if obj == nil {
		return nil, nil
	}

	creator, err := obj.GetCreator()
	if err != nil {
		return nil, reportError(ctx, err, "GetPostCreator")
	}

	return getPublicProfile(ctx, creator), nil
}

// Receiver resolves the `receiver` property of the post query. It retrieves the related record from the database.
func (r *postResolver) Receiver(ctx context.Context, obj *models.Post) (*PublicProfile, error) {
	if obj == nil {
		return nil, nil
	}

	receiver, err := obj.GetReceiver()
	if err != nil {
		return nil, reportError(ctx, err, "GetPostReceiver")
	}

	return getPublicProfile(ctx, receiver), nil
}

// Provider resolves the `provider` property of the post query. It retrieves the related record from the database.
func (r *postResolver) Provider(ctx context.Context, obj *models.Post) (*PublicProfile, error) {
	if obj == nil {
		return nil, nil
	}

	provider, err := obj.GetProvider()
	if err != nil {
		return nil, reportError(ctx, err, "GetPostProvider")
	}

	return getPublicProfile(ctx, provider), nil
}

// Organization resolves the `organization` property of the post query. It retrieves the related record from the
// database.
func (r *postResolver) Organization(ctx context.Context, obj *models.Post) (*models.Organization, error) {
	if obj == nil {
		return nil, nil
	}

	organization, err := obj.GetOrganization()
	if err != nil {
		return nil, reportError(ctx, err, "GetPostOrganization")
	}

	return organization, nil
}

// Description resolves the `description` property, converting a nulls.String to a *string.
func (r *postResolver) Description(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return models.GetStringFromNullsString(obj.Description), nil
}

// NeededBefore resolves the `neededBefore` property of the post query, converting a nulls.Time to a *time.Time.
func (r *postResolver) NeededBefore(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}

	return models.GetStringFromNullsTime(obj.NeededBefore), nil
}

// Destination resolves the `destination` property of the post query, retrieving the related record from the database.
func (r *postResolver) Destination(ctx context.Context, obj *models.Post) (*models.Location, error) {
	if obj == nil {
		return nil, nil
	}

	destination, err := obj.GetDestination()
	if err != nil {
		return nil, reportError(ctx, err, "GetPostDestination")
	}

	return destination, nil
}

// Origin resolves the `origin` property of the post query, retrieving the related record from the database.
func (r *postResolver) Origin(ctx context.Context, obj *models.Post) (*models.Location, error) {
	if obj == nil {
		return nil, nil
	}

	origin, err := obj.GetOrigin()
	if err != nil {
		return nil, reportError(ctx, err, "GetPostOrigin")
	}

	return origin, nil
}

// Threads resolves the `threads` property of the post query, retrieving the related records from the database.
func (r *postResolver) Threads(ctx context.Context, obj *models.Post) ([]models.Thread, error) {
	if obj == nil {
		return nil, nil
	}

	user := models.GetCurrentUserFromGqlContext(ctx)
	threads, err := obj.GetThreads(user)
	if err != nil {
		extras := map[string]interface{}{
			"user": user.UUID,
		}
		return nil, reportError(ctx, err, "GetPostThreads", extras)
	}

	return threads, nil
}

// URL resolves the `url` property of the post query, converting nulls.String to a *string
func (r *postResolver) URL(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return models.GetStringFromNullsString(obj.URL), nil
}

// Kilograms resolves the `kilograms` property of the post query as a pointer to a float64
func (r *postResolver) Kilograms(ctx context.Context, obj *models.Post) (*float64, error) {
	if obj == nil {
		return nil, nil
	}
	if !obj.Kilograms.Valid {
		return nil, nil
	}

	return &obj.Kilograms.Float64, nil
}

// Photo retrieves the file attached as the primary photo
func (r *postResolver) Photo(ctx context.Context, obj *models.Post) (*models.File, error) {
	if obj == nil {
		return nil, nil
	}

	photo, err := obj.GetPhoto()
	if err != nil {
		return nil, reportError(ctx, err, "GetPostPhoto")
	}

	return photo, nil
}

// Files retrieves the list of files attached to the post, not including the primary photo
func (r *postResolver) Files(ctx context.Context, obj *models.Post) ([]models.File, error) {
	if obj == nil {
		return nil, nil
	}
	files, err := obj.GetFiles()
	if err != nil {
		return nil, reportError(ctx, err, "GetPostFiles")
	}

	return files, nil
}

// Meeting resolves the `meeting` property of the post query, retrieving the related record from the database.
func (r *postResolver) Meeting(ctx context.Context, obj *models.Post) (*models.Meeting, error) {
	if obj == nil {
		return nil, nil
	}

	meeting, err := obj.Meeting()
	if err != nil {
		return nil, reportError(ctx, err, "GetPostMeeting")
	}

	return meeting, nil
}

// IsEditable indicates whether the user is allowed to edit the post
func (r *postResolver) IsEditable(ctx context.Context, obj *models.Post) (bool, error) {
	if obj == nil {
		return false, nil
	}
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	return obj.IsEditable(cUser)
}

// Posts resolves the `posts` query
func (r *queryResolver) Posts(ctx context.Context) ([]models.Post, error) {
	posts := models.Posts{}
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	if err := posts.FindByUser(ctx, cUser); err != nil {
		extras := map[string]interface{}{
			"user": cUser.UUID,
		}
		return nil, reportError(ctx, err, "GetPosts", extras)
	}

	return posts, nil
}

// Post resolves the `post` query
func (r *queryResolver) Post(ctx context.Context, id *string) (*models.Post, error) {
	if id == nil {
		return nil, nil
	}
	var post models.Post
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	if err := post.FindByUserAndUUID(ctx, cUser, *id); err != nil {
		extras := map[string]interface{}{
			"user": cUser.UUID,
		}
		return nil, reportError(ctx, err, "GetPost", extras)
	}

	return &post, nil
}

// convertGqlPostInputToDBPost takes a `PostInput` and either finds a record matching the UUID given in `input.ID` or
// creates a new `models.Post` with a new UUID. In either case, all properties that are not `nil` are set to the value
// provided in `input`
func convertGqlPostInputToDBPost(ctx context.Context, input postInput, currentUser models.User) (models.Post, error) {
	post := models.Post{}

	if input.ID != nil {
		if err := post.FindByUUID(*input.ID); err != nil {
			return post, err
		}
	} else {
		if err := post.NewWithUser(*input.Type, currentUser); err != nil {
			return post, err
		}
	}

	if input.OrgID != nil {
		var org models.Organization
		err := org.FindByUUID(*input.OrgID)
		if err != nil {
			return models.Post{}, err
		}
		post.OrganizationID = org.ID
	}

	setStringField(input.Title, &post.Title)

	if input.NeededBefore == nil {
		post.NeededBefore = nulls.Time{}
	} else {
		neededBefore, err := domain.ConvertStringPtrToDate(input.NeededBefore)
		if err != nil {
			return models.Post{}, err
		}
		post.NeededBefore = nulls.NewTime(neededBefore)
	}

	setOptionalStringField(input.Description, &post.Description)

	if input.Size != nil {
		post.Size = *input.Size
	}

	setOptionalStringField(input.URL, &post.URL)
	setFloatField(input.Kilograms, &post.Kilograms)

	if input.Visibility == nil {
		post.Visibility = models.PostVisibilitySame
	} else {
		post.Visibility = *input.Visibility
	}

	if input.PhotoID == nil {
		if post.ID > 0 {
			if err := post.RemovePhoto(); err != nil {
				return models.Post{}, err
			}
		}
	} else {
		if _, err := post.AttachPhoto(*input.PhotoID); err != nil {
			return models.Post{}, err
		}
	}

	if input.MeetingID == nil {
		post.MeetingID = nulls.Int{}
	} else {
		var meeting models.Meeting
		if err := meeting.FindByUUID(*input.MeetingID); err != nil {
			return models.Post{}, fmt.Errorf("invalid meetingID, %s", err)
		}
		post.MeetingID = nulls.NewInt(meeting.ID)
		post.DestinationID = meeting.LocationID
	}

	return post, nil
}

type postInput struct {
	ID           *string
	OrgID        *string
	Type         *models.PostType
	Title        *string
	Description  *string
	NeededBefore *string
	Destination  *LocationInput
	Origin       *LocationInput
	Size         *models.PostSize
	URL          *string
	Kilograms    *float64
	PhotoID      *string
	MeetingID    *string
	Visibility   *models.PostVisibility
}

// CreatePost resolves the `createPost` mutation.
func (r *mutationResolver) CreatePost(ctx context.Context, input postInput) (*models.Post, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user": cUser.UUID,
	}

	post, err := convertGqlPostInputToDBPost(ctx, input, cUser)
	if err != nil {
		return nil, reportError(ctx, err, "CreatePost.ProcessInput", extras)
	}

	if !post.MeetingID.Valid {
		dest := convertGqlLocationInputToDBLocation(*input.Destination)
		if err = dest.Create(); err != nil {
			return nil, reportError(ctx, err, "CreatePost.SetDestination", extras)
		}
		post.DestinationID = dest.ID
	}

	if err = post.Create(); err != nil {
		return nil, reportError(ctx, err, "CreatePost", extras)
	}

	if input.Origin != nil {
		if err = post.SetOrigin(convertGqlLocationInputToDBLocation(*input.Origin)); err != nil {
			return nil, reportError(ctx, err, "CreatePost.SetOrigin", extras)
		}
	}

	return &post, nil
}

// UpdatePost resolves the `updatePost` mutation.
func (r *mutationResolver) UpdatePost(ctx context.Context, input postInput) (*models.Post, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user": cUser.UUID,
	}

	post, err := convertGqlPostInputToDBPost(ctx, input, cUser)
	if err != nil {
		return nil, reportError(ctx, err, "UpdatePost.ProcessInput", extras)
	}

	var dbPost models.Post
	_ = dbPost.FindByID(post.ID)
	if editable, err := dbPost.IsEditable(cUser); err != nil {
		return nil, reportError(ctx, err, "UpdatePost.GetEditable", extras)
	} else if !editable {
		return nil, reportError(ctx, errors.New("attempt to update a non-editable post"),
			"UpdatePost.NotEditable", extras)
	}

	if err := post.Update(); err != nil {
		return nil, reportError(ctx, err, "UpdatePost", extras)
	}

	if input.Destination != nil {
		if err := post.SetDestination(convertGqlLocationInputToDBLocation(*input.Destination)); err != nil {
			return nil, reportError(ctx, err, "UpdatePost.SetDestination", extras)
		}
	}

	if input.Origin == nil {
		if err := post.RemoveOrigin(); err != nil {
			return nil, reportError(ctx, err, "UpdatePost.RemoveOrigin", extras)
		}
	} else {
		if err := post.SetOrigin(convertGqlLocationInputToDBLocation(*input.Origin)); err != nil {
			return nil, reportError(ctx, err, "UpdatePost.SetOrigin", extras)
		}
	}

	return &post, nil
}

// UpdatePostStatus resolves the `updatePostStatus` mutation.
func (r *mutationResolver) UpdatePostStatus(ctx context.Context, input UpdatePostStatusInput) (*models.Post, error) {
	var post models.Post
	if err := post.FindByUUID(input.ID); err != nil {
		return nil, reportError(ctx, err, "UpdatePostStatus.FindPost")
	}

	cUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user":      cUser.UUID,
		"oldStatus": post.Status,
		"newStatus": input.Status,
	}
	if !cUser.CanUpdatePostStatus(post, input.Status) {
		return nil, reportError(ctx, errors.New("not allowed to change post status"),
			"UpdatePostStatus.Unauthorized", extras)
	}

	post.SetProviderWithStatus(input.Status, cUser)
	if err := post.Update(); err != nil {
		return nil, reportError(ctx, err, "UpdatePostStatus", extras)
	}

	return &post, nil
}

// SearchRequests resolves the `searchRequests` query by finding requests that contain
//  a certain string in their Title or Description
func (r *queryResolver) SearchRequests(ctx context.Context, text string) ([]models.Post, error) {
	posts := models.Posts{}
	cUser := models.GetCurrentUserFromGqlContext(ctx)

	if err := posts.FilterByUserTypeAndContents(ctx, cUser, models.PostTypeRequest, text); err != nil {
		extras := map[string]interface{}{
			"user": cUser.UUID,
		}
		return nil, reportError(ctx, err, "GetPosts", extras)
	}

	return posts, nil
}
