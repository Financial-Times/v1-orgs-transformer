# v1-orgs-transformer

[![CircleCI](https://circleci.com/gh/Financial-Times/v1-orgs-transformer.svg?style=svg)](https://circleci.com/gh/Financial-Times/v1-orgs-transformer) [![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/v1-orgs-transformer)](https://goreportcard.com/report/github.com/Financial-Times/v1-orgs-transformer) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/v1-orgs-transformer/badge.svg?branch=master)](https://coveralls.io/github/Financial-Times/v1-orgs-transformer?branch=master)

Retrieves Organisations taxonomy from TME and transforms the organisations to the internal UP json model.
The service exposes endpoints for getting all the organisations and for getting organisation by uuid.

# Usage
`go get -u github.com/Financial-Times/v1-orgs-transformer`

`$GOPATH/bin/v1-orgs-transformer --port=8080 --base-url="http://localhost:8080/transformers/organisations/" --tme-base-url="https://tme.ft.com" --tme-username="user" --tme-password="pass" --token="token"`

```
export|set PORT=8080
export|set BASE_URL="http://localhost:8080/transformers/organisations/"
export|set TME_BASE_URL="https://tme.ft.com"
export|set TME_USERNAME="user"
export|set TME_PASSWORD="pass"
export|set TOKEN="token"
export|set CACHE_FILE_NAME="cache.db"
$GOPATH/bin/v1-orgs-transformer
```

### With Docker:

`docker build -t coco/v1-orgs-transformer .`

`docker run -ti --env BASE_URL=<base url> --env TME_BASE_URL=<structure service url> --env TME_USERNAME=<user> --env TME_PASSWORD=<pass> --env TOKEN=<token> --env CACHE_FILE_NAME=<file> coco/v1-orgs-transformer`

# Endpoints

* `/transformers/organisations` - Get all organisations as APIURLs
* `/transformers/organisations/{uuid}` - Get organisation data of the given uuid