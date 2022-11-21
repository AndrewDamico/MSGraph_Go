# MSGraph_Go

The MSGraph_Go package connects with the MS Graph API using Golang to fetch all calendar events for a user and saves them to a local django PostgreSQL database.

Currently an .env file must be placed in the both the directory with the go.mod file, as well as the parent directory of the built .so or .dll file, with the following variables:

* CLIENT_ID={value}  
* CLIENT_SECRET={value}  
* TENANT_ID={value}  
* AUTH_TENANT={value}  
* GRAPH_USER_SCOPES={value}  
* USER_ID={value}  