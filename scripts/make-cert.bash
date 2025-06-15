if [[ -e ./cert/local.weblens.io.crt ]]; then
    exit 0
fi

mkdir -p ./cert

openssl req -x509 -newkey rsa:4096 -sha256 -days 3650 -nodes -keyout ./cert/local.weblens.io.key -out ./cert/local.weblens.io.crt -subj "/CN=local.weblens.io" -addext "subjectAltName=DNS:local.weblens.io,DNS:*.local.weblens.io"
