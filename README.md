# Feature Flagger API
## _Simple Boolean Based Feature Flag API_

Feature Flagger API is a a bare bones feature flag api that currently only supports boolean based features.

## Features

- List, Get, Create, Update, Delete features by name
- Fast Redis database to store your features

## Installation
Dillinger requires [Go](https://go.dev/doc/install) v1.17+ to run.

Create an appropriate .env with these arguments, then start the server

```
# Values here are for a defualt local dev run
REDIS_DB_HOST="redis"
REDIS_DB_PORT="6379"
REDIS_DB_ID="0"
REDIS_DB_PASSWORD=""
SERVER_PORT=8080
```

```sh
git clone https://github.com/charliemenke/feature-flagger-api.git
cd feature-flagger-api
docker-compose up --build
```

## API Endpoints
---

# List Features

allows caller to see all features in database.

**URL** : `/api/features`

**Method** : `GET`

### Success Responses

**Condition** : Server is up and running properly.

**Code** : `200 OK`

```json
[
    {
        "name": "feature1",
        "enabled": true
    },
    {
        "name": "feature2",
        "enabled": false
    }
]
```

### Error Response

**Condition** : Server is not connected to Redis Databse properly or bad data is in Redis.

**Code** : `500 Internal Server Error`

**Content example** :

```string
Error getting features || Error preparing features to be returned
```
---
# Create Feature

allows caller to create new feature in database.

**URL** : `/api/features`

**Method** : `POST`

**Data constraints**

```json
{
    "name": string *required*,
    "enabled": boolean *optional*
}
```
`enabled` is a optional field, if not supplied the feature will be defaulted to not being enabled

### Success Responses

**Condition** : featureName is supplied and is unique in the database.

**Code** : `200 OK`

```string
Succesfully created feature <featureName>
```

### Error Response
**Condition** : Bad data is sent in the body.

**Code** : `400 Bad Request`

**Content example** :

```string
Error reading request.
```

**Condition** : No name is supplied in the body.

**Code** : `400 Bad Request`

**Content example** :

```string
You must specify feature name.
```

**Condition** : Failure checking database for unique value.

**Code** : `500 Internal Server Error`

**Content example** :

```string
Error checking if feature already exists.
```

**Condition** : Feature with featureName already exists in database.

**Code** : `409 Conflict`

**Content example** :

```string
This feature already exists, please update or delete it instead.
```

**Condition** : Failed to create new feature in database.

**Code** : `500 Internal Server Error`

**Content example** :

```string
Error adding feature: <error> .
```
---
# Get Feature

allows caller to see all features in database.

**URL** : `/api/features/{featureName}`

**Method** : `GET`

### Success Responses

**Condition** : featureName exists in the database.

**Code** : `200 OK`

```json
{
    "name": "feature1",
    "enabled": true
}
```

### Error Response

**Condition** : featureName does not exist in the database.

**Code** : `404 Not Found`

**Content example** :
```string
Feature: '<featureName>' does not exist in the database. Please check the feature name and try again.
```

**Condition** : Failed to check database for feature.

**Code** : `500 Internal Server Error`

**Content example** :
```string
Error getting feature: <error>.
```

**Condition** : Feature to return is malformed.

**Code** : `500 Internal Server Error`

**Content example** :
```string
Error preparing feature to be returned: <error>.
```
---
# Update Feature

allows caller to update enabled status for a feature in database.

**URL** : `/api/features/{featureName}`

**Method** : `PUT`

**Data constraints**

```json
{
    "enabled": boolean *required*
}
```

### Success Responses

**Condition** : featureName exists and was successfully updated.

**Code** : `200 OK`

```string
Succesfully updated feature: <featureName>
```

### Error Response

**Condition** : Bad data is sent in the body.

**Code** : `400 Bad Request`

**Content example** :

```string
Error reading request <error>.
```

**Condition** : No enabled status is supplied in the body.

**Code** : `400 Bad Request`

**Content example** :

```string
Request must supply 'enabled' field.
```

**Condition** : featureName does not exist in the database.

**Code** : `404 Not Found`

**Content example** :
```string
Feature: '<featureName>' does not exist in the database. Please check the feature name and try again.
```

**Condition** : Failed to check database for feature.

**Code** : `500 Internal Server Error`

**Content example** :
```string
Error getting feature: <error>.
```

**Condition** : Failure to update feature in database.

**Code** : `500 Internal Server Error`

**Content example** :
```string
Error updating feature: <error>.
```
---
# Delete Feature

allows caller to delete feature in database.

**URL** : `/api/features/{featureName}`

**Method** : `DELETE`

### Success Responses

**Condition** : featureName exists and was successfully updated.

**Code** : `200 OK`

```string
Succesfully deleted key: <featureName>
```

### Error Response

**Condition** : Failure to delete feature in database.

**Code** : `500 Internal Server Error`

**Content example** :

```string
Error deleting feature: <error>.
```

**Condition** : No feature with featureName to delete in database.

**Code** : `404 Bad Request`

**Content example** :

```string
Could not find feature to delete: <featureName>.
```