set +e
set -x

pkill weblens || true

export APP_ROOT=$(pwd)

mkdir -p ./build/bin
go build -v -o ./build/bin/weblens ./cmd/weblens/main.go

mkdir -p ~/weblens/core-test
mkdir -p ~/weblens/backup-test

mongosh --eval "use weblens-core-test" --eval "db.dropDatabase()"
mongosh --eval "use weblens-backup-test" --eval "db.dropDatabase()"

ENV_FILE=$(pwd)/config/core-test.env ./build/bin/weblens &
echo $?

counter=0
while [ "$(curl -s --location 'http://localhost:8089/api/info' 2> /dev/null | jq '.started')" != "true" ]
do
  echo "Waiting 500ms for Weblens startup"
  sleep 0.5
  ((counter++))
  if [ $counter -ge 20 ]; then
    echo "Failed to connect to weblens core after 10 seconds, exiting..."
    return 1
  fi
done

curl --location 'http://localhost:8089/api/init' \
--header 'Content-Type: application/json' \
--data '{
    "name": "core-test",
    "role": "core",
    "username": "ethan",
    "password": "password"
}' 2> /dev/null
echo "CURL STATUS $?"

token=$(curl --location 'http://localhost:8089/api/login' \
        --header 'Content-Type: application/json' \
        --data '{
            "username": "ethan",
            "password": "password"
        }' 2> /dev/null | jq -r '.token')

echo $token

apiKey=$(curl --location --request POST 'http://localhost:8089/api/key' \
--header "Authorization: Bearer $token" 2> /dev/null | jq -r '.key."key"')

echo $apiKey
export CORE_API_KEY=$apiKey
export CORE_ADDRESS="http://localhost:8089"

#ENV_FILE=$(pwd)/config/backup-test.env ./build/bin/weblens &