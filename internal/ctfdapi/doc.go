// Package ctfdapi provides access to the CTFd REST API for automating the CTFd instance. As the OpenAPI spec provided
// by CTFd at https://demo.ctfd.io/api/v1/swagger.json is incomplete (the payload for most requests is missing), we
// can not auto-generate the REST client from the API spec. We therefore have to handcraft every endpoint we want to
// automate.
package ctfdapi
