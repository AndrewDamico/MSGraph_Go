package graphhelper

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	auth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/me"
	//"github.com/microsoftgraph/msgraph-sdk-go/me/calendars"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"os"
	"strings"

	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/microsoftgraph/msgraph-sdk-go/users/item/calendar/events"
	"github.com/microsoftgraph/msgraph-sdk-go/users/item/calendars"
)

type GraphHelper struct {
	deviceCodeCredential   *azidentity.DeviceCodeCredential
	userClient             *msgraphsdk.GraphServiceClient
	graphUserScopes        []string
	clientSecretCredential *azidentity.ClientSecretCredential
	appClient              *msgraphsdk.GraphServiceClient
}

func NewGraphHelper() *GraphHelper {
	g := &GraphHelper{}

	return g
}

//Create a graph client instance

func (g *GraphHelper) InitializeGraphForUserAuth() error {
	// Creates an authorization for user. Not currently used as all usable functions are application level.
	clientId := os.Getenv("CLIENT_ID")
	authTenant := os.Getenv("AUTH_TENANT")
	scopes := os.Getenv("GRAPH_USER_SCOPES")
	g.graphUserScopes = strings.Split(scopes, ",")

	// Create the device code credential
	credential, err := azidentity.NewDeviceCodeCredential(&azidentity.DeviceCodeCredentialOptions{
		ClientID: clientId,
		TenantID: authTenant,
		UserPrompt: func(ctx context.Context, message azidentity.DeviceCodeMessage) error {
			fmt.Println(message.Message)
			return nil
		},
	})
	if err != nil {
		return err
	}

	g.deviceCodeCredential = credential

	// Create an auth provider using the credential
	authProvider, err := auth.NewAzureIdentityAuthenticationProviderWithScopes(credential, g.graphUserScopes)
	if err != nil {
		return err
	}

	// Create a request adapter using the auth provider
	adapter, err := msgraphsdk.NewGraphRequestAdapter(authProvider)
	if err != nil {
		return err
	}

	// Create a Graph client using request adapter
	client := msgraphsdk.NewGraphServiceClient(adapter)
	g.userClient = client

	return nil
}

func (g *GraphHelper) GetUserToken() (*string, error) {
	// Fetches the token for the logged in user.
	token, err := g.deviceCodeCredential.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: g.graphUserScopes,
	})
	if err != nil {
		return nil, err
	}

	return &token.Token, nil
}

func (g *GraphHelper) GetUser() (models.Userable, error) {
	// Fetches the authorized users name and email
	query := me.MeRequestBuilderGetQueryParameters{
		// Only request specific properties
		Select: []string{"displayName", "mail", "userPrincipalName"},
	}

	return g.userClient.Me().Get(context.Background(),
		&me.MeRequestBuilderGetRequestConfiguration{
			QueryParameters: &query,
		})
}

func (g *GraphHelper) EnsureGraphForAppOnlyAuth() error {
	// Checks that application permissions provisioned
	if g.clientSecretCredential == nil {
		clientId := os.Getenv("CLIENT_ID")
		tenantId := os.Getenv("TENANT_ID")
		clientSecret := os.Getenv("CLIENT_SECRET")
		credential, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, nil)
		if err != nil {
			return err
		}

		g.clientSecretCredential = credential
	}

	if g.appClient == nil {
		// Create an auth provider using the credential
		authProvider, err := auth.NewAzureIdentityAuthenticationProviderWithScopes(g.clientSecretCredential, []string{
			"https://graph.microsoft.com/.default",
		})
		if err != nil {
			return err
		}

		// Create a request adapter using the auth provider
		adapter, err := msgraphsdk.NewGraphRequestAdapter(authProvider)
		if err != nil {
			return err
		}

		// Create a Graph client using request adapter
		client := msgraphsdk.NewGraphServiceClient(adapter)
		g.appClient = client
	}

	return nil
}

func (g *GraphHelper) GetUsers() (models.UserCollectionResponseable, error) {
	// Fetches list of all users

	// make sure that authorization obtained; if not, then request login
	err := g.EnsureGraphForAppOnlyAuth()
	if err != nil {
		return nil, err
	}

	var topValue int32 = 25 //Fetches top 25

	query := users.UsersRequestBuilderGetQueryParameters{
		// Only request specific properties
		Select: []string{"displayName", "id", "mail"},
		// Get at most 25 results
		Top: &topValue,
		// Sort by display name
		Orderby: []string{"displayName"},
	}

	return g.appClient.Users().
		Get(context.Background(),
			&users.UsersRequestBuilderGetRequestConfiguration{
				QueryParameters: &query,
			})
}

func (g *GraphHelper) GetCalendars(user string) (models.CalendarCollectionResponseable, error) {
	//Fetches all calendars for user in .env

	err := g.EnsureGraphForAppOnlyAuth()
	if err != nil {
		return nil, err
	}

	// requests list of calendars for user provided in request
	query := calendars.CalendarsRequestBuilderGetQueryParameters{
		// Only request specific properties
		Select: []string{"name", "owner"},
	}

	return g.
		// previously 'userClient.Me().' requested calendar based on user level auth.
		appClient.
		UsersById(user).
		Calendars().
		Get(context.Background(),
			&calendars.CalendarsRequestBuilderGetRequestConfiguration{
				QueryParameters: &query,
			})
}

func (g *GraphHelper) GetEvents(user string) (models.EventCollectionResponseable, error) {
	// Fetches all events for user's default calendar
	err := g.EnsureGraphForAppOnlyAuth()
	if err != nil {
		return nil, err
	}

	// requests that body be returned in text, not html
	headers := map[string]string{
		"Prefer":
		//"outlook.timezone=\"America/Chicago\", //PostgreSQL saves time in UTC
		"outlook.body-content-type=\"text\"",
	}

	query := events.EventsRequestBuilderGetQueryParameters{
		//query := events.EventsRequestBuilderGetQueryParameters{
		// Only request specific properties
		Select: []string{
			"iCalUId", "subject", "body", "bodyPreview",
			"categories", "changeKey", "organizer", "attendees",
			"start", "end", "location", "isAllDay", "showAs",
		},
	}

	return g.appClient.
		UsersById(user). //.userClient.Me().
		Calendar().
		Events().
		Get(context.Background(),
			//&events.EventsRequestBuilderGetRequestConfiguration{
			&events.EventsRequestBuilderGetRequestConfiguration{
				Headers:         headers,
				QueryParameters: &query,
			})
}
