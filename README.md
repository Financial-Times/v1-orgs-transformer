# v1-orgs-transformer

[![Circle CI](https://circleci.com/gh/Financial-Times/v1-orgs-transformer/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/v1-orgs-transformer/tree/master)

Retrieves Organisations taxonomy from TME and transforms the organisations to the internal UP json model.
The service exposes endpoints for getting all the organisations and for getting organisation by uuid.

# Usage
`go get github.com/Financial-Times/v1-orgs-transformer`

`$GOPATH/bin/v1-orgs-transformer --port=8080 --base-url="http://localhost:8080/transformers/organisations/" --tme-base-url="https://tme-live.internal.ft.com:40001" --tme-username="user" --tme-password="pass"`
```
export|set PORT=8080
export|set BASE_URL="http://localhost:8080/transformers/organisations/"
export|set TME_BASE_URL="https://tme-live.internal.ft.com:40001"
export|set TME_USERNAME="user"
export|set TME_PASSWORD="pass"
$GOPATH/bin/v1-orgs-transformer
```

With Docker:

`docker build -t coco/v1-orgs-transformer .`

`docker run -ti --env BASE_URL=<base url> --env TME_BASE_URL=<structure service url> --env TME_USERNAME=<user> --env TME_PASSWORD=<pass> coco/v1-orgs-transformer`
