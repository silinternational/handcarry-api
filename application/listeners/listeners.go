package listeners

import (
	"time"

	"github.com/gobuffalo/envy"

	"github.com/gobuffalo/events"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

const (
	UserAccessTokensCleanupDelayMinutes = 480
)

var UserAccessTokensNextCleanupTime time.Time

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
	domain.EventApiUserCreated: []apiListener{
		{
			name:     "user-created",
			listener: userCreated,
		},
	},

	domain.EventApiAuthUserLoggedIn: []apiListener{
		{
			name:     "trigger-user-access-tokens-cleanup",
			listener: userAccessTokensCleanup,
		},
	},

	domain.EventApiMessageCreated: []apiListener{
		{
			name:     "send-new-message-notification",
			listener: sendNewMessageNotification,
		},
	},

	domain.EventApiPostStatusUpdated: []apiListener{
		{
			name:     "post-status-updated-notification",
			listener: sendPostStatusUpdatedNotification,
		},
	},
}

// RegisterListeners registers all the listeners to be used by the app
func RegisterListeners() {
	for _, listeners := range apiListeners {
		for _, l := range listeners {
			_, err := events.NamedListen(l.name, l.listener)
			if err != nil {
				domain.ErrLogger.Print("Failed registering listener: " + l.name)
			}
		}
	}
}

func userAccessTokensCleanup(e events.Event) {
	if e.Kind != domain.EventApiAuthUserLoggedIn {
		return
	}

	now := time.Now()
	if !now.After(UserAccessTokensNextCleanupTime) {
		return
	}

	UserAccessTokensNextCleanupTime = now.Add(time.Duration(time.Minute * UserAccessTokensCleanupDelayMinutes))

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

	domain.Logger.Printf("%s User Created ... %s", domain.GetCurrentTime(), e.Message)
}

func sendNewMessageNotification(e events.Event) {
	if e.Kind != domain.EventApiMessageCreated {
		return
	}

	domain.Logger.Printf("%s Message Created ... %s", domain.GetCurrentTime(), e.Message)

	mEData, ok := e.Payload["eventData"].(models.MessageCreatedEventData)
	if !ok {
		domain.ErrLogger.Print("unable to parse Message Created event payload")
		return
	}

	uiUrl := envy.Get(domain.UIURLEnv, "")
	data := map[string]interface{}{
		"postURL":        uiUrl + "/#/requests/" + mEData.PostUUID,
		"postTitle":      mEData.PostTitle,
		"messageContent": mEData.MessageContent,
		"sentByNickname": mEData.MessageCreatorNickName,
		"threadURL":      uiUrl + "/#/messages/" + mEData.ThreadUUID,
	}

	for _, r := range mEData.MessageRecipients {
		msg := notifications.Message{
			Template:  domain.MessageTemplateNewMessage,
			Data:      data,
			FromName:  mEData.MessageCreatorNickName,
			FromEmail: mEData.MessageCreatorEmail,
			ToName:    r.Nickname,
			ToEmail:   r.Email,
		}
		if err := notifications.Send(msg); err != nil {
			domain.ErrLogger.Printf("error sending 'New Message' notification, %s", err)
		}
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

	if pEData.Post.Type != models.PostTypeRequest {
		return
	}

	switch pEData.NewStatus {
	case models.PostStatusOpen:
		RequestNewStatusOpen(pEData)
	case models.PostStatusCommitted:
		RequestNewStatusCommitted(pEData)
	case models.PostStatusAccepted:
		RequestNewStatusAccepted(pEData)
	case models.PostStatusDelivered:
		RequestNewStatusDelivered(pEData)
	case models.PostStatusReceived:
		RequestNewStatusReceived(pEData)
	case models.PostStatusCompleted:
		RequestStatusCompleted(pEData)
	case models.PostStatusRemoved:
		RequestNewStatusRemoved(pEData)
	}
}
