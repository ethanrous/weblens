pkill weblens || true

mkdir -p ./build/bin
go build -v -o ./build/bin/weblens ./cmd/weblens/main.go

mkdir -p ~/weblens/core-test
mkdir -p ~/weblens/backup-test

mongosh --eval "use weblens-core-test" --eval "db.dropDatabase()"
mongosh --eval "use weblens-backup-test" --eval "db.dropDatabase()"

ENV_FILE=$(pwd)/config/core-test.env ./build/bin/weblens &

while [ "$(curl -s --location 'http://127.0.0.1:8084/api/info' 2> /dev/null | jq '.started')" != "true" ]
do
  echo "Waiting 500ms for Weblens startup"
  sleep 0.5
done

curl --location 'http://127.0.0.1:8084/api/init' \
--header 'Content-Type: application/json' \
--data '{
    "name": "core-test",
    "role": "core",
    "username": "ethan",
    "password": "password"
}'

token=$(curl --location 'http://127.0.0.1:8084/api/login' \
        --header 'Content-Type: application/json' \
        --data '{
            "username": "ethan",
            "password": "password"
        }' 2> /dev/null | jq -r '.token')

echo $token

apiKey=$(curl --location --request POST 'http://127.0.0.1:8084/api/key' \
--header "Authorization: Bearer $token" 2> /dev/null | jq -r '.key."key"')

echo $apiKey
export CORE_API_KEY=$apiKey
export CORE_ADDRESS="http://127.0.0.1:8084"

#ENV_FILE=$(pwd)/config/backup-test.env ./build/bin/weblens &