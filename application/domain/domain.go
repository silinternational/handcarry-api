package domain

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	uuid2 "github.com/gofrs/uuid"

	"github.com/silinternational/handcarry-api/models"
)

const ClientIDKey = "client_id"

const ErrorLevelWarn = "warn"
const ErrorLevelError = "error"
const ErrorLevelCritical = "critical"

const AdminRoleSuperDuperAdmin = "SuperDuperAdmin"

const EmptyUUID = "00000000-0000-0000-0000-000000000000"

type AppError struct {
	Err   error
	Code  int
	Level string
}

func GetRequestData(r *http.Request) (map[string][]string, error) {
	data := map[string][]string{}

	if r.Method == "GET" {
		return r.URL.Query(), nil
	}

	if r.Method == "POST" || r.Method == "PUT" {
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			return data, fmt.Errorf("error getting POST data: %v", err.Error())
		}

		data = r.PostForm
	}

	return data, nil
}

func GetFirstStringFromSlice(strSlice []string) string {
	if len(strSlice) > 0 {
		return strSlice[0]
	}

	return ""
}

func GetBearerTokenFromRequest(r *http.Request) string {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		return ""
	}

	re := regexp.MustCompile(`^Bearer (.*)$`)
	matches := re.FindSubmatch([]byte(authorizationHeader))
	if len(matches) < 2 {
		return ""
	}

	return string(matches[1])
}

func GetSubPartKeyValues(inString, outerDelimiter, innerDelimiter string) map[string]string {
	keyValues := map[string]string{}
	allPairs := strings.Split(inString, outerDelimiter)

	for _, p := range allPairs {
		pParts := strings.Split(p, innerDelimiter)
		if len(pParts) == 2 {
			keyValues[pParts[0]] = pParts[1]
		}
	}

	return keyValues
}

func GetCurrentUser(c buffalo.Context) models.User {
	user := c.Value("current_user")

	switch user.(type) {
	case models.User:
		return user.(models.User)
	}

	return models.User{}
}

func GetCurrentUserFromGqlContext(ctx context.Context) models.User {
	bc, ok := ctx.Value("BuffaloContext").(buffalo.Context)
	if !ok {
		return models.User{}
	}
	return GetCurrentUser(bc)
}

// ConvertTimeToStringPtr is intended to convert the
// CreatedAt and UpdatedAt fields of database objects
// to pointers to strings to populate the same gqlgen fields
func ConvertTimeToStringPtr(inTime time.Time) *string {
	inTimeStr := inTime.Format(time.RFC3339)
	return &inTimeStr
}

func ConvertStrPtrToString(inPtr *string) string {
	if inPtr == nil {
		return ""
	}

	return *inPtr
}

func GetUuid() uuid2.UUID {
	uuid, _ := uuid2.NewV4()
	return uuid
}

func GetUuidAsString() string {
	return GetUuid().String()
}
