# Authentication

Rel uses JWT tokens stored in http-only cookies to tell if a browser is authenticated or not. The cookie is named `accesstoken` by default, but its name can be changed through the `jwt_cookie_name` configuration option.

Whenever a request is made to the server, the accesstoken cookie is searched for and decoded if it exists, looking for the `role` key at the root of the object, which _should_ contain a real user role.

## OAuth

OAuth's callback are at `/oauth/{provider}/callback`

## SAML

### Certificates

SAML's callback are at `/saml/{provider}/callback`

## User / Password
