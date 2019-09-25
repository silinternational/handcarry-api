package gqlgen

import (
	"context"

	"github.com/99designs/gqlgen/graphql"

	"github.com/silinternational/wecarry-api/models"
)

func OrganizationFields() map[string]string {
	return map[string]string{
		"id":         "uuid",
		"name":       "name",
		"url":        "url",
		"authType":   "auth_type",
		"authConfig": "auth_config",
		"createdAt":  "created_at",
		"updatedAt":  "updated_at",
	}
}

func (r *Resolver) Organization() OrganizationResolver {
	return &organizationResolver{r}
}

type organizationResolver struct{ *Resolver }

func (r *organizationResolver) ID(ctx context.Context, obj *models.Organization) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.Uuid.String(), nil
}

func (r *organizationResolver) URL(ctx context.Context, obj *models.Organization) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return GetStringFromNullsString(obj.Url), nil
}

func (r *organizationResolver) Domains(ctx context.Context, obj *models.Organization) ([]*models.OrganizationDomain, error) {
	if obj == nil {
		return nil, nil
	}

	if err := models.DB.Load(obj, "OrganizationDomains"); err != nil {
		return nil, err
	}
	domains := obj.OrganizationDomains
	dp := make([]*models.OrganizationDomain, len(domains))
	for i, d := range domains {
		dp[i] = &d
	}
	return dp, nil
}

func getSelectFieldsForOrganizations(ctx context.Context) []string {
	selectFields := GetSelectFieldsFromRequestFields(OrganizationFields(), graphql.CollectAllFields(ctx))
	selectFields = append(selectFields, "id")
	return selectFields
}
