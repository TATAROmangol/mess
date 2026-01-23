export ACCESS_TOKEN=$( curl -s -X POST \
  http://localhost:7070/realms/main/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password&client_id=main&client_secret=main&username=test&password=test" \
| jq -r '.access_token')

wscat -c ws://localhost:8082/ws -H "Authorization: Bearer $ACCESS_TOKEN"