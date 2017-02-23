Using Web Hooks
===============

You can use khan's web hooks to send detailed event information to other servers. The use cases for this are plenty, since the need for integrating micro-services is very common.

One possible use case would be to notify the chat channel for your game with information about people joining/leaving the clan or for new applications.

## Webhook Specific Configuration

Khan has some configuration entries specific to webhooks:

* `redis.host` - Redis server host used for [GoWorkers](https://github.com/jrallison/go-workers);
* `redis.port` - Redis server port used for [GoWorkers](https://github.com/jrallison/go-workers);
* `redis.database` - Redis server database used for [GoWorkers](https://github.com/jrallison/go-workers);
* `redis.pool` - Redis connection pool size used for [GoWorkers](https://github.com/jrallison/go-workers);
* `redis.password` - Redis password used for [GoWorkers](https://github.com/jrallison/go-workers);
* `webhooks.timeout` - Timeout for webhook HTTP connections;
* `webhooks.workers` - Number of [GoWorkers](https://github.com/jrallison/go-workers) to start with each instance of Khan;
* `webhooks.runStats` - Will the [GoWorkers](https://github.com/jrallison/go-workers) stats server run in each Khan instance?;
* `webhooks.statsPort` - Port that the stats server of [GoWorkers](https://github.com/jrallison/go-workers) will run in.

## Registering a Web Hook

Registering a web hook is done using the [Create Web Hook Route](API.html#create-hook). A hook can also be removed using the [Remove Web Hook Route](API.html#remove-hook). Just make sure you keep the PublicID that was returned by the Create Hook route as it is required to remove a hook.

## How do Web Hooks work?

When an event happen in Khan, it will look for all the hooks registered for the game that the event happened in.

It will then do a POST request to all of them with the payload relevant to the event (more about the payloads in the Events section). The payload for the events is just a JSON object in the body of the request.

Let's say we want to keep track of all the clans that are created in our game. We could create a web hook for the `ClanCreated` event and Khan would call our hook with this payload:

    {
      gameID: "my-game",
      publicID: "clan-public-id",
      name: "My Fancy Clan"
    }

We could then use this information to store this clan in our Database, to integrate with a chat channel, to provision some third-party system for clans, etc.

## URL Format and Flexibility

When registering a new URL, Khan allows you to specify the URL as a Template.

Let's assume you need to use the clan public ID in your URL to store a newly created clan at `http://my-system.com:3030/clans/some-clan/`. Instead of some-clan, we need to use the clan's public ID.

This is very easy to do with Khan. You can interpolate any of the keys in the payload using `{{key}}`. Bear in mind that only two types of keys can be used:

* top-level keys - if you use a key that's in the first level of depth in the payload, you are good with any type of key;
* dot separated keys - this type of key will only work if all the keys in the path are objects (except the last one).

Let's try with the payload for the player created event:

```
    {
        "success": true,
        "gameID":  "some-game",
        "publicID": "playerPublicID",
        "name": "Player Name",
        "metadata": {
            "score": 1200,
            "league": {
                ranking: "diamond",
                position: 30
            }
        }
    }
```

Now imagine we want the player to be included in the league he belongs to. We could use an URL like `http://my-server.com:3030/players/{{publicID}}/leagues/{{metadata.league.ranking}}/`. This would be translated by Khan to `http://my-server.com:3030/players/playerPublicID/leagues/diamond/`.

## Event Types

So what types of events can you [create Web Hooks for](http://khan-api.readthedocs.io/en/latest/API.html#create-hook)?

### Game Hooks

#### Game Updated

Event Type: `0`

Payload:

    {
        "success": true,
        "type": 0,                                  // Event Type
        "publicID": [string],                       // Game ID
        "name": [string],                           // Game Name
        "metadata": [JSON],                         // JSON Object containing game metadata.
        "membershipLevels": [JSON],                 // JSON Object mapping membership levels
        "minLevelToAcceptApplication": [int],       // Minimum level of membership required
                                                    // to accept players into clan
        "minLevelToCreateInvitation": [int],        // Minimum level of membership required
                                                    // to invite players into clan
        "minLevelToRemoveMember": [int],            // Minimum level of membership required
                                                    // to remove players from clan
        "minLevelOffsetToRemoveMember": [int],      // A player must be at least this offset
                                                    // higher than the player being removed.
        "minLevelOffsetToPromoteMember": [int],     // A player must be at least this offset
                                                    // higher than the player being promoted
        "minLevelOffsetToDemoteMember": [int],      // A player must be at least this offset
                                                    // higher than the player being demoted
        "maxMembers": [int],                        // Maximum number of players in the clan
        "maxClansPerPlayer": [int]                  // Maximum number of clans the player can be
                                                    // member of
    }

### Player Hooks

#### Player Created

Event Type: `1`

Payload:

    {
        "gameID":  [string],                        // Game ID
        "type": 1,                                  // Event Type
        "publicID": [string],                       // Created Player PublicID This id should
                                                    // be used when referring to the player in
                                                    // future operations.
        "name": [string],                           // Player Name
        "metadata": [JSON],                         // JSON Object containing player metadata
        "membershipCount": [int],                   // Number of clans this player is a member of
        "ownershipCount":  [int],                   // Number of clans this player is an owner of
        "id": [UUID],                               // unique id that identifies the hook
        "timestamp": [timestamp]                    // timestamp in the RFC3339 format
    }

#### Player Updated

Event Type: `2`

Payload:

    {
        "gameID":  [string],                        // Game ID
        "type": 2,                                  // Event Type
        "publicID": [string],                       // Created Player PublicID This id should
                                                    // be used when referring to the player in
                                                    // future operations.
        "name": [string],                           // Player Name
        "metadata": [JSON],                         // JSON Object containing player metadata
        "membershipCount": [int],                   // Number of clans this player is a member of
        "ownershipCount":  [int],                   // Number of clans this player is an owner of
        "id": [UUID],                               // unique id that identifies the hook
        "timestamp": [timestamp]                    // timestamp in the RFC3339 format
    }


### Clan Hooks

#### Clan Created

Event Type: `3`

Payload:

    {
        "gameID":  [string],                        // Game ID
        "type": 3,                                  // Event Type
        "clan": {
          "publicID": [string],                     // Created Clan PublicID This id should
                                                    // be used when referring to the clan in
                                                    // future operations.
          "name": [string],                         // Clan Name
          "metadata": [JSON],                       // JSON Object containing clan's metadata
          "allowApplication": [bool],               // Indicates whether this clan acceps applications
          "autoJoin": [bool]                        // Indicates whether this clan automatically
                                                    // accepts applications
        }
        "id": [UUID],                               // unique id that identifies the hook
        "timestamp": [timestamp]                    // timestamp in the RFC3339 format
    }

#### Clan Updated

Event Type: `4`

Payload:

    {
        "gameID":  [string],                        // Game ID
        "type": 4,                                  // Event Type
        "clan": {
          "publicID": [string],                     // Updated Clan PublicID This id should
                                                    // be used when referring to the clan in
                                                    // future operations.
          "name": [string],                         // Clan Name
          "metadata": [JSON],                       // JSON Object containing clan's metadata
          "allowApplication": [bool],               // Indicates whether this clan acceps applications
          "autoJoin": [bool]                        // Indicates whether this clan automatically
                                                    // accepts applications
        }
        "id": [UUID],                               // unique id that identifies the hook
        "timestamp": [timestamp]                    // timestamp in the RFC3339 format
    }

**WARNING**: This event may be skipped if the game's `clanHookFieldsWhitelist` field is not empty and none of the fields that actually changed is in the whitelist. Imagine you have a document like this:

```
{
    "trophies": 30,
    "country": "US",
}
```

The game is configured with `clanHookFieldsWhitelist = "country"`. That means that changing the number of trophies won't dispatch hooks, but changing the country (or any other clan property) will.

#### Clan Owner Left

Event Type: `5`

Payload:

    {
        "gameID": [string],
        "type": 5,                                      // Event Type
        "isDeleted": [bool],                            //Indicates whether the clan was deleted
                                                        //because there were no members left
        "clan": {
            "publicID": [string],                       // Updated Clan PublicID
            "name": [string],                           // Clan Name
            "metadata": [JSON],                         // JSON Object containing clan's metadata
            "allowApplication": [bool]                  // Indicates whether this clan acceps applications
            "autoJoin": [bool],                         // Indicates whether this clan automatically
                                                        // accepts applications
            "membershipCount":  [int],                  // Number of members in clan
        },
        "previousOwner": {                              // The owner that left
            "publicID": [string],                       // Previous Owner PublicID
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int]                    // Number of clans this player is an owner of
        },
        "newOwner": {                                   // After the owner left, this is the new owner
            "publicID": [string],                       // New Owner PublicID
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int]                    // Number of clans this player is an owner of
        },
        "id": [UUID],                                   // unique id that identifies the hook
        "timestamp": [timestamp]                        // timestamp in the RFC3339 format
    }

#### Clan Ownership Transferred to Another Player

Event Type: `6`

Payload:

    {
        "gameID": [string],
        "type": 6,                                  // Event Type
        "clan": {
            "publicID": [string],                       // Updated Clan PublicID
            "name": [string],                           // Clan Name
            "metadata": [JSON],                         // JSON Object containing clan's metadata
            "allowApplication": [bool]                  // Indicates whether this clan acceps applications
            "autoJoin": [bool],                         // Indicates whether this clan automatically
                                                        // accepts applications
            "membershipCount":  [int],                  // Number of members in clan
        },
        "previousOwner": {                                   // The previous owner
            "publicID": [string],                       // Previous Owner PublicID
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int]                    // Number of clans this player is an owner of
        },
        "newOwner": {                                   // Player that the owner transferred ownership to
            "publicID": [string],                       // New Owner PublicID
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int]                    // Number of clans this player is an owner of
        },
        "id": [UUID],                                   // unique id that identifies the hook
        "timestamp": [timestamp]                        // timestamp in the RFC3339 format
    }

### Membership Hooks

#### Membership Created

This event occurs if a player applies to a clan (player == requestor) or if a player is invited to a clan (player != requestor).

Event Type: `7`

Payload:

    {
        "gameID": [string],
        "type": 7,                                  // Event Type
        "clan": {
            "publicID": [string],                       // Clan that player is applying to
            "name": [string],                           // Clan Name
            "metadata": [JSON],                         // JSON Object containing clan's metadata
            "allowApplication": [bool]                  // Indicates whether this clan acceps applications
            "autoJoin": [bool],                         // Indicates whether this clan automatically
                                                        // accepts applications
            "membershipCount":  [int],                  // Number of members in clan
        },
        "player": {                                     // Player that is applying/being invited to the clan
            "publicID": [string],                       // Applicant PublicID
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int],                   // Number of clans this player is an owner of
            "membershipLevel":  [string]                // The level of the player's membership
        },
        "requestor": {                                  // Player that requested this membership application/invite
            "publicID": [string],                       // Requestor PublicID
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int]                    // Number of clans this player is an owner of
        },
        "id": [UUID],                                   // unique id that identifies the hook
        "timestamp": [timestamp]                        // timestamp in the RFC3339 format
    }

#### Membership Approved

This event occurs if an application or an invite to a clan gets approved. If the membership that was approved was an invitation into the clan, the requestor and the player will be the same Player. Otherwise, the requestor will be whomever approved the membership.

Event Type: `8`

Payload:

    {
        "gameID": [string],
        "type": 8,                                  // Event Type
        "clan": {
            "publicID": [string],                       // Clan that membership was approved
            "name": [string],                           // Clan Name
            "metadata": [JSON],                         // JSON Object containing clan's metadata
            "allowApplication": [bool]                  // Indicates whether this clan acceps applications
            "autoJoin": [bool],                         // Indicates whether this clan automatically
                                                        // accepts applications
            "membershipCount":  [int],                  // Number of members in clan
        },
        "player": {                                     // Player that was approved into the clan
            "publicID": [string],                       // Player PublicID
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int],                   // Number of clans this player is an owner of
            "membershipLevel":  [string]                // The level of the player's membership
        },
        "requestor": {                                  // Player that approved the membership
            "publicID": [string],                       // Requestor PublicID
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int]                    // Number of clans this player is an owner of
        },
        "creator": {                                    // Player that created the membership (invited or applied)
            "publicID": [string],                       // Creator PublicID
            "name": [string],                           // Creator Name
            "metadata": [JSON],                         // JSON Object containing creator metadata
            "membershipCount": [int],                   // Number of clans this creator is a member of
            "ownershipCount":  [int]                    // Number of clans this creator is an owner of
        },
        "id": [UUID],                                   // unique id that identifies the hook
        "timestamp": [timestamp]                        // timestamp in the RFC3339 format
    }

#### Membership Denied

This event occurs if an application or an invite to a clan gets denied. If the membership that was denied was an invitation into the clan, the requestor and the player will be the same Player. Otherwise, the requestor will be whomever denied the membership.

Event Type: `9`

Payload:

    {
        "gameID": [string],
        "type": 9,                                  // Event Type
        "clan": {
            "publicID": [string],                       // Clan that membership was denied
            "name": [string],                           // Clan Name
            "metadata": [JSON],                         // JSON Object containing clan's metadata
            "allowApplication": [bool]                  // Indicates whether this clan acceps applications
            "autoJoin": [bool],                         // Indicates whether this clan automatically
                                                        // accepts applications
            "membershipCount":  [int],                  // Number of members in clan
        },
        "player": {                                     // Player that was denied into the clan
            "publicID": [string],                       // Player PublicID
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int],                   // Number of clans this player is an owner of
            "membershipLevel":  [string]                // The level of the player's membership
        },
        "requestor": {                                  // Player that denied the membership
            "publicID": [string],                       // Requestor PublicID
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int]                    // Number of clans this player is an owner of
        },
        "creator": {                                    // Player that created the membership (invited or applied)
            "publicID": [string],                       // Creator PublicID
            "name": [string],                           // Creator Name
            "metadata": [JSON],                         // JSON Object containing creator metadata
            "membershipCount": [int],                   // Number of clans this creator is a member of
            "ownershipCount":  [int]                    // Number of clans this creator is an owner of
        },
        "id": [UUID],                                   // unique id that identifies the hook
        "timestamp": [timestamp]                        // timestamp in the RFC3339 format
    }

#### Member Promoted

Event Type: `10`

Payload:

    {
        "gameID": [string],
        "type": 10,                                  // Event Type
        "clan": {
            "publicID": [string],                       // Clan that member was promoted
            "name": [string],                           // Clan Name
            "metadata": [JSON],                         // JSON Object containing clan's metadata
            "allowApplication": [bool]                  // Indicates whether this clan acceps applications
            "autoJoin": [bool],                         // Indicates whether this clan automatically
                                                        // accepts applications
            "membershipCount":  [int],                  // Number of members in clan
        },
        "player": {                                     // Player that was promoted
            "publicID": [string],                       // Player PublicID
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int],                   // Number of clans this player is an owner of
            "membershipLevel":  [string]                // The new level of the player's membership
        },
        "requestor": {                                  // Player that promoted this member
            "publicID": [string],                       // Requestor PublicID
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int]                    // Number of clans this player is an owner of
        },
        "id": [UUID],                                   // unique id that identifies the hook
        "timestamp": [timestamp]                        // timestamp in the RFC3339 format
    }

#### Member Demoted

Event Type: `11`

Payload:

    {
        "gameID": [string],
        "type": 11,                                  // Event Type
        "clan": {
            "publicID": [string],                       // Clan that member was demoted
            "name": [string],                           // Clan Name
            "metadata": [JSON],                         // JSON Object containing clan's metadata
            "allowApplication": [bool]                  // Indicates whether this clan acceps applications
            "autoJoin": [bool],                         // Indicates whether this clan automatically
                                                        // accepts applications
            "membershipCount":  [int],                  // Number of members in clan
        },
        "player": {                                     // Player that was demoted
            "publicID": [string],                       // Player PublicID
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int]                    // Number of clans this player is an owner of,
            "membershipLevel":  [string]                // The new level of the player's membership
        },
        "requestor": {                                  // Player that demoted this member
            "publicID": [string],                       // Requestor PublicID
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int]                    // Number of clans this player is an owner of
        },
        "id": [UUID],                                   // unique id that identifies the hook
        "timestamp": [timestamp]                        // timestamp in the RFC3339 format
    }

#### Member Left

Event Type: `12`

Payload:

    {
        "gameID": [string],
        "type": 12,                                  // Event Type
        "clan": {
            "publicID": [string],                       // Clan that member left
            "name": [string],                           // Clan Name
            "metadata": [JSON],                         // JSON Object containing clan's metadata
            "allowApplication": [bool]                  // Indicates whether this clan acceps applications
            "autoJoin": [bool],                         // Indicates whether this clan automatically
                                                        // accepts applications
            "membershipCount":  [int],                  // Number of members in clan
        },
        "player": {                                     // Player that left
            "publicID": [string],                       // Player PublicID
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int],                   // Number of clans this player is an owner of
            "membershipLevel":  [string]                // The level of the player's membership
        },
        "requestor": {                                  // Player that removed leaving player (if they left
                                                        // on their own, then this is the same as player)
            "publicID": [string],                       // Requestor PublicID
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int]                    // Number of clans this player is an owner of
        },
        "id": [UUID],                                   // unique id that identifies the hook
        "timestamp": [timestamp]                        // timestamp in the RFC3339 format
    }
