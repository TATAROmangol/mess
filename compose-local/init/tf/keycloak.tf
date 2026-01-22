//realms
resource "keycloak_realm" "realm-main" {
  realm   = "main"
  enabled = true
}

//clients
resource "keycloak_openid_client" "client-main" {
  realm_id  = keycloak_realm.realm-main.id
  client_id = "main"
  client_secret = "main"

  name      = "from e2e tests"
  enabled   = true

  access_type = "CONFIDENTIAL"
  standard_flow_enabled = true
  direct_access_grants_enabled = true
  service_accounts_enabled = true

  valid_redirect_uris = [
    "*"
  ]
}

//users
resource "keycloak_user" "user" {
  realm_id = keycloak_realm.realm-main.id

  username = "main"
  enabled  = true

  first_name = "main"
  last_name  = "main"
  email      = "main@main.main"

  initial_password {
    value     = "main"
    temporary = false
  }
}
resource "keycloak_user" "user-2" {
  realm_id = keycloak_realm.realm-main.id

  username = "test"
  enabled  = true

  first_name = "test"
  last_name  = "test"
  email      = "test@test.test"

  initial_password {
    value     = "test"
    temporary = false
  }
}

//vars
variable "keycloak_url" {
  type = string
}
variable "keycloak_user" {
  type = string
}
variable "keycloak_password" {
  type      = string
  sensitive = true
}
