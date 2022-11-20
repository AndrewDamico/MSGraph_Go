package graphhelper

import (
	"context"
	"fmt"
	"github.com/microsoftgraph/msgraph-sdk-go/me/events"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	auth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/me"
	"github.com/microsoftgraph/msgraph-sdk-go/me/calendars"

	"github.com/microsoftgraph/msgraph-sdk-go/users"
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
	token, err := g.deviceCodeCredential.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: g.graphUserScopes,
	})
	if err != nil {
		return nil, err
	}

	return &token.Token, nil
}

// Fetches the authorized users name and email

func (g *GraphHelper) GetUser() (models.Userable, error) {
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

// Fetches list of all users

func (g *GraphHelper) GetUsers() (models.UserCollectionResponseable, error) {

	err := g.EnsureGraphForAppOnlyAuth()
	if err != nil {
		return nil, err
	} // make sure that authorization obtained; if not, then request login

	var topValue int32 = 25 //Fetches top 25

	query := users.UsersRequestBuilderGetQueryParameters{
		// Only request specific properties
		Select: []string{"displayName", "id", "mail"},
		// Get at most 25 results
		Top: &topValue,
		// Sort by display name
		Orderby: []string{"displayName"},
	}

	// API: result, err := graphClient.Users().Get(context.Background(), nil)

	return g.appClient.Users().
		Get(context.Background(),
			&users.UsersRequestBuilderGetRequestConfiguration{
				QueryParameters: &query,
			})
}

func (g *GraphHelper) GetCalendars() (models.CalendarCollectionResponseable, error) {

	//err := g.EnsureGraphForAppOnlyAuth()
	//if err != nil { return nil, err }

	//var topValue int32 = 25 //Fetches top 25

	query := calendars.CalendarsRequestBuilderGetQueryParameters{
		// Only request specific properties
		Select: []string{"name", "owner"},
		// Get at most 25 results
		//Top: &topValue,
		// Sort by display name
		//Orderby: []string{"name"},
	}

	return g.userClient.Me().
		Calendars().
		Get(context.Background(),
			&calendars.CalendarsRequestBuilderGetRequestConfiguration{
				QueryParameters: &query,
			})
}

func (g *GraphHelper) GetEvents() (models.EventCollectionResponseable, error) {

	//err := g.EnsureGraphForAppOnlyAuth()
	//if err != nil { return nil, err }

	//var topValue int32 = 25 //Fetches top 25
	headers := map[string]string{
		"Prefer": "outlook.timezone=\"Central Standard Time\"",
	}

	query := events.EventsRequestBuilderGetQueryParameters{
		// Only request specific properties
		Select: []string{"iCalUId", "subject", "body", "bodyPreview", "categories", "changeKey", "organizer", "attendees", "start", "end", "location", "isAllDay", "showAs"},
		// Get at most 25 results
		//Top: &topValue,
		// Sort by display name
		//Orderby: []string{"name"},
	}

	return g.userClient.Me().
		Events().
		Get(context.Background(),
			&events.EventsRequestBuilderGetRequestConfiguration{
				Headers:         headers,
				QueryParameters: &query,
			})
}
