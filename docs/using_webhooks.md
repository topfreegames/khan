Using Web Hooks
===============

You can use khan's web hooks to send detailed event information to other servers. The use cases for this are plenty, since the need for integrating micro-services is very common.

One possible use case would be to notify the chat channel for your game with information about people joining/leaving the clan or for new applications.

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

So what types of events can you create Web Hooks for?

### Game Hooks

#### Game Updated

Event Type: `0`

Payload:

    {
        "success": true,
        "publicID": [string],                       // Game ID.
        "name": [string],                           // Game Name.
        "metadata": [JSON],                         // JSON Object containing game metadata.
        "membershipLevels": [JSON],                 // JSON Object mapping membership levels
        "minLevelToAcceptApplication": [int],       // Minimum level of membership required
                                                    // to accept players into clan.
        "minLevelToCreateInvitation": [int],        // Minimum level of membership required
                                                    // to invite players into clan.
        "minLevelToRemoveMember": [int],            // Minimum level of membership required
                                                    // to remove players from clan.
        "minLevelOffsetToRemoveMember": [int],      // A player must be at least this offset
                                                    // higher than the player being removed.
        "minLevelOffsetToPromoteMember": [int],     // A player must be at least this offset
                                                    // higher than the player being promoted.
        "minLevelOffsetToDemoteMember": [int],      // A player must be at least this offset
                                                    // higher than the player being demoted.
        "maxMembers": [int],                        // Maximum number of players in the clan.
        "maxClansPerPlayer": [int]                  // Maximum number of clans the player can be
                                                    // member of.
    }

### Player Hooks

#### Player Created

Event Type: `1`

Payload:

    {
        "gameID":  [string],                        // Game ID
        "publicID": [string],                       // Created Player PublicID. This id should
                                                    // be used when referring to the player in
                                                    // future operations.
        "name": [string],                           // Player Name
        "metadata": [JSON],                         // JSON Object containing player metadata
		"membershipCount": [int],                   // Number of clans this player is a member of
		"ownershipCount":  [int]                    // Number of clans this player is an owner of
    }

#### Player Updated

Event Type: `2`

Payload:

    {
        "gameID":  [string],                        // Game ID
        "publicID": [string],                       // Created Player PublicID. This id should
                                                    // be used when referring to the player in
                                                    // future operations.
        "name": [string],                           // Player Name
        "metadata": [JSON],                         // JSON Object containing player metadata
		"membershipCount": [int],                   // Number of clans this player is a member of
		"ownershipCount":  [int]                    // Number of clans this player is an owner of
    }


### Clan Hooks

#### Clan Created

Event Type: `3`

Payload:

    {
        "gameID":  [string],                        // Game ID
        "publicID": [string],                       // Created Clan PublicID. This id should
                                                    // be used when referring to the clan in
                                                    // future operations.
        "name": [string],                           // Clan Name
        "metadata": [JSON],                         // JSON Object containing clan's metadata
        "allowApplication": [bool]                  // Indicates whether this clan acceps applications
        "autoJoin": [bool]                          // Indicates whether this clan automatically
                                                    // accepts applications
    }

#### Clan Updated

Event Type: `4`

Payload:

    {
        "gameID":  [string],                        // Game ID
        "publicID": [string],                       // Updated Clan PublicID. This id should
                                                    // be used when referring to the clan in
                                                    // future operations.
        "name": [string],                           // Clan Name
        "metadata": [JSON],                         // JSON Object containing clan's metadata
        "allowApplication": [bool]                  // Indicates whether this clan acceps applications
        "autoJoin": [bool]                          // Indicates whether this clan automatically
                                                    // accepts applications
    }

#### Leave Clan

Event Type: `5`

Payload:

    {
        "gameID": [string],
        "clan": {
            "publicID": [string],                       // Updated Clan PublicID.
            "name": [string],                           // Clan Name
            "metadata": [JSON],                         // JSON Object containing clan's metadata
            "allowApplication": [bool]                  // Indicates whether this clan acceps applications
            "autoJoin": [bool],                         // Indicates whether this clan automatically
                                                        // accepts applications
            "membershipCount":  [int],                  // Number of members in clan
        },
        "newOwner": {                                   // After the owner left, this is the new owner
            "publicID": [string],                       // New Owner PublicID.
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int]                    // Number of clans this player is an owner of
        }
    }

#### Transfer Clan Ownership

Event Type: `6`

Payload:

    {
        "gameID": [string],
        "clan": {
            "publicID": [string],                       // Updated Clan PublicID.
            "name": [string],                           // Clan Name
            "metadata": [JSON],                         // JSON Object containing clan's metadata
            "allowApplication": [bool]                  // Indicates whether this clan acceps applications
            "autoJoin": [bool],                         // Indicates whether this clan automatically
                                                        // accepts applications
            "membershipCount":  [int],                  // Number of members in clan
        },
        "newOwner": {                                   // Player that the owner transferred ownership to
            "publicID": [string],                       // New Owner PublicID.
            "name": [string],                           // Player Name
            "metadata": [JSON],                         // JSON Object containing player metadata
            "membershipCount": [int],                   // Number of clans this player is a member of
            "ownershipCount":  [int]                    // Number of clans this player is an owner of
        }
    }
