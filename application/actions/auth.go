package actions

import (
	"errors"
	"fmt"

	"net/http"
	"strconv"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

const (
	// http param for access token
	AccessTokenParam = "access-token"

	// http param and session key for Auth Email
	AuthEmailParam      = "auth-email"
	AuthEmailSessionKey = "AuthEmail"

	// http param and session key for Client ID
	ClientIDParam      = "client-id"
	ClientIDSessionKey = "ClientID"

	// http param for expires utc
	ExpiresUTCParam = "expires-utc"

	// logout http param for what is normally the bearer token
	LogoutToken = "token"

	// http param for organization id
	OrgIDParam      = "org-id"
	OrgIDSessionKey = "OrgID"

	// http param and session key for ReturnTo
	ReturnToParam      = "return-to"
	ReturnToSessionKey = "ReturnTo"

	// http param for token type
	TokenTypeParam = "token-type"
)

type AuthOrgOption struct {
	ID      string `json:"ID"`
	Name    string `json:"Name"`
	LogoURL string `json:"LogoURL"`
}

type AuthUser struct {
	ID                   string          `json:"ID"`
	Name                 string          `json:"Name"`
	Nickname             string          `json:"Nickname"`
	Email                string          `json:"Email"`
	Organizations        []AuthOrgOption `json:"Organizations"`
	AccessToken          string          `json:"AccessToken"`
	AccessTokenExpiresAt int64           `json:"AccessTokenExpiresAt"`
	IsNew                bool            `json:"IsNew"`
}

type AuthResponse struct {
	Error          *domain.AppError `json:"Error,omitempty"`
	AuthOrgOptions *[]AuthOrgOption `json:"AuthOrgOptions,omitempty"`
	RedirectURL    string           `json:"RedirectURL,omitempty"`
	User           *AuthUser        `json:"User,omitempty"`
}

func getOrSetReturnTo(c buffalo.Context) string {
	returnTo := c.Param(ReturnToParam)

	if returnTo == "" {
		var ok bool
		returnTo, ok = c.Session().Get(ReturnToSessionKey).(string)
		if !ok {
			returnTo = "/#"
		}

		return returnTo
	}

	c.Session().Set(ReturnToSessionKey, returnTo)

	return returnTo
}

func getOrgAndUserOrgs(
	authEmail string,
	c buffalo.Context) (models.Organization, models.UserOrganizations, error) {
	var orgID int
	oid := c.Param(OrgIDParam)
	if oid == "" {
		orgID = 0
	} else {
		orgID, _ = strconv.Atoi(oid)
	}

	var org models.Organization
	var userOrgs models.UserOrganizations
	err := userOrgs.FindByAuthEmail(authEmail, orgID)
	if len(userOrgs) == 1 {
		org = userOrgs[0].Organization
	}

	// no user_organization records yet, see if we have an organization for user's email domain
	if len(userOrgs) == 0 {
		err = org.FindByDomain(domain.EmailDomain(authEmail))
		if err != nil {
			return org, userOrgs, err
		}
		if org.AuthType == "" {
			return org, userOrgs, errors.New("unable to find organization by email domain")
		}
	}

	return org, userOrgs, nil
}

func provideOrgOptions(userOrgs models.UserOrganizations, c buffalo.Context) error {
	var orgOpts []AuthOrgOption
	for _, uo := range userOrgs {
		orgOpts = append(orgOpts, AuthOrgOption{
			ID:      strconv.Itoa(uo.ID),
			Name:    uo.Organization.Name,
			LogoURL: "",
		})
	}

	resp := AuthResponse{
		AuthOrgOptions: &orgOpts,
	}
	return c.Render(http.StatusOK, render.JSON(resp))
}

func createAuthUser(
	clientID string,
	user models.User,
	org models.Organization) (AuthUser, error) {
	accessToken, expiresAt, err := user.CreateAccessToken(org, clientID)

	if err != nil {
		return AuthUser{}, err
	}

	var uos []AuthOrgOption
	for _, uo := range user.Organizations {
		uos = append(uos, AuthOrgOption{
			ID:      uo.UUID.String(),
			Name:    uo.Name,
			LogoURL: uo.Url.String,
		})
	}

	isNew := false
	if time.Since(user.CreatedAt) < time.Duration(time.Second*30) {
		isNew = true
	}

	authUser := AuthUser{
		ID:                   user.UUID.String(),
		Name:                 user.FirstName + " " + user.LastName,
		Nickname:             user.Nickname,
		Email:                user.Email,
		Organizations:        uos,
		AccessToken:          accessToken,
		AccessTokenExpiresAt: expiresAt,
		IsNew:                isNew,
	}

	return authUser, nil
}

func authRequest(c buffalo.Context) error {
	clientID := c.Param(ClientIDParam)
	if clientID == "" {
		return authRequestError(c, http.StatusBadRequest, domain.MissingClientID,
			ClientIDParam+" is required to login")
	}

	c.Session().Set(ClientIDSessionKey, clientID)

	authEmail := c.Param(AuthEmailParam)
	if authEmail == "" {
		return authRequestError(c, http.StatusBadRequest, domain.MissingAuthEmail,
			AuthEmailParam+" is required to login")
	}
	c.Session().Set(AuthEmailSessionKey, authEmail)

	getOrSetReturnTo(c)

	extras := map[string]interface{}{"authEmail": authEmail}

	// find org for auth config and processing
	org, userOrgs, err := getOrgAndUserOrgs(authEmail, c)
	if err != nil {
		return authRequestError(c, http.StatusInternalServerError, domain.ErrorFindingOrgUserOrgs,
			fmt.Sprintf("error getting org and userOrgs ... %v", err), extras)
	}

	// User has more than one organization affiliation, return list to choose from
	if len(userOrgs) > 1 {
		return provideOrgOptions(userOrgs, c)
	}

	if org.ID == 0 {
		return authRequestError(c, http.StatusBadRequest, domain.CannotFindOrg,
			"unable to find an organization for this user", extras)
	}

	orgID := org.UUID.String()
	c.Session().Set(OrgIDSessionKey, orgID)
	err = c.Session().Save()
	if err != nil {
		return authRequestError(c, http.StatusInternalServerError, domain.ErrorSavingAuthRequestSession,
			fmt.Sprintf("unable to save session ... %v", err), extras)
	}

	sp, err := org.GetAuthProvider()
	if err != nil {
		return authRequestError(c, http.StatusInternalServerError, domain.ErrorLoadingAuthProvider,
			fmt.Sprintf("unable to load auth provider for '%s' ... %v", org.Name, err), extras)
	}

	redirectURL, err := sp.AuthRequest(c)
	if err != nil {
		return authRequestError(c, http.StatusInternalServerError, domain.ErrorGettingAuthURL,
			fmt.Sprintf("unable to determine what the authentication url should be for '%s' ... %v", org.Name, err))
	}

	resp := AuthResponse{RedirectURL: redirectURL}

	// Reply with a 200 and leave it to the UI to do the redirect
	return c.Render(http.StatusOK, render.JSON(resp))

}

// authCallback assumes the user has logged in to the IDP or Oauth service and now their browser
// has been redirected back with the final response
func authCallback(c buffalo.Context) error {
	clientID, ok := c.Session().Get(ClientIDSessionKey).(string)
	if !ok {
		return logErrorAndRedirect(c, domain.MissingSessionClientID,
			ClientIDSessionKey+" session entry is required to complete login")
	}

	authEmail, ok := c.Session().Get(AuthEmailSessionKey).(string)
	if !ok {
		return logErrorAndRedirect(c, domain.MissingSessionAuthEmail,
			AuthEmailSessionKey+" session entry is required to complete login")
	}

	returnTo := getOrSetReturnTo(c)

	err := c.Session().Save()
	if err != nil {
		extras := map[string]interface{}{"authEmail": authEmail}
		return logErrorAndRedirect(c, domain.ErrorSavingAuthCallbackSession,
			fmt.Sprintf("error saving session ... %v", err), extras)
	}

	orgID, ok := c.Session().Get(OrgIDSessionKey).(string)
	if !ok {
		return logErrorAndRedirect(c, domain.MissingSessionOrgID,
			OrgIDSessionKey+" session entry is required to complete login")
	}

	org := models.Organization{}
	err = org.FindByUUID(orgID)
	if err != nil {
		return logErrorAndRedirect(c, domain.ErrorFindingOrg,
			fmt.Sprintf("error finding org with UUID %s ... %v", orgID, err.Error()))
	}

	ap, err := org.GetAuthProvider()
	if err != nil {
		extras := map[string]interface{}{"authEmail": authEmail}
		return logErrorAndRedirect(c, domain.ErrorLoadingAuthProvider,
			fmt.Sprintf("error loading auth provider for '%s' ... %v", org.Name, err), extras)
	}

	authResp := ap.AuthCallback(c)
	if authResp.Error != nil {
		extras := map[string]interface{}{"authEmail": authEmail}
		return logErrorAndRedirect(c, domain.ErrorAuthProvidersCallback, authResp.Error.Error(), extras)
	}

	// if we have an authuser, find or create user in local db and finish login
	var user models.User
	if authResp.AuthUser != nil {
		// login was success, clear session so new login can be initiated if needed
		c.Session().Clear()
		err := c.Session().Save()
		if err != nil {
			extras := map[string]interface{}{"authEmail": authEmail}
			return logErrorAndRedirect(c, domain.ErrorSavingAuthCallbackSession,
				fmt.Sprintf("error saving session after clear... %v", err), extras)
		}

		err = user.FindOrCreateFromAuthUser(org.ID, authResp.AuthUser)
		if err != nil {
			return logErrorAndRedirect(c, domain.ErrorWithAuthUser, err.Error())
		}
	}

	authUser, err := createAuthUser(clientID, user, org)
	if err != nil {
		return err
	}

	// set person on rollbar session
	domain.RollbarSetPerson(c, authUser.ID, authUser.Nickname, authUser.Email)

	return c.Redirect(302, getLoginSuccessRedirectURL(authUser, returnTo))
}

func mergeExtras(code string, extras ...map[string]interface{}) map[string]interface{} {
	allExtras := map[string]interface{}{"code": code}

	for _, e := range extras {
		for k, v := range e {
			allExtras[k] = v
		}
	}
	return allExtras
}

// Make extras variadic, so that it can be omitted from the params
func authRequestError(c buffalo.Context, httpStatus int, errorCode, message string, extras ...map[string]interface{}) error {
	allExtras := mergeExtras(errorCode, extras...)

	domain.Error(c, message, allExtras)

	authError := domain.AppError{
		Code: errorCode,
	}
	return c.Render(httpStatus, render.JSON(authError))
}

// Make extras variadic, so that it can be omitted from the params
func logErrorAndRedirect(c buffalo.Context, code, message string, extras ...map[string]interface{}) error {
	allExtras := mergeExtras(code, extras...)

	domain.Error(c, message, allExtras)

	uiUrl := domain.Env.UIURL + "/#/login?error=true"
	return c.Redirect(http.StatusFound, uiUrl)
}

// authDestroy uses the bearer token to find the user's access token and
//  calls the appropriate provider's logout function.
func authDestroy(c buffalo.Context) error {
	tokenParam := c.Param(LogoutToken)
	if tokenParam == "" {
		return logErrorAndRedirect(c, domain.MissingLogoutToken,
			LogoutToken+" is required to logout")
	}

	var uat models.UserAccessToken
	err := uat.FindByBearerToken(tokenParam)
	if err != nil {
		return logErrorAndRedirect(c, domain.ErrorFindingAccessToken, err.Error())
	}

	org, err := uat.GetOrganization()
	if err != nil {
		return logErrorAndRedirect(c, domain.ErrorFindingOrgForAccessToken, err.Error())
	}

	// set person on rollbar session
	domain.RollbarSetPerson(c, uat.User.UUID.String(), uat.User.Nickname, uat.User.Email)

	authPro, err := org.GetAuthProvider()
	if err != nil {
		return logErrorAndRedirect(c, domain.ErrorLoadingAuthProvider, err.Error())
	}

	authResp := authPro.Logout(c)
	if authResp.Error != nil {
		return logErrorAndRedirect(c, domain.ErrorAuthProvidersLogout, authResp.Error.Error())
	}

	redirectURL := domain.Env.UIURL

	if authResp.RedirectURL != "" {
		var uat models.UserAccessToken
		err = uat.DeleteByBearerToken(tokenParam)
		if err != nil {
			return logErrorAndRedirect(c, domain.ErrorDeletingAccessToken, err.Error())
		}
		c.Session().Clear()
		redirectURL = authResp.RedirectURL
	}

	return c.Redirect(http.StatusFound, redirectURL)
}

func setCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		bearerToken := domain.GetBearerTokenFromRequest(c.Request())
		if bearerToken == "" {
			return c.Error(http.StatusUnauthorized, errors.New("no Bearer token provided"))
		}

		var userAccessToken models.UserAccessToken
		err := userAccessToken.FindByBearerToken(bearerToken)
		if err != nil {
			domain.ErrLogger.Print(err.Error())
			return c.Error(http.StatusUnauthorized, errors.New("invalid bearer token"))
		}

		isExpired, err := userAccessToken.DeleteIfExpired()
		if err != nil {
			domain.ErrLogger.Print(err.Error())
		}

		if isExpired {
			return c.Error(http.StatusUnauthorized, errors.New("expired bearer token"))
		}

		user, err := userAccessToken.GetUser()
		if err != nil {
			return c.Error(http.StatusInternalServerError, fmt.Errorf("error finding user by access token, %s", err.Error()))
		}
		c.Set("current_user", user)

		// set person on rollbar session
		domain.RollbarSetPerson(c, user.UUID.String(), user.Nickname, user.Email)
		msg := fmt.Sprintf("user %s authenticated with bearer token from ip %s", user.Email, c.Request().RemoteAddr)
		extras := map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      c.Request().RemoteAddr,
		}
		domain.Info(c, msg, extras)

		if err := userAccessToken.Renew(); err != nil {
			domain.Error(c, fmt.Sprintf("error renewing access token, %s", c.Request().RemoteAddr), extras)
		}

		return next(c)
	}
}

// getLoginSuccessRedirectURL generates the URL for redirection after a successful login
func getLoginSuccessRedirectURL(authUser AuthUser, returnTo string) string {

	uiUrl := domain.Env.UIURL + "/#"

	tokenExpiry := time.Unix(authUser.AccessTokenExpiresAt, 0).Format(time.RFC3339)
	params := fmt.Sprintf("?%s=Bearer&%s=%s&%s=%s",
		TokenTypeParam, ExpiresUTCParam, tokenExpiry, AccessTokenParam, authUser.AccessToken)

	if authUser.IsNew {
		uiUrl += "/welcome"
		if len(returnTo) > 0 {
			params += "&" + ReturnToParam + "=" + returnTo
		}
	} else {
		if len(returnTo) > 0 && returnTo[0] != '/' {
			returnTo = "/" + returnTo
		}
		uiUrl += returnTo
	}

	url := fmt.Sprintf("%s%s", uiUrl, params)

	return url
}
