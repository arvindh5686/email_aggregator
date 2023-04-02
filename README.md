<h2>Email Aggregator</h2>

<h3>Settings</h3>

The following env vars are available to configure sendgrid and mailgun providers:

<b>MailGun</b>
* MAILGUN_BASE_URL - defaulted to https://api.mailgun.net
* MAILGUN_VERSION - defaulted to v3
* MAILGUN_DOMAIN
* MAILGUN_PASSWORD
* MAILGUN_USERNAME

<b>SendGrid</b>

* IS_SENDGRID_DEFAULT - defaulted to true 
* SENDGRID_API_KEY
* SENDGRID_BASE_URL - defaulted to https://api.sendgrid.com
* SENDGRID_VERSION - defaulted to v3

<h3>How to Install</h3>

1. go build cmd/server/main.go
2. ./main

This will start the web server at port 8080. This can be changed by setting the HTTP_PORT env var

<h3>How to run tests</h3>

go test ./...

<h3>Choice of tools</h3>

* Programming language - GO
  * GO is highly performant and is used for building highly scalable systems. It is easy to use and statically typed minimizing chances of errors
* Web framework - GIN
  * Easy to set up, highly performant and supports adding middlewares

<h3>Trade offs</h3>

* The implementation includes an additional `service` layer which acts as the entrypoint for the business logic. While it is possible the HTTP API `handler` layer could have directly used the HTTP client for making requests for simplicity, having the `service` layer will support additional API types in the future like gRPC, GraphQL etc. Some of the validations in the `handler` needs to be moved to the service layer to keep the API layer thin and for maximum code re-use in the `service` layer.

<h3>Future optimizations</h3>

* The database access layer is left unimplemented but can be used for storing input data. In case the service fails to send emails, the stored data can be used for retrying.
* The current implementation uses a flag to switch between sendgrid and mailgun. This can be improved by using circuit breakers and automatically falling back if the current provider is down.
  * This can be done locally i.e within each pod or at the service mesh level. Istio supports setting circuit breaking per service.


