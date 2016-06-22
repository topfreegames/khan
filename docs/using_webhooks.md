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

## Event Types

So what types of events can you create Web Hooks for?

### Game Updated

Event Type: `0`

Payload: 

    {
        "success": true,
        "publicID": [string],                       // Game ID.
        "name": [string],                           // Game Name.
        "metadata": [JSON],                         // JSON Object containing game metadata.
        "minMembershipLevel": [int],                // Minimum level of membership.
        "maxMembershipLevel": [int],                // Maximum level of membership.
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
        "maxMembers": [int]                         // Maximum number of players in the clan.
    }

### Player Created

Event Type: `1`

Payload:

    {
        "success": true,
        "gameID":  [string],                        // Game ID
        "publicID": [string],                       // Created Player PublicID. This id should
                                                    // be used when referring to the player in 
                                                    // future operations.
        "name": [string],                           // Player Name
        "metadata": [JSON],                         // JSON Object containing player metadata
    }

### Player Updated

Event Type: `2`

Payload:

    {
        "success": true,
        "gameID":  [string],                        // Game ID
        "publicID": [string],                       // Created Player PublicID. This id should
                                                    // be used when referring to the player in 
                                                    // future operations.
        "name": [string],                           // Player Name
        "metadata": [JSON],                         // JSON Object containing player metadata
    }
