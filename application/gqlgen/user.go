package gqlgen

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// PostRoleMap is used to convert PostRole gql enum values to values used by models
var PostRoleMap = map[PostRole]string{
	PostRoleCreatedby: models.PostsCreated,
	PostRoleReceiving: models.PostsReceiving,
	PostRoleProviding: models.PostsProviding,
}

// UserFields maps GraphQL fields to their equivalent database fields. For related types, the
// foreign key field name is provided.
func UserFields() map[string]string {
	return map[string]string{
		"id":          "uuid",
		"email":       "email",
		"nickname":    "nickname",
		"accessToken": "access_token",
		"createdAt":   "created_at",
		"updatedAt":   "updated_at",
		"adminRole":   "admin_role",
		"photoURL":    "photo_url",
		"photoFile":   "photo_file_id",
		"location":    "location_id",
	}
}

// User is required by gqlgen
func (r *Resolver) User() UserResolver {
	return &userResolver{r}
}

type userResolver struct{ *Resolver }

// ID provides the UUID instead of the autoincrement ID.
func (r *userResolver) ID(ctx context.Context, obj *models.User) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.Uuid.String(), nil
}

// AdminRole converts the models admin roles to gql AdminRole enum values
func (r *userResolver) AdminRole(ctx context.Context, obj *models.User) (*Role, error) {
	if obj == nil {
		return nil, nil
	}
	a := Role(obj.AdminRole.String)
	return &a, nil
}

// Organizations retrieves the list of Organizations to which the queried user is associated
func (r *userResolver) Organizations(ctx context.Context, obj *models.User) ([]models.Organization, error) {
	if obj == nil {
		return nil, nil
	}

	organizations, err := obj.GetOrganizations()
	if err != nil {
		return nil, reportError(ctx, err, "GetUserOrganizations")
	}

	return organizations, nil
}

// Posts retrieves the list of Posts associated with the queried user, where association is defined by the given `role`.
func (r *userResolver) Posts(ctx context.Context, obj *models.User, role PostRole) ([]models.Post, error) {
	if obj == nil {
		return nil, nil
	}

	posts, err := obj.GetPosts(PostRoleMap[role])
	if err != nil {
		extras := map[string]interface{}{
			"role": role,
		}
		return nil, reportError(ctx, err, "GetUserPosts", extras)
	}

	return posts, nil
}

// PhotoURL retrieves a URL for the user profile photo or avatar. It can either be an attached photo or
// a photo belonging to an external profile such as Gravatar or Google.
func (r *userResolver) PhotoURL(ctx context.Context, obj *models.User) (string, error) {
	if obj == nil {
		return "", nil
	}

	photoURL, err := obj.GetPhotoURL()
	if err != nil {
		return "", reportError(ctx, err, "GetUserPhotoURL")
	}

	return photoURL, nil
}

// Location retrieves the queried user's location.
func (r *userResolver) Location(ctx context.Context, obj *models.User) (*models.Location, error) {
	if obj == nil {
		return nil, nil
	}

	location, err := obj.GetLocation()
	if err != nil {
		return nil, reportError(ctx, err, "GetUserLocation")
	}

	return location, nil
}

// UnreadMessageCount calculates the number of unread messages for the queried user
func (r *userResolver) UnreadMessageCount(ctx context.Context, obj *models.User) (int, error) {
	if obj == nil {
		return 0, nil
	}
	mCounts, err := obj.UnreadMessageCount()

	if err != nil {
		return 0, reportError(ctx, err, "GetUserUnreadMessageCount")
	}
	total := 0
	for _, c := range mCounts {
		total += c.Count
	}

	return total, nil
}

// Users retrieves a list of users
func (r *queryResolver) Users(ctx context.Context) ([]models.User, error) {
	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	role := currentUser.AdminRole.String
	if role != domain.AdminRoleSuperDuperAdmin {
		err := errors.New("insufficient permissions")
		extras := map[string]interface{}{
			"role": role,
		}
		return nil, reportError(ctx, err, "GetUsers.NotAllowed", extras)
	}

	users := models.Users{}
	selectFields := GetSelectFieldsForUsers(ctx)
	if err := users.All(selectFields...); err != nil {
		extras := map[string]interface{}{
			"fields": selectFields,
		}
		return nil, reportError(ctx, err, "GetUsers", extras)
	}

	return users, nil
}

// User retrieves a single user
func (r *queryResolver) User(ctx context.Context, id *string) (*models.User, error) {
	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	if id == nil {
		return &currentUser, nil
	}

	role := currentUser.AdminRole.String
	if role != domain.AdminRoleSuperDuperAdmin && currentUser.Uuid.String() != *id {
		err := errors.New("insufficient permissions")
		extras := map[string]interface{}{
			"role": role,
		}
		return nil, reportError(ctx, err, "GetUser.NotAllowed", extras)
	}

	dbUser := models.User{}
	selectFields := GetSelectFieldsForUsers(ctx)
	if err := dbUser.FindByUUID(*id, selectFields...); err != nil {
		extras := map[string]interface{}{
			"fields": selectFields,
		}
		return nil, reportError(ctx, err, "GetUser", extras)
	}

	return &dbUser, nil
}

// GetSelectFieldsForUsers returns a list of database fields appropriate for the current query. Foreign keys
// will be included as needed.
func GetSelectFieldsForUsers(ctx context.Context) []string {
	selectFields := GetSelectFieldsFromRequestFields(UserFields(), graphql.CollectAllFields(ctx))
	selectFields = append(selectFields, "id")
	if domain.IsStringInSlice("photoURL", graphql.CollectAllFields(ctx)) {
		selectFields = append(selectFields, "photo_file_id")
	}
	return selectFields
}

// UpdateUser takes data from the GraphQL `UpdateUser` mutation and updates the database. If the
// user ID is provided and the current user is allowed to edit profiles, that user will be updated.
// Otherwise, the current authenticated user is updated.
func (r *mutationResolver) UpdateUser(ctx context.Context, input UpdateUserInput) (*models.User, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	var user models.User

	if input.ID != nil {
		if err := user.FindByUUID(*(input.ID)); err != nil {
			return nil, reportError(ctx, err, "UpdateUser.NotFound")
		}
	} else {
		user = cUser
	}

	if cUser.AdminRole.String != domain.AdminRoleSuperDuperAdmin && cUser.ID != user.ID {
		err := errors.New("insufficient permissions")
		return nil, reportError(ctx, err, "UpdateUser.NotAllowed")
	}

	if input.PhotoID != nil {
		var file models.File
		if err := file.FindByUUID(*input.PhotoID); err != nil {
			return nil, reportError(ctx, err, "UpdateUser.PhotoNotFound")
		}
		user.PhotoFileID = nulls.NewInt(file.ID)
	}

	if input.Location != nil {
		if err := user.SetLocation(convertGqlLocationInputToDBLocation(*input.Location)); err != nil {
			return nil, reportError(ctx, err, "UpdateUser.SetLocationError")
		}
	}

	if err := user.Save(); err != nil {
		return nil, reportError(ctx, err, "UpdateUser")
	}

	return &user, nil
}
