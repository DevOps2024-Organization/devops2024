package flash

// Name of the cookie.
import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

// name of the cookie
const sessionName = "session"

func GetCookieStore() *sessions.CookieStore {
	sessionKey := "token" //a random value to match a key in the session
	return sessions.NewCookieStore([]byte(sessionKey))
}

// Set adds a new message into the cookie storage.
func SetFlash(c *gin.Context, name, value string) {
	session, _ := GetCookieStore().Get(c.Request, sessionName)
	session.AddFlash(value, name)
	session.Save(c.Request, c.Writer)

}

// Get gets flash messages from the cookie storage.
func GetFlash(c *gin.Context, name string) []string {
	session, _ := GetCookieStore().Get(c.Request, sessionName)
	fmt.Println(session.Values)
	flashMessage := session.Flashes(name)
	fmt.Println(flashMessage)
	//if we have some messages
	if len(flashMessage) > 0 {
		session.Save(c.Request, c.Writer)

		//string slice to return messages

		var flashes []string
		for _, f := range flashMessage {
			///add Message to slice
			flashes = append(flashes, f.(string))
		}
		return flashes
	}
	return nil
}
