Khan API
========

## Healthcheck Routes

  ### Healthcheck

  `GET /healthcheck`

  Validates that the app is still up, including the database connection.

  * Success Response
    * Code: `200`
    * Content:

      ```
        "WORKING"
      ```

    * Headers:

      It will add an `KHAN-VERSION` header with the current khan module version.

  * Error Response

    It will return an error if it failed to connect to the database.

    * Code: `500`
    * Content:

      ```
        "Error connecting to database: <error-details>"
      ```

## Game Routes

  ### Create Game
  `POST /games`

  Creates a new game with the given parameters.

  * Payload

    ```
    {
    	"publicID":                      [string],  // 36 characters max, must be unique
    	"name":                          [string],  // 2000 characters max
    	"metadata":                      [JSON],
    	"minMembershipLevel":            [int],
    	"maxMembershipLevel":            [int],
    	"minLevelToAcceptApplication":   [int],     
    	"minLevelToCreateInvitation":    [int],
    	"minLevelToRemoveMember":        [int],
      "MinLevelOffsetToRemoveMember":  [int],
    	"minLevelOffsetToPromoteMember": [int],
    	"minLevelOffsetToDemoteMember":  [int],
    	"maxMembers":                    [int]
  	}
    ```

    Metadata is intended to include all game specific configuration that is not directly related to the khan's clan management. For example, clank ranking, clan trophies, clan description, etc.

    * Some parameters require special attention:

      **minMembershipLevel**: Lowest membership level a clan member can have, if the member reaches this level they cannot be demoted.

      **maxMembershipLevel**: Highest membership level a clan member can have, if the member reaches this level they cannot be promoted.

      **minLevelToAcceptApplication**: A member cannot accept a player's application to join the clan unless their level is greater or equal to this parameter.

      **minLevelToCreateInvitation**: A member cannot invite a player to join the clan unless their level is greater or equal to this parameter.

      **MinLevelOffsetToRemoveMember**: A member cannot remove another member unless their level is at least `MinLevelOffsetToRemoveMember` levels greater than the level of the member they wish to promote.

      **minLevelOffsetToPromoteMember**: A member cannot promote another member unless their level is at least `minLevelOffsetToPromoteMember` levels greater than the level of the member they wish to promote.

      **minLevelOffsetToDemoteMember**: A member cannot demote another member unless their level is at least `minLevelOffsetToDemoteMember` levels greater than the level of the member they wish to demote.

      **maxMembers**: Maximum number of members a clan of this game can have.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "publicID": [string]  // game public id
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    It will return an error if there are invalid parameters.

    * Code: `422`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Update Game
  `PUT /games/:gameID`

  Updates the game with that has publicID `gameID`.

  * Payload

    ```
    {
    	"name":                          [string],  // 2000 characters max
    	"metadata":                      [JSON],
    	"minMembershipLevel":            [int],
    	"maxMembershipLevel":            [int],
    	"minLevelToAcceptApplication":   [int],
    	"minLevelToCreateInvitation":    [int],
    	"minLevelToRemoveMember":        [int],
    	"minLevelOffsetToPromoteMember": [int],
    	"minLevelOffsetToDemoteMember":  [int],
    	"maxMembers":                    [int]
  	}
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    It will return an error if there are invalid parameters.

    * Code: `422`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

## Hook Routes

More about web hooks can be found in [Using WebHooks](using_webhooks.html).

  ### Supported Web Hook Event Types

  * `0 Game Created` - Same payload as the response of the Create Game Route
  * `1 Game Updated` - Same payload as the response of the Update Game Route

  ### Create Hook

  `POST /games/:gameID/hooks`

  Creates a new web hook for the specified game when the specified event type happens.

  * Payload

    ```
    {
      "type": [int],             // Event Type
      "hookURL": [string]        // the URL to call with the payload
                                 // for the specified event.
  	}
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true
        "publicID": [uuid]       // This is the id required to remove the hook.
                                 // It should be stored with the client app.
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Remove Hook

  `POST /games/:gameID/hooks/:hookPublicID`

  Removes a web hook created with the Create Hook route. No payload is required for this route.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true
      }
      ```

  * Error Response

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

## Player Routes

  ### Create Player
  `POST /games/:gameID/players`

  Creates a new player for the game with publicID=`gameID` with the given parameters.

  * Payload

    ```
    {
      "publicID":                      [string],  // 255 characters max, must be unique for a given game
    	"name":                          [string],  // 2000 characters max
    	"metadata":                      [JSON]
  	}
    ```

    Metadata is intended to include all the player's game specific informations that are not directly related to the khan's clan management. For example, player ranking, player trophies, player level, etc.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "publicID": [string]  // player public id
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    It will return an error if there are invalid parameters.

    * Code: `422`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Update Player
  `PUT /games/:gameID/players/:playerPublicID`

  Updates the player with the given publicID.


  * Payload

    ```
    {
    	"name":                          [string],  // 2000 characters max
    	"metadata":                      [JSON]
  	}
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    It will return an error if there are invalid parameters.

    * Code: `422`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

## Clan Routes

  ### Search Clans
  `GET /games/:gameID/clan-search`

  Searches for clans of a given game where the name or the publicID include the term passed in the query string.

  * URL Parameters

    ```
      term=[string]
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "clans": [
          {
            "name": [string],
            "metadata": [JSON],
            "publicID": [string]
          },
          {
            "name": [string],
            "metadata": [JSON],
            "publicID": [string]
          }
        ]
      }
      ```

      An empty list will be returned if no clans match the term.

  * Error Response

    It will return an error if an empty search term is sent.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": "A search term was not provided to find a clan."
      }
      ```

  ### List Clans
  `GET /games/:gameID/clans`

  List all clans for the game with publicID=`gameID`.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "clans": [
          {
            "name": [string],
            "metadata": [JSON],
            "publicID": [string]
          },
          {
            "name": [string],
            "metadata": [JSON],
            "publicID": [string]
          }
        ]
      }
      ```

      An empty list will be returned if there are no clans for the given game.

  ### Create Clan
  `POST /games/:gameID/clans`

  Creates a new clan for the game with publicID=`gameID` with the given parameters.

  * Payload

    ```
    {
      "publicID":                      [string],  // 255 characters max, must be unique for a given game
      "name":                          [string],  // 2000 characters max
      "metadata":                      [JSON],    
      "ownerPublicID":                 [string],  // must reference an existing player
      "allowApplication":              [boolean],
      "autoJoin":                      [boolean]
  	}
    ```

    Metadata is intended to include all the clan's game specific informations that are not directly related to the khan's clan management.

    * Some parameters require special attention:

      **allowApplication**: if set to false only the clan owner and members can invite players to join the clan, otherwise players can request to be added to a given clan.

      **autoJoin**: if set to true, when a player applies their membership is automatically approved. If set to false, the clan owner or one of its members must approve or deny the player's application.


  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "publicID": [string]  // clan public id
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Update Clan
  `PUT /games/:gameID/clans/:clanPublicID`

  Updates the clan with the given publicID.


  * Payload

    ```
    {
    	"name":                          [string],  // 2000 characters max
    	"metadata":                      [JSON],    
      "ownerPublicID":                 [string],  // must match the clan owner's public id
      "AllowApplication":              [boolean],
      "AutoJoin":                      [boolean]
  	}
    ```

    All parameters but the `ownerPublicID` will be updated.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Retrieve Clan
  `GET /games/:gameID/clans/:clanPublicID`

  Retrieves the clan with the given publicID. It will list all the clan information and its members.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "name": [string],
        "metadata": [JSON],
        "members": [
          {      
            "membershipLevel":     [int],
            "membershipApproved":  [boolean],
            "membershipDenied":    [boolean],
            "membershipCreatedAt": [int],     // milliseconds since epoch
            "membershipUpdatedAt": [int],     // milliseconds since epoch
            "playerPublicID":      [string],
            "playerName":          [string],
            "playerMetadata":      [JSON],
            "requestorPublicID":   [string],
            "requestorName":       [string],
          }
        ]
      }
      ```

  * Error Response

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Leave Clan
  `POST /games/:gameID/clans/:clanPublicID/leave`

  Allows the owner to leave the clan. If there are no clan members the clan will be deleted. Otherwise, the new clan owner will be the member with the highest level which has the oldest creation date.

  * Payload

    ```
    {
      "ownerPublicID": [string]  // must match the clan owner's public id
  	}
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Transfer Clan Ownership
  `POST /games/:gameID/clans/:clanPublicID/transfer-ownership`

  Allows the owner to transfer the clan's ownership to another clan member of their choice. The previous owner will then be a member with the maximum level allowed for the clan.

  * Payload

    ```
    {
      "ownerPublicID": [string],  // must match the clan owner's public id
      "playerPublicID": [string]  // must match a clan member's public id
  	}
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

## Membership Routes

  ### Apply For Membership

  `POST /games/:gameID/clans/:clanPublicID/memberships/application`

  Allows a player to ask to join the clan with the given publicID. If the clan's autoJoin property is true the member will be automatically approved. Otherwise, the membership must be approved by the clan owner or one of the clan members.

  * Payload

    ```
    {
      "level": [int],            // the level of the membership
      "playerPublicID": [string] // the player's public id
  	}
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Approve Or Deny Membership Application

  `POST /games/:gameID/clans/:clanPublicID/memberships/application/:action`

  `:action` must be either 'approve' or 'deny'.

  Allows the clan owner or a clan member to approve or deny a player's application to join the clan. The member's membership level must be at least the game's `minLevelToAcceptApplication`.

  * Payload

    ```
    {
      "level": [int],               // the level of the membership
      "playerPublicID": [string]    // the public id of player who made the application
      "RequestorPublicID": [string] // the public id of the clan member or the owner who will approve or deny the application
  	}
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Invite For Membership

  `POST /games/:gameID/clans/:clanPublicID/memberships/invitation`

  Allows a the clan owner or a clan member to invite a player to join the clan with the given publicID. If the request is made by a member of the clan, their membership level must be at least the game's `minLevelToCreateInvitation`. The membership must be approved by the player being invited.

  * Payload

    ```
    {
      "level": [int],               // the level of the membership
      "playerPublicID": [string],   // the public id player being invited
      "requestorPublicID": [string] // the public id of the member or the clan owner who is inviting
  	}
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Approve Or Deny Membership Invitation

  `POST /games/:gameID/clans/:clanPublicID/memberships/invitation/:action`

  `:action` must be either 'approve' or 'deny'.

  Allows a player member to approve or deny a player's invitation to join a given clan.

  * Payload

    ```
    {
      "playerPublicID": [string] // the public id of player who was invited
  	}
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Promote Or Demote Member

  `POST /games/:gameID/clans/:clanPublicID/memberships/:action`

  `:action` must be either 'promote' or 'demote'.

  Allows a the clan owner or a clan member to promote or demote another member. When promoting, the member's membership level will be increased by one, when demoting it will be decreased by one. The member's membership level must be at least `minLevelOffsetToPromoteMember` or `minLevelOffsetToDemoteMember` levels greater than the level of the player being promoted or demoted.

  * Payload

    ```
    {
      "playerPublicID": [string],   // the public id player being promoted or demoted
      "requestorPublicID": [string] // the public id of the member or the clan owner who is promoting or demoting
  	}
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Delete Membership

  `POST /games/:gameID/clans/:clanPublicID/memberships/delete`

  Allows a the clan owner or a clan member to remove another member from the clan. The member's membership level must be at least `minLevelToRemoveMember`. A member can leave the clan by sending the same `playerPublicID` and `requestorPublicID`.

  * Payload

    ```
    {
      "playerPublicID": [string],   // the public id player being deleted
      "requestorPublicID": [string] // the public id of the member or the clan owner who is deleting the membership
  	}
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```
