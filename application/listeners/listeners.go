package listeners

import (
	"errors"
	"time"

	"github.com/gobuffalo/events"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/job"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

const (
	userAccessTokensCleanupDelayMinutes = 480
)

var userAccessTokensNextCleanupAfter time.Time

type apiListener struct {
	name     string
	listener func(events.Event)
}

//
// Register new listener functions here.  Remember, though, that these groupings just
// describe what we want.  They don't make it happen this way. The listeners
// themselves still need to verify the event kind
//
var apiListeners = map[string][]apiListener{
	domain.EventApiUserCreated: {
		{
			name:     "user-created",
			listener: userCreated,
		},
	},

	domain.EventApiAuthUserLoggedIn: {
		{
			name:     "trigger-user-access-tokens-cleanup",
			listener: userAccessTokensCleanup,
		},
	},

	domain.EventApiMessageCreated: {
		{
			name:     "send-new-message-notification",
			listener: sendNewThreadMessageNotification,
		},
	},

	domain.EventApiPostStatusUpdated: {
		{
			name:     "post-status-updated-notification",
			listener: sendPostStatusUpdatedNotification,
		},
	},

	domain.EventApiPostCreated: {
		{
			name:     "post-created-notification",
			listener: sendPostCreatedNotifications,
		},
	},

	domain.EventApiPotentialProviderCreated: {
		{
			name:     "potentialprovider-created-notification",
			listener: potentialProviderCreated,
		},
	},

	domain.EventApiPotentialProviderSelfDestroyed: {
		{
			name:     "potentialprovider-self-destroyed-notification",
			listener: potentialProviderSelfDestroyed,
		},
	},

	domain.EventApiPotentialProviderRejected: {
		{
			name:     "potentialprovider-rejected-notification",
			listener: potentialProviderRejected,
		},
	},
}

// RegisterListeners registers all the listeners to be used by the app
func RegisterListeners() {
	for _, listeners := range apiListeners {
		for _, l := range listeners {
			_, err := events.NamedListen(l.name, l.listener)
			if err != nil {
				domain.ErrLogger.Print("Failed registering listener:", l.name, err)
			}
		}
	}
}

func userAccessTokensCleanup(e events.Event) {
	if e.Kind != domain.EventApiAuthUserLoggedIn {
		return
	}

	now := time.Now()
	if !now.After(userAccessTokensNextCleanupAfter) {
		return
	}

	userAccessTokensNextCleanupAfter = now.Add(time.Duration(time.Minute * userAccessTokensCleanupDelayMinutes))

	var uats models.UserAccessTokens
	deleted, err := uats.DeleteExpired()
	if err != nil {
		domain.ErrLogger.Printf("%s Last error deleting expired user access tokens during cleanup ... %v",
			domain.GetCurrentTime(), err)
	}

	domain.Logger.Printf("%s Deleted %v expired user access tokens during cleanup", domain.GetCurrentTime(), deleted)
}

func userCreated(e events.Event) {
	if e.Kind != domain.EventApiUserCreated {
		return
	}

	user, ok := e.Payload["user"].(*models.User)
	if !ok {
		domain.Logger.Printf("Failed to get User from event payload for notification. Event message: %s", e.Message)
		return
	}

	domain.Logger.Printf("User Created: %s", e.Message)

	if err := sendNewUserWelcome(*user); err != nil {
		domain.Logger.Printf("Failed to send new user welcome to %s. Error: %s",
			user.UUID.String(), err)
	}
}

func sendNewThreadMessageNotification(e events.Event) {
	if e.Kind != domain.EventApiMessageCreated {
		return
	}

	domain.Logger.Printf("%s Thread Message Created ... %s", domain.GetCurrentTime(), e.Message)

	id, ok := e.Payload[domain.ArgMessageID].(int)
	if !ok {
		domain.ErrLogger.Print("sendNewThreadMessageNotification: unable to read message ID from event payload")
		return
	}

	if err := job.SubmitDelayed(job.NewThreadMessage, domain.NewMessageNotificationDelay,
		map[string]interface{}{domain.ArgMessageID: id}); err != nil {
		domain.ErrLogger.Printf("error starting 'New Message' job, %s", err)
	}
}

func sendPostStatusUpdatedNotification(e events.Event) {
	if e.Kind != domain.EventApiPostStatusUpdated {
		return
	}

	pEData, ok := e.Payload["eventData"].(models.PostStatusEventData)
	if !ok {
		domain.ErrLogger.Print("unable to parse Post Status Updated event payload")
		return
	}

	pid := pEData.PostID

	post := models.Post{}
	if err := post.FindByID(pid); err != nil {
		domain.ErrLogger.Printf("unable to find post from event with id %v ... %s", pid, err)
	}

	if post.Type != models.PostTypeRequest {
		return
	}

	requestStatusUpdatedNotifications(post, pEData)
}

func sendPostCreatedNotifications(e events.Event) {
	if e.Kind != domain.EventApiPostCreated {
		return
	}

	eventData, ok := e.Payload["eventData"].(models.PostCreatedEventData)
	if !ok {
		domain.ErrLogger.Printf("Post Created event payload incorrect type: %T", e.Payload["eventData"])
		return
	}

	var post models.Post
	if err := post.FindByID(eventData.PostID); err != nil {
		domain.ErrLogger.Printf("unable to find post %d from post-created event, %s", eventData.PostID, err)
	}

	users, err := post.GetAudience()
	if err != nil {
		domain.ErrLogger.Print("unable to get post audience in event listener, ", err.Error())
		return
	}

	sendNewPostNotifications(post, users)
}

func potentialProviderCreated(e events.Event) {
	if e.Kind != domain.EventApiPotentialProviderCreated {
		return
	}

	eventData, ok := e.Payload["eventData"].(models.PotentialProviderEventData)
	if !ok {
		domain.ErrLogger.Printf("PotentialProvider event payload incorrect type: %T", e.Payload["eventData"])
		return
	}

	var potentialProvider models.User
	if err := potentialProvider.FindByID(eventData.UserID); err != nil {
		domain.ErrLogger.Printf("unable to find PotentialProvider User %d, %s", eventData.UserID, err)
	}

	var post models.Post
	if err := post.FindByID(eventData.PostID); err != nil {
		domain.ErrLogger.Printf("unable to find post %d from PotentialProvider event, %s", eventData.PostID, err)
	}

	creator := post.CreatedBy

	sendPotentialProviderCreatedNotification(potentialProvider.Nickname, creator, post)
}

func potentialProviderSelfDestroyed(e events.Event) {
	if e.Kind != domain.EventApiPotentialProviderSelfDestroyed {
		return
	}

	eventData, ok := e.Payload["eventData"].(models.PotentialProviderEventData)
	if !ok {
		domain.ErrLogger.Printf("PotentialProvider event payload incorrect type: %T", e.Payload["eventData"])
		return
	}

	var potentialProvider models.User
	if err := potentialProvider.FindByID(eventData.UserID); err != nil {
		domain.ErrLogger.Printf("unable to find PotentialProvider User %d, %s", eventData.UserID, err)
	}

	var post models.Post
	if err := post.FindByID(eventData.PostID); err != nil {
		domain.ErrLogger.Printf("unable to find post %d from PotentialProvider event, %s", eventData.PostID, err)
	}

	creator := post.CreatedBy

	sendPotentialProviderSelfDestroyedNotification(potentialProvider.Nickname, creator, post)
}

func potentialProviderRejected(e events.Event) {
	if e.Kind != domain.EventApiPotentialProviderRejected {
		return
	}

	eventData, ok := e.Payload["eventData"].(models.PotentialProviderEventData)
	if !ok {
		domain.ErrLogger.Printf("PotentialProvider event payload incorrect type: %T", e.Payload["eventData"])
		return
	}

	var potentialProvider models.User
	if err := potentialProvider.FindByID(eventData.UserID); err != nil {
		domain.ErrLogger.Printf("unable to find PotentialProvider User %d, %s", eventData.UserID, err)
	}

	var post models.Post
	if err := post.FindByID(eventData.PostID); err != nil {
		domain.ErrLogger.Printf("unable to find post %d from PotentialProvider event, %s", eventData.PostID, err)
	}

	creator := post.CreatedBy.Nickname

	sendPotentialProviderRejectedNotification(potentialProvider, creator, post)
}

func sendNewUserWelcome(user models.User) error {
	if user.Email == "" {
		return errors.New("'To' email address is required")
	}

	language := user.GetLanguagePreference()
	subject := domain.GetTranslatedSubject(language, "Email.Subject.Welcome", map[string]string{})

	msg := notifications.Message{
		Template:  domain.MessageTemplateNewUserWelcome,
		ToName:    user.GetRealName(),
		ToEmail:   user.Email,
		FromEmail: domain.EmailFromAddress(nil),
		Subject:   subject,
		Data: map[string]interface{}{
			"appName":      domain.Env.AppName,
			"uiURL":        domain.Env.UIURL,
			"supportEmail": domain.Env.SupportEmail,
			"userEmail":    user.Email,
			"firstName":    user.FirstName,
		},
	}
	return notifications.Send(msg)
}
