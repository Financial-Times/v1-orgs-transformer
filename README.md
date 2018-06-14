*DECOMISSIONED*
See [Basic TME Transformer](https://github.com/Financial-Times/basic-tme-transformer) instead

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

* `GET /transformers/organisations`
    * Returns a JSON list of APIURLs to each organisation stored in the transformer cache.
    * A successful GET returns a 200.

* `GET /transformers/organisations/{uuid}` 
    * Get organisation data of the given uuid
    * Returns a 200 if the organisation is found, a 404 if not.

* `GET /transformers/organisations/__ids`
    * Gives a list of JSON objects containing each ID of an organisation
    * A successful GET returns a 200.

* `GET /transformers/organisations/__count`
    * Gives the number of organisations stored in the cache.
    * A successful GET returns a 200.

* `POST /transformers/organisations/__reload`
    * Reloads the information from TME and rebuilds the cache.
    * A successful POST returns a 200.

## Admin endpoints
* Healthcheck - `/__health`
* Ping - `/__ping` or `/ping`
* Build-info - `/__build-info` or `/build-info`
* Good-to-go - `__gtg`




