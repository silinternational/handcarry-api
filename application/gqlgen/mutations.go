package gqlgen

import (
	"context"
	"errors"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type mutationResolver struct{ *Resolver }

// CreateOrganization adds a new organization, if the current user has appropriate permissions.
func (r *mutationResolver) CreateOrganization(ctx context.Context, input CreateOrganizationInput) (*models.Organization, error) {
	cUser := models.CurrentUser(ctx)

	if !cUser.CanCreateOrganization() {
		domain.NewExtra(ctx, "user.admin_role", cUser.AdminRole)
		err := errors.New("insufficient permissions")
		return &models.Organization{}, domain.ReportError(ctx, err, "CreateOrganization.Unauthorized")
	}

	org := models.Organization{
		Name:       input.Name,
		Url:        models.ConvertStringPtrToNullsString(input.URL),
		AuthType:   input.AuthType,
		AuthConfig: input.AuthConfig,
	}

	if input.LogoFileID != nil {
		if _, err := org.AttachLogo(*input.LogoFileID); err != nil {
			return &models.Organization{}, domain.ReportError(ctx, err, "CreateOrganization.LogoFileNotFound")
		}
	}

	if err := org.Save(); err != nil {
		return &models.Organization{}, domain.ReportError(ctx, err, "CreateOrganization")
	}

	return &org, nil
}

// UpdateOrganization updates an organization, if the current user has appropriate permissions.
func (r *mutationResolver) UpdateOrganization(ctx context.Context, input UpdateOrganizationInput) (*models.Organization, error) {
	cUser := models.CurrentUser(ctx)

	var org models.Organization
	if err := org.FindByUUID(input.ID); err != nil {
		return &models.Organization{}, domain.ReportError(ctx, err, "UpdateOrganization.NotFound")
	}

	if !cUser.CanEditOrganization(org.ID) {
		err := errors.New("insufficient permissions")
		return &models.Organization{}, domain.ReportError(ctx, err, "UpdateOrganization.Unauthorized")
	}

	org.Url = models.ConvertStringPtrToNullsString(input.URL)

	if input.LogoFileID != nil {
		if _, err := org.AttachLogo(*input.LogoFileID); err != nil {
			return &models.Organization{}, domain.ReportError(ctx, err, "UpdateOrganization.LogoFileNotFound")
		}
	} else {
		if err := org.RemoveFile(); err != nil {
			return &models.Organization{}, domain.ReportError(ctx, err, "UpdateOrganization.RemoveLogo")
		}
	}

	org.Name = input.Name
	org.AuthType = input.AuthType
	org.AuthConfig = input.AuthConfig
	if err := org.Save(); err != nil {
		return &models.Organization{}, domain.ReportError(ctx, err, "UpdateOrganization")
	}

	return &org, nil
}

// CreateOrganizationDomain is the resolver for the `createOrganizationDomain` mutation
func (r *mutationResolver) CreateOrganizationDomain(ctx context.Context, input CreateOrganizationDomainInput) ([]models.OrganizationDomain, error) {
	cUser := models.CurrentUser(ctx)

	var org models.Organization
	if err := org.FindByUUID(input.OrganizationID); err != nil {
		return nil, domain.ReportError(ctx, err, "CreateOrganizationDomain.NotFound")
	}

	if !cUser.CanEditOrganization(org.ID) {
		err := errors.New("insufficient permissions")
		return nil, domain.ReportError(ctx, err, "CreateOrganizationDomain.Unauthorized")
	}

	if err := org.AddDomain(input.Domain, input.AuthType, domain.ConvertStrPtrToString(input.AuthConfig)); err != nil {
		return nil, domain.ReportError(ctx, err, "CreateOrganizationDomain")
	}

	domains, err2 := org.Domains()
	if err2 != nil {
		// don't return an error since the AddDomain operation succeeded
		_ = domain.ReportError(ctx, err2, "")
	}

	return domains, nil
}

// UpdateOrganizationDomain is the resolver for the `updateOrganizationDomain` mutation
func (r *mutationResolver) UpdateOrganizationDomain(ctx context.Context, input CreateOrganizationDomainInput) ([]models.OrganizationDomain, error) {
	cUser := models.CurrentUser(ctx)

	var org models.Organization
	if err := org.FindByUUID(input.OrganizationID); err != nil {
		return nil, domain.ReportError(ctx, err, "UpdateOrganizationDomain.NotFound")
	}

	if !cUser.CanEditOrganization(org.ID) {
		err := errors.New("insufficient permissions")
		return nil, domain.ReportError(ctx, err, "UpdateOrganizationDomain.Unauthorized")
	}

	var orgDomain models.OrganizationDomain
	if err := orgDomain.FindByDomain(input.Domain); err != nil {
		return nil, domain.ReportError(ctx, err, "UpdateOrganizationDomain.NotFound")
	}

	orgDomain.AuthType = input.AuthType
	orgDomain.AuthConfig = domain.ConvertStrPtrToString(input.AuthConfig)
	if err := orgDomain.Save(); err != nil {
		return nil, domain.ReportError(ctx, err, "UpdateOrganizationDomain.SaveError")
	}

	domains, err2 := org.Domains()
	if err2 != nil {
		// don't return an error since the operation succeeded
		_ = domain.ReportError(ctx, err2, "")
	}

	return domains, nil
}

// RemoveOrganizationDomain is the resolver for the `removeOrganizationDomain` mutation
func (r *mutationResolver) RemoveOrganizationDomain(ctx context.Context, input RemoveOrganizationDomainInput) ([]models.OrganizationDomain, error) {
	cUser := models.CurrentUser(ctx)

	var org models.Organization
	if err := org.FindByUUID(input.OrganizationID); err != nil {
		return nil, domain.ReportError(ctx, err, "RemoveOrganizationDomain.NotFound")
	}

	if !cUser.CanEditOrganization(org.ID) {
		err := errors.New("insufficient permissions")
		return nil, domain.ReportError(ctx, err, "RemoveOrganizationDomain.Unauthorized")
	}

	if err := org.RemoveDomain(input.Domain); err != nil {
		return nil, domain.ReportError(ctx, err, "RemoveOrganizationDomain")
	}

	domains, err2 := org.Domains()
	if err2 != nil {
		// don't return an error since the RemoveDomain operation succeeded
		_ = domain.ReportError(ctx, err2, "")
	}

	return domains, nil
}

// SetThreadLastViewedAt sets the last viewed time for the current user on the given thread
func (r *mutationResolver) SetThreadLastViewedAt(ctx context.Context, input SetThreadLastViewedAtInput) (*models.Thread, error) {
	cUser := models.CurrentUser(ctx)

	var thread models.Thread
	if err := thread.FindByUUID(input.ThreadID); err != nil {
		return &models.Thread{}, domain.ReportError(ctx, err, "SetThreadLastViewedAt.NotFound")
	}

	if err := thread.UpdateLastViewedAt(cUser.ID, input.Time); err != nil {
		return &models.Thread{}, domain.ReportError(ctx, err, "SetThreadLastViewedAt")
	}

	return &thread, nil
}

// CreateOrganizationTrust establishes a OrganizationTrust between two organizations
func (r *mutationResolver) CreateOrganizationTrust(ctx context.Context, input CreateOrganizationTrustInput) (*models.Organization, error) {
	cUser := models.CurrentUser(ctx)

	var organization models.Organization
	if err := organization.FindByUUID(input.PrimaryID); err != nil {
		return &models.Organization{}, domain.ReportError(ctx, err, "CreateOrganizationTrust.FindPrimaryOrganization")
	}

	if !cUser.CanCreateOrganizationTrust() {
		err := errors.New("insufficient permissions")
		return &models.Organization{}, domain.ReportError(ctx, err, "CreateOrganizationTrust.Unauthorized")
	}

	if err := organization.CreateTrust(input.SecondaryID); err != nil {
		return &models.Organization{}, domain.ReportError(ctx, err, "CreateOrganizationTrust")
	}

	return &organization, nil
}

// RemoveOrganizationTrust removes a OrganizationTrust between two organizations
func (r *mutationResolver) RemoveOrganizationTrust(ctx context.Context, input RemoveOrganizationTrustInput) (*models.Organization, error) {
	cUser := models.CurrentUser(ctx)

	var organization models.Organization
	if err := organization.FindByUUID(input.PrimaryID); err != nil {
		return &models.Organization{}, domain.ReportError(ctx, err, "RemoveOrganizationTrust.FindPrimaryOrganization")
	}

	if !cUser.CanRemoveOrganizationTrust(organization.ID) {
		err := errors.New("insufficient permissions")
		return &models.Organization{}, domain.ReportError(ctx, err, "RemoveOrganizationTrust.Unauthorized")
	}

	if err := organization.RemoveTrust(input.SecondaryID); err != nil {
		return &models.Organization{}, domain.ReportError(ctx, err, "RemoveOrganizationTrust")
	}

	return &organization, nil
}

// CreateMeetingInvites implements the `createMeetingInvites` mutation
func (r *mutationResolver) CreateMeetingInvites(ctx context.Context, input CreateMeetingInvitesInput) (
	[]models.MeetingInvite, error) {

	cUser := models.CurrentUser(ctx)

	var m models.Meeting
	if err := m.FindByUUID(input.MeetingID); err != nil {
		return nil, domain.ReportError(ctx, err, "CreateMeetingInvite.FindMeeting")
	}

	c := domain.GetBuffaloContext(ctx)
	if !cUser.CanCreateMeetingInvite(c, m) {
		err := errors.New("insufficient permissions")
		return nil, domain.ReportError(ctx, err, "CreateMeetingInvite.Unauthorized")
	}

	inv := models.MeetingInvite{
		MeetingID: m.ID,
		InviterID: cUser.ID,
	}

	badEmails := make([]string, 0)
	for _, email := range input.Emails {
		inv.Email = email
		if err := inv.Create(); err != nil {
			badEmails = append(badEmails, email)
			domain.ErrLogger.Printf("error creating meeting invite for email '%s', %s", email, err)
		}
	}
	if len(badEmails) > 0 {
		emailList := strings.Join(badEmails, ", ")
		graphql.AddError(ctx, gqlerror.Errorf("problem creating invite for %v", emailList))
	}

	invites, err := m.Invites(domain.GetBuffaloContext(ctx))
	if err != nil {
		return nil, domain.ReportError(ctx, err, "CreateMeetingInvite.ListInvites")
	}
	return invites, nil
}

func (r *mutationResolver) RemoveMeetingInvite(ctx context.Context, input RemoveMeetingInviteInput) ([]models.MeetingInvite, error) {
	var meeting models.Meeting
	if err := meeting.FindByUUID(input.MeetingID); err != nil {
		return nil, domain.ReportError(ctx, err, "RemoveMeetingInvite.FindMeeting")
	}

	c := domain.GetBuffaloContext(ctx)
	cUser := models.CurrentUser(ctx)
	if !cUser.CanRemoveMeetingInvite(c, meeting) {
		err := errors.New("insufficient permissions")
		return nil, domain.ReportError(ctx, err, "RemoveMeetingInvite.Unauthorized")
	}

	if err := meeting.RemoveInvite(c, input.Email); err != nil {
		return nil, domain.ReportError(ctx, err, "RemoveMeetingInvite")
	}

	invites, err := meeting.Invites(c)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "RemoveMeetingInvite.ListInvites")
	}
	return invites, nil
}

// CreateMeetingParticipant implements the `createMeetingParticipant` mutation
func (r *mutationResolver) CreateMeetingParticipant(ctx context.Context, input CreateMeetingParticipantInput) (
	*models.MeetingParticipant, error) {

	var meeting models.Meeting
	if err := meeting.FindByUUID(input.MeetingID); err != nil {
		return &models.MeetingParticipant{}, domain.ReportError(ctx, err, "CreateMeetingParticipant.FindMeeting")
	}

	var participant models.MeetingParticipant
	if err := participant.FindOrCreate(ctx, meeting, input.Code); err != nil {
		// MeetingParticipant.Create returns localized error messages
		return &models.MeetingParticipant{}, err
	}
	return &participant, nil
}

func (r *mutationResolver) RemoveMeetingParticipant(ctx context.Context, input RemoveMeetingParticipantInput) ([]models.MeetingParticipant, error) {
	var meeting models.Meeting
	if err := meeting.FindByUUID(input.MeetingID); err != nil {
		return nil, domain.ReportError(ctx, err, "RemoveMeetingParticipant.FindMeeting")
	}

	c := domain.GetBuffaloContext(ctx)
	cUser := models.CurrentUser(ctx)
	if !cUser.CanRemoveMeetingParticipant(c, meeting) {
		err := errors.New("insufficient permissions")
		return nil, domain.ReportError(ctx, err, "RemoveMeetingParticipant.Unauthorized")
	}

	if err := meeting.RemoveParticipant(c, input.UserID); err != nil {
		return nil, domain.ReportError(ctx, err, "RemoveMeetingParticipant")
	}

	participants, err := meeting.Participants(c)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "RemoveMeetingParticipant.ListParticipants")
	}
	return participants, nil
}
