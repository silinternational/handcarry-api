package notifications

import (
	"fmt"

	"github.com/silinternational/wecarry-api/domain"
)

type MobileService interface {
	Send(msg Message) error
}

type DummyMobileService struct {
	numberSent int
}

type msgTemplate struct {
	subject, body string
}

var templateData = map[string]msgTemplate{
	domain.MessageTemplateNewThreadMessage: {
		subject: "new message", body: "You have a new message.",
	},
	domain.MessageTemplateNewRequest: {
		subject: "new request", body: "There is a new request for an item from your location.",
	},
}

func (t *DummyMobileService) Send(msg Message) error {
	fmt.Printf("new message sent: %s", templateData[msg.Template].subject)
	t.numberSent++
	return nil
}
