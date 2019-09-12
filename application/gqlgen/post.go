package gqlgen

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"

	"github.com/gobuffalo/nulls"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gobuffalo/pop"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
	"github.com/vektah/gqlparser/gqlerror"
)

func PostFields() map[string]string {
	return map[string]string{
		"id":           "uuid",
		"createdBy":    "created_by_id",
		"organization": "organization_id",
		"type":         "type",
		"title":        "title",
		"description":  "description",
		"destination":  "destination",
		"origin":       "origin",
		"size":         "size",
		"receiver":     "receiver_id",
		"provider":     "provider_id",
		"neededAfter":  "needed_after",
		"neededBefore": "needed_before",
		"category":     "category",
		"status":       "status",
		"createdAt":    "created_at",
		"updatedAt":    "updated_at",
		"url":          "url",
		"cost":         "cost",
	}
}

func (r *Resolver) Post() PostResolver {
	return &postResolver{r}
}

type postResolver struct{ *Resolver }

func (r *postResolver) ID(ctx context.Context, obj *models.Post) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.Uuid.String(), nil
}

func (r *postResolver) Type(ctx context.Context, obj *models.Post) (PostType, error) {
	if obj == nil {
		return "", nil
	}
	return PostType(obj.Type), nil
}

func (r *postResolver) CreatedBy(ctx context.Context, obj *models.Post) (*models.User, error) {
	if obj == nil {
		return nil, nil
	}
	return obj.GetCreator(GetSelectFieldsFromRequestFields(UserFields(), graphql.CollectAllFields(ctx)))
}

func (r *postResolver) Receiver(ctx context.Context, obj *models.Post) (*models.User, error) {
	if obj == nil {
		return nil, nil
	}
	return obj.GetReceiver(GetSelectFieldsFromRequestFields(UserFields(), graphql.CollectAllFields(ctx)))
}

func (r *postResolver) Provider(ctx context.Context, obj *models.Post) (*models.User, error) {
	if obj == nil {
		return nil, nil
	}
	return obj.GetProvider(GetSelectFieldsFromRequestFields(UserFields(), graphql.CollectAllFields(ctx)))
}

func (r *postResolver) Organization(ctx context.Context, obj *models.Post) (*models.Organization, error) {
	if obj == nil {
		return nil, nil
	}
	selectFields := GetSelectFieldsFromRequestFields(OrganizationFields(), graphql.CollectAllFields(ctx))
	return obj.GetOrganization(selectFields)
}

func (r *postResolver) Description(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return GetStringFromNullsString(obj.Description), nil
}

func (r *postResolver) Destination(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return GetStringFromNullsString(obj.Destination), nil
}

func (r *postResolver) Origin(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return GetStringFromNullsString(obj.Origin), nil
}

func (r *postResolver) NeededAfter(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.NeededAfter), nil
}

func (r *postResolver) NeededBefore(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.NeededBefore), nil
}

func (r *postResolver) Threads(ctx context.Context, obj *models.Post) ([]*models.Thread, error) {
	if obj == nil {
		return nil, nil
	}
	selectFields := GetSelectFieldsFromRequestFields(ThreadFields(), graphql.CollectAllFields(ctx))
	return obj.GetThreads(selectFields)
}

func (r *postResolver) MyThreadID(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return obj.GetThreadIdForUser(models.GetCurrentUserFromGqlContext(ctx, TestUser))
}

func (r *postResolver) URL(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return GetStringFromNullsString(obj.URL), nil
}

func (r *postResolver) Cost(ctx context.Context, obj *models.Post) (*string, error) {
	if (obj == nil) || (!obj.Cost.Valid) {
		return nil, nil
	}

	c := strconv.FormatFloat(obj.Cost.Float64, 'f', -1, 64)
	return &c, nil
}

func (r *queryResolver) Posts(ctx context.Context) ([]*models.Post, error) {
	var posts []*models.Post
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	selectFields := getSelectFieldsForPosts(ctx)
	if err := models.DB.Select(selectFields...).Scope(scopeUserOrgs(cUser)).All(&posts); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting posts: %v", err.Error()))
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return []*models.Post{}, err
	}

	return posts, nil
}

func (r *queryResolver) Post(ctx context.Context, id *string) (*models.Post, error) {
	post := models.Post{}
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	selectFields := getSelectFieldsForPosts(ctx)
	if err := models.DB.Select(selectFields...).Scope(scopeUserOrgs(cUser)).Where("uuid = ?", id).First(&post); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting post: %v", err.Error()))
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), map[string]interface{}{"post_id": *id})
		return &models.Post{}, err
	}

	return &post, nil
}

// convertGqlPostInputToDBPost takes a `PostInput` and either finds a record matching the UUID given in `input.ID` or
// creates a new `models.Post` with a new UUID. In either case, all properties that are not `nil` are set to the value
// provided in `input`
func convertGqlPostInputToDBPost(input postInput, createdByUser models.User) (models.Post, error) {
	post := models.Post{}

	if input.ID != nil {
		if err := post.FindByUUID(*input.ID); err != nil {
			return post, err
		}
	} else {
		post.Uuid = domain.GetUuid()
		post.CreatedByID = createdByUser.ID
	}

	if input.Status != nil {
		post.Status = input.Status.String()
	}

	if input.OrgID != nil {
		org, err := models.FindOrgByUUID(*input.OrgID)
		if err != nil {
			return models.Post{}, err
		}
		post.OrganizationID = org.ID
	}

	if input.Type != nil {
		post.Type = input.Type.String()
	}

	if input.Title != nil {
		post.Title = *input.Title
	}

	if input.Description != nil {
		post.Description = nulls.NewString(*input.Description)
	}

	if input.Destination != nil {
		post.Destination = nulls.NewString(*input.Destination)
	}

	if input.Origin != nil {
		post.Origin = nulls.NewString(*input.Origin)
	}

	if input.Size != nil {
		post.Size = *input.Size
	}

	if input.NeededAfter != nil {
		neededAfter, err := domain.ConvertStringPtrToDate(input.NeededAfter)
		if err != nil {
			err = fmt.Errorf("error converting NeededAfter %v ... %v", input.NeededAfter, err.Error())
			return models.Post{}, err
		}
		post.NeededAfter = neededAfter
	}

	if input.NeededBefore != nil {
		neededBefore, err := domain.ConvertStringPtrToDate(input.NeededBefore)
		if err != nil {
			err = fmt.Errorf("error converting NeededBefore %v ... %v", input.NeededBefore, err.Error())
			return models.Post{}, err
		}
		post.NeededBefore = neededBefore
	}

	if input.Category != nil {
		post.Category = *input.Category
	}

	if input.URL != nil {
		post.URL = nulls.NewString(*input.URL)
	}

	if input.Cost != nil {
		c, err := strconv.ParseFloat(*input.Cost, 64)
		if err != nil {
			err = fmt.Errorf("error converting cost %v ... %v", input.Cost, err.Error())
			return models.Post{}, err
		}
		post.Cost = nulls.NewFloat64(c)
	}

	return post, nil
}

func getSelectFieldsForPosts(ctx context.Context) []string {
	selectFields := GetSelectFieldsFromRequestFields(PostFields(), graphql.CollectAllFields(ctx))
	selectFields = append(selectFields, "id")
	return selectFields
}

// scope query to only include organizations for current user
func scopeUserOrgs(cUser models.User) pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		orgs := cUser.GetOrgIDs()

		// convert []int to []interface{}
		s := make([]interface{}, len(orgs))
		for i, v := range orgs {
			s[i] = v
		}

		return q.Where("organization_id IN (?)", s...)
	}
}

type postInput struct {
	ID           *string
	Status       *PostStatus
	OrgID        *string
	Type         *PostType
	Title        *string
	Description  *string
	Destination  *string
	Origin       *string
	Size         *string
	NeededAfter  *string
	NeededBefore *string
	Category     *string
	URL          *string
	Cost         *string
}

func (r *mutationResolver) CreatePost(ctx context.Context, input postInput) (*models.Post, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	post, err := convertGqlPostInputToDBPost(input, cUser)
	if err != nil {
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return &models.Post{}, err
	}

	if err := models.DB.Create(&post); err != nil {
		return &models.Post{}, err
	}

	return &post, nil
}

func (r *mutationResolver) UpdatePost(ctx context.Context, input postInput) (*models.Post, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	post, err := convertGqlPostInputToDBPost(input, cUser)
	if err != nil {
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return &models.Post{}, err
	}

	if err := models.DB.Update(&post); err != nil {
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return &models.Post{}, err
	}

	return &post, nil
}

func (r *queryResolver) PostImage(ctx context.Context, id *string) (*File, error) {
	var post models.Post
	if err := post.FindByUUID(*id); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("error getting post: %s", err))
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
	}

	var image models.Image
	if err := models.DB.Where("post_id = ?", post.ID).First(&image); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("error getting image: %s", err))
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return &File{}, err
	}

	return &File{ID: image.UUID.String()}, nil
}

func (r *mutationResolver) UploadPostImage(ctx context.Context, input NewPostImage) (*File, error) {
	if input.File.File == nil {
		return &File{}, fmt.Errorf("file is nil")
	}

	content, err := ioutil.ReadAll(input.File.File)
	if err != nil {
		return &File{}, fmt.Errorf("error reading file, %s", err)
	}

	var image models.Image
	if err := image.Store(input.PostID, content); err != nil {
		return &File{}, fmt.Errorf("failed to store image: %s", err)
	}

	if input.File.Size > math.MaxInt32 {
		return &File{}, fmt.Errorf("exceeded max file size: %d", input.File.Size)
	}
	return &File{ID: image.UUID.String(), Name: input.File.Filename, Size: int(input.File.Size)}, nil
}
