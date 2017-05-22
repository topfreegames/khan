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

## Status Routes

  ### Status

  `GET /status`

  Returns statistics on the health of khan.

  * Success Response
    * Code: `200`
    * Content:

      ```
        {
          "app": {
            "errorRate": [float]        // Exponentially Weighted Moving Average Error Rate
          },
          "dispatch": {
            "pendingJobs": [int]        // Pending hook jobs to be sent
          }
        }
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
      "membershipLevels":              [JSON],
      "minLevelToAcceptApplication":   [int],
      "minLevelToCreateInvitation":    [int],
      "minLevelToRemoveMember":        [int],
      "minLevelOffsetToRemoveMember":  [int],
      "minLevelOffsetToPromoteMember": [int],
      "minLevelOffsetToDemoteMember":  [int],
      "maxMembers":                    [int],
      "maxClansPerPlayer":             [int],
      "cooldownAfterDeny":             [int],
      "cooldownAfterDelete":           [int],
      "cooldownBeforeInvite":          [int],
      "cooldownBeforeApply":           [int],
      "maxPendingInvites":             [int],
      "clanHookFieldsWhitelist":       [string],
      "playerHookFieldsWhitelist":     [string],
    }
    ```

    Metadata is intended to include all game specific configuration that is not directly related to the khan's clan management. For example, clank ranking, clan trophies, clan description, etc.

    * Some parameters require special attention:

      **membershipLevels**: Should be a JSON mapping levels (strings) to integers:
      ```
      {
        "Member": 1,
        "Elder": 2,
        "CoLeader": 3
      }
      ```
      The integer part of the level will be used to compare with `minLevel...` and `minLevelOffset...` when performing membership operations.

      **minLevelToAcceptApplication**: A member cannot accept a player's application to join the clan unless their level is greater or equal to this parameter.

      **minLevelToCreateInvitation**: A member cannot invite a player to join the clan unless their level is greater or equal to this parameter.

      **minLevelOffsetToRemoveMember**: A member cannot remove another member unless their level is at least `MinLevelOffsetToRemoveMember` levels greater than the level of the member they wish to promote.

      **minLevelOffsetToPromoteMember**: A member cannot promote another member unless their level is at least `minLevelOffsetToPromoteMember` levels greater than the level of the member they wish to promote.

      **minLevelOffsetToDemoteMember**: A member cannot demote another member unless their level is at least `minLevelOffsetToDemoteMember` levels greater than the level of the member they wish to demote.

      **maxMembers**: Maximum number of members a clan of this game can have.

      **maxClansPerPlayer**: Maximum number of clans a player can be member of.

      **cooldownAfterDeny**: Time (in seconds) the player must wait before applying/being invited to a new membership after the last membership application/invite was denied.

      **cooldownAfterDelete**:  Time (in seconds) the player must wait before applying/being invited to a new membership after the last membership application/invite was deleted.

      **cooldownBeforeInvite**:  Time (in seconds) a clan member must wait before inviting a member to a new membership after the last membership application/invite was created.

      **cooldownBeforeApply**:  Time (in seconds) a player must wait before applying for a clan after the last membership application/invite was created.

      **maxPendingInvites**:  Maximum number of pending invites each player can have withstanding. Set this value to -1 if your game has no limits on maximum pending invites.

      **clanHookFieldsWhitelist**: If you change metadata very frequently in clans, you can specify here the fields in your metadata document for which you'd like to have the clan updated hook triggered. If no fields are specified, the hook will be triggered in all updates. If you don't want any metadata changes to trigger hooks, just set this to "none" or any key that does not exist in your metadata document.

      **playerHookFieldsWhitelist**: If you change metadata very frequently in players, you can specify here the fields in your metadata document for which you'd like to have the player updated hook triggered. If no fields are specified, the hook will be triggered in all updates. If you don't want any metadata changes to trigger hooks, just set this to "none" or any key that does not exist in your metadata document.

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
      "membershipLevels":              [JSON],
      "minLevelToAcceptApplication":   [int],
      "minLevelToCreateInvitation":    [int],
      "minLevelToRemoveMember":        [int],
      "minLevelOffsetToPromoteMember": [int],
      "minLevelOffsetToDemoteMember":  [int],
      "maxMembers":                    [int],
      "maxClansPerPlayer":             [int],
      "cooldownAfterDeny":             [int],
      "cooldownAfterDelete":           [int],
      "cooldownBeforeInvite":          [int],
      "cooldownBeforeApply":           [int],
      "maxPendingInvites":             [int],
      "clanHookFieldsWhitelist":       [string],
      "playerHookFieldsWhitelist":     [string]
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

  * `0 Game Updated` - Happens when a game is updated;
  * `1 Player Created` - Happens when a new player is created;
  * `2 Player Updated` - Happens when an existing player is updated;
  * `3 Clan Created` - Happens when a new clan is created;
  * `4 Clan Updated` - Happens when an existing clan is updated;
  * `5 Clan Owner Left` - Happens when the owner of a clan leaves the clan without transferring the ownership to someone else;
  * `6 Ownership of the Clan transferred` - Happens when the owner of a clan transfers ownership to another player;
  * `7 Membership Created` - Happens when either someone applies to a clan or is invited;
  * `8 Membership Approved` - Happens when a pending membership to a clan is approved;
  * `9 Membership Denied` - Happens when a pending membership to a clan is denied;
  * `10 Member Promoted` - Happens when a member of the clan is promoted;
  * `11 Member Demoted` - Happens when a pending member of the clan is demoted;
  * `12 Member Left` - Happens when a member of the clan is either removed or leaves the clan.

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

  `DELETE /games/:gameID/hooks/:hookPublicID`

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
      "publicID":                      [string],   // 255 characters max, must be unique for a given game
      "name":                          [string],   // 2000 characters max
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

  ### Retrieve Player
  `GET /games/:gameID/players/:playerPublicID`

  Gets the player with the given publicID.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "publicID": [string], // Player publicID
        "name": [string], // Player Name
        "metadata": [JSON], // Player Metadata
        "createdAt": [int64], // timestamp in milliseconds of when the player was created
        "updatedAt": [int64]  // timestamp in milliseconds of when the player was last updated

        //All clans the player is involved with show here
        "clans":{
          // Clans the player ownership
          "owned":[
            { "name": [string], "publicID": [string] }, // clan name and publicID
          ],

          // Clans the player has been approved and is currently a member of
          "approved":[
            { "name": [string], "publicID": [string] }, // clan name and publicID
          ],

          // Clans the player has been banned from
          "banned":[
            { "name": [string], "publicID": [string] }, // clan name and publicID
          ],

          // Clans the player has been rejected from
          "denied":[
            { "name": [string], "publicID": [string] }, // clan name and publicID
          ],

          // Clans the player has pending applications to
          "pendingApplications":[
            { "name": [string], "publicID": [string] }, // clan name and publicID
          ],

          // Clans the player has pending invites to
          "pendingInvites":[
            { "name": [string], "publicID": [string] }, // clan name and publicID
          ]
        },

        // All memberships this player has with details
        "memberships":[
          {
            //if approved, denied and banned are false, the membership is pending approval
            "approved": [bool],
            "denied": [bool],
            "banned": [bool],

            //clan the player applied to
            "clan":{
              "metadata": [JSON],
              "name": [string],
              "publicID": [string],
              "membershipCount": [int]
            },
            "createdAt":  [int64], // timestamp the player applied to a clan
            "updatedAt":  [int64], // timestamp that the membership was last updated
            "deletedAt":  [int64], // timestamp that the player was banned
            "approvedAt": [int64], // timestamp that the player was approved
            "deniedAt":   [int64], // timestamp that the player was denied

            "level": [string],    // level of the player in this clan

            "message": [string], // empty if membership not created using an application
                                 // or the application message otherwise

            // Player that requested membership
            // If the player was invited, this should be another player.
            // Otherwise, this is the same as the player.
            // If the membership level is 'owner' this key does not exist.
            "requestor":{
              "publicID": [string]
              "name": [string],
              "metadata": [JSON],
            },

            // Player that approved this membership
            // If the membership is not yet approved or the level is 'owner' this key does not exist
            "approver":{
              "publicID": [string]
              "name": [string],
              "metadata": [JSON],
            },

            // Player that denied this membership
            // If the membership is not yet denied
            "denier":{
              "publicID": [string]
              "name": [string],
              "metadata": [JSON],
            },

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

## Clan Routes

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
      "allowApplication":              [boolean],
      "autoJoin":                      [boolean]
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

  The roster, as well as the memberships return a list of players, following this structure:

    {
        "level": [int],  // not returned for denied/banned memberships
        "message": [string], // the message sent with the application
                             // or "" if the membership was not created with and application
        "player": {
          "publicID": [string],
          "name":     [string],
          "metadata": [JSON]      
          "approver": {  // player that approved this membership
            "publicID": [string],
            "name":     [string],
          }
        }
    }

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "name": [string],
        "metadata": [JSON],
        "allowApplication": [bool],
        "autoJoin": [bool],
        "membershipCount": [int],
        "owner": [
            "publicID": [string],
            "name":     [string],
            "metadata": [JSON],
        ],
        "roster": [
          [membership],     //a list of the above membership structure
        ],
        "memberships": [
          "pendingApplications": [
            [membership],   //a list of all the pending applications in this clan
          ],
          "pendingInvites": [
            [membership],   //a list of all the pending invites in this clan
          ],
          "denied": [
            [membership],   //a list of all the denied memberships in this clan
          ],
          "banned": [
            [membership],   //a list of all the banned memberships in this clan
          ],
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

  ### Clan Summary
  `GET /games/:gameID/clans/:clanPublicID/summary`

  Returns a summary of the details of the clan with the given publicID.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "publicID": [string],
        "name": [string],
        "metadata": [JSON],
        "allowApplication": [bool],
        "autoJoin": [bool],
        "membershipCount": [int]
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

  ### Clans Summary
  `GET /games/:gameID/clans-summary?clanPublicIds=clan1,clan2,clan3`

  Returns a summary of the details for each one of the clans with the given publicIDs.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        clans: [
          {
            "publicID": [string],
            "name": [string],
            "metadata": [JSON],
            "allowApplication": [bool],
            "autoJoin": [bool],
            "membershipCount": [int]
          },
          {
            "publicID": [string],
            "name": [string],
            "metadata": [JSON],
            "allowApplication": [bool],
            "autoJoin": [bool],
            "membershipCount": [int]
          },
          ...    
        ]
      }
      ```

  If no clanPublicIDs are provided:

  * Error Response

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  If at least one of the clan was not found:

  * Error Response

    * Code: `404`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  In case of other errors:

  * Error Response

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### List Clans
  `GET /games/:gameID/clans`

  List all clans for the game with publicID=`gameID`.

  **Warning**

  Depending on the number of clans in your game this can be a **VERY** expensive operation! Be wary of using this. A better way of getting clans is using clan search.

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
            "membershipCount": [int],
            "publicID": [string],
            "allowApplication": [bool],
            "autoJoin": [bool]
          }
        ]
      }
      ```

      An empty list will be returned if there are no clans for the given game.

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
            "membershipCount": [int],
            "publicID": [string],
            "allowApplication": [bool],
            "autoJoin": [bool]
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

  ### Leave Clan
  `POST /games/:gameID/clans/:clanPublicID/leave`

  Allows the owner to leave the clan. If there are no clan members the clan will be deleted. Otherwise, the new clan owner will be the member with the highest level which has the oldest creation date.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "isDeleted":         [bool],       //Indicates whether the clan was deleted
                                           //because there were no members left
        "previousOwner": {
          "publicID":        [string],
          "name":            [string],
          "metadata":        [JSON],
          "membershipCount": [int],
          "ownershipCount":  [int]
        },
        "newOwner": {
          "publicID":        [string],
          "name":            [string],
          "metadata":        [JSON],
          "membershipCount": [int],
          "ownershipCount":  [int]
        }
      }
      ```

  * Error Response

    It will return an error if clan is not found.

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
      "playerPublicID": [string]  // must match a clan member's public id
    }
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "previousOwner": {
          "publicID":        [string],
          "name":            [string],
          "metadata":        [JSON],
          "membershipCount": [int],
          "ownershipCount":  [int]
        },
        "newOwner": {
          "publicID":        [string],
          "name":            [string],
          "metadata":        [JSON],
          "membershipCount": [int],
          "ownershipCount":  [int]
        }
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
      "level": [string],          // the level of the membership
      "playerPublicID": [string], // the player's public id
      "message": [string]         // optional, a message sent by the player
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
      "playerPublicID": [string]    // the public id of player who made the application
      "requestorPublicID": [string] // the public id of the clan member or the owner who will approve or deny the application
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
      "level": [string],            // the level of the membership
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
