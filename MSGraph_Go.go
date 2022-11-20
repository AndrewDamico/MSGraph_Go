package main

import (
	"database/sql"
	"fmt"
	//"github.com/cjlapao/common-go/identity"
	//"github.com/microsoftgraph/msgraph-sdk-go/models"
	"MSGraph_Go/graphhelper"
	"log"
	"os"
	//"bufio"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"strconv"
	// "strings"
)

func main() {

	fmt.Println("Go Graph Tutorial")
	fmt.Println()

	// Load .env files
	// .env.local takes precedence (if present)
	godotenv.Load(".env.local")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}
	fmt.Println("Loading Graph Helper")
	fmt.Println()

	graphHelper := graphhelper.NewGraphHelper()
	fmt.Println("Initializing Graph Helper")
	fmt.Println()

	initializeGraph(graphHelper)

	fmt.Println("Graph Helper Initialized")
	fmt.Println()

	var choice int64 = -1

	for {
		fmt.Println("Please choose one of the following options:")
		fmt.Println("0. Exit")
		fmt.Println("1. Sync Events")
		fmt.Println("2. Get Calendars")

		_, err = fmt.Scanf("%d", &choice)
		if err != nil {
			choice = -1
		}

		switch choice {
		case 0:
			// Exit the program
			fmt.Println("Goodbye...")
		case 2:
			// Run any Graph code
			GetCalendars(graphHelper)
		case 1:
			// Run any Graph code
			SyncEvents(graphHelper)
		default:
			fmt.Println("Invalid choice! Please try again.")
		}

		if choice == 0 {
			break
		}
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

func initializeGraph(graphHelper *graphhelper.GraphHelper) {
	err := graphHelper.InitializeGraphForUserAuth()
	if err != nil {
		log.Panicf("Error initializing Graph for user auth: %v\n", err)
	}
}

func GetCalendars(graphHelper *graphhelper.GraphHelper) {
	items, err := graphHelper.GetCalendars()
	if err != nil {
		log.Panicf("Error making Graph call: %v", err)
	}

	for _, items := range items.GetValue() {
		fmt.Printf("Events: %s\n", *items.GetName())
	}
}

func SyncEvents(graphHelper *graphhelper.GraphHelper) {
	items, err := graphHelper.GetEvents()
	if err != nil {
		log.Panicf("Error making Graph call: %v", err)
	}

	for _, items := range items.GetValue() {

		// sync_calendar(*items.GetID(), *items)
		id := *items.GetICalUId()
		subject := *items.GetSubject()
		body := *items.GetBody().GetContent()
		// fmt.Printf("Events: %s\n", *items.GetCategories())
		changekey := *items.GetChangeKey()
		organizer := *items.GetOrganizer().GetEmailAddress().GetName()
		starttime := *items.GetStart().GetDateTime()
		endtime := *items.GetEnd().GetDateTime()
		// fmt.Printf("All Day: %s\n", *items.GetIsAllDay())
		//show_as := *items.GetShowAs()

		sync_calendar(id, subject, body, changekey, organizer, starttime, endtime)
	}
}

func sync_calendar(
	id string,
	subject string,
	body string,
	changekey string,
	organizer string,
	starttime string,
	endtime string,
	// allday *string,
	// showas string
) error {

	host := os.Getenv("A2DAM_HOST")
	port := getEnvAsInt("A2DAM_PORT", 1)
	user := os.Getenv("A2DAM_USER")
	password := os.Getenv("A2DAM_PASSWORD")
	dbname := os.Getenv("A2DAM_DBNAME")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	sqlStatement := `
INSERT INTO "msGraph_outlookevent" (id, subject, body, change_key, organizer, start_time, end_time, show_as)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO UPDATE
    SET subject = excluded.subject,
    body = excluded.body,
    change_key = excluded.change_key,
    organizer = excluded.organizer,
    start_time = excluded.start_time,
    end_time = excluded.end_time,
    show_as = excluded.show_as
RETURNING id`

	err = db.QueryRow(sqlStatement, id, subject, body, changekey, organizer, starttime, endtime, "default").Scan(&id)
	if err != nil {
		panic(err)
	}
	fmt.Println("Event Created", subject)

	return nil
}
