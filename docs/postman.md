Using Postman with Khan
=======================

Khan supports [Postman](https://www.getpostman.com) to make it easier on users to test their Khan server.

Using [Postman](https://www.getpostman.com) with Khan is as simple as importing it's [operations](https://raw.githubusercontent.com/topfreegames/khan/master/postman/operations.postman_collection.json) and [environment](https://raw.githubusercontent.com/topfreegames/khan/master/postman/local.postman_environment.json) into [Postman](https://www.getpostman.com).

## Importing Khan's environment

To import Khan's environment, download it's [environment file](https://raw.githubusercontent.com/topfreegames/khan/master/postman/local.postman_environment.json) and [import it in Postman](https://www.getpostman.com/docs/environments).

## Importing Khan's operations

To import Khan's operations, download it's [operations file](https://raw.githubusercontent.com/topfreegames/khan/master/postman/operations.postman_collection.json) and [import it in Postman](https://www.getpostman.com/docs/collections).

## Running Khan's operations with a different environment

Just configure a new environment and make sure it contains the `baseKhanURL` variable with a value like `http://my-khan-server/`. Do not forget the ending slash or it won't work.
