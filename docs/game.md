Game Configuration
==================

Being a multi-tenant clan server, Khan allows for many different configurations per tenant. Each tenant is a different game and is identified by it's game ID.

Before any clan operation can be performed, you must create a game in Khan. The good news here is that creating/updating games are idempotent operations. You can keep executing it any time your game changes. That's ideal to be executed in a deploy script, for instance.

## Creating/Updating a Game

We recommend that the `Update` operation of the Game resource be used in detriment of the `Create` one. The reasoning here is that the Ùpdate` operation is idempotent (you can run it as many times as you want with the same result). If your game does not exist yet, it will create it, otherwise just updated it with the new configurations.

To Create/Update your game, just do a `PUT` request to `http://my-khan-server/games/my-game-public-id`, where `my-game-public-id` is the ID you'll be using for all your game's operations in the future. The payload for the request is a JSON object in the body and should be as follows:

``` 
    {
      "name":                          [string],
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
      "playerHookFieldsWhitelist":     [string],
    }
```

If the operation is successful, you'll receive a JSON object saying it succeeded:

```
      {
        "success": true
      }

```

## Game Configuration Settings

As can be seen from the previous section, there are a lot of different configurations you can do per game. These will be thoroughly explained in this section.

### name

The name of your game. This is used mainly for easier reasoning of what this game is when debugging.

**Type**: `string`<br />
**Sample Value**: `My Sample Game`

### metadata

Metadata related to your clan. This is a JSON object and can store anything you need to. Each game will probably have a different usage for this attribute: clan nationality, clan flag image URL, number of victories for the clan to date, etc.

This value is a black box as far as Khan is concerned. It's not used to decide any rules for clan management.

**Type**: `JSON`<br />
**Sample Value**: `{ "country": "BR", "language": "pt-BR" }`

### membershipLevels

The available membership levels the specified game supports. This is a way to specify the hierarchy between members in your game's clans. These levels are used in other configuration settings like `minLevelToRemoveMember` or `minLevelOffsetToPromoteMember`, among others.

The membership values (integer) should grow in importance, with the highest number being the highest member level.

**Type**: `JSON`<br />
**Sample Value**: `{ "member": 1, "leader": 2, "owner": 3 }`

### minLevelToAcceptApplication

The minimum member level (as specified in the membershipLevels configuration) required to accept a pending application to the clan.

**Type**: `integer`<br />
**Sample Value**: `2`

### minLevelToCreateInvitation

The minimum member level (as specified in the membershipLevels configuration) required to invite someone into a clan.

**Type**: `integer`<br />
**Sample Value**: `2`

### minLevelToRemoveMember

The minimum member level (as specified in the membershipLevels configuration) required to remove someone from the clan.

**Type**: `integer`<br />
**Sample Value**: `2`

### minLevelOffsetToRemoveMember

This configuration specifies the required difference in level between a player and the player being removed from the clan.

Let's look at an example to make things easier:

```
John has a membership level of 3, Paul has a membership level of 2 and
Ted has a membership level of 1.

If the clan has a minLevelOffsetToRemoveMember of 2, that means that only
John can remove Ted, but if that configuration is 1, then both John and Paul
can remove Ted.
```

**Type**: `integer`<br />
**Sample Value**: `2`

### minLevelOffsetToPromoteMember

This configuration specifies the required difference in level between a player and the player being promoted (calculated BEFORE the promotion).

What this means is that a player can only promote another player if that player is `minLevelOffsetToPromoteMember` levels below their own level before the promotion takes place.

If the `minLevelOffsetToPromoteMember` is greater than 1, then only the clan owner can promote someone to the highest available level(s).

Let's look at an example to make things easier:

```
John has a membership level of 5, Paul has a membership level of 3 and
Ted has a membership level of 1.

If the clan has a minLevelOffsetToPromoteMember of 2, that means that only
John can promote Ted up to level 4, but if that configuration is 1,
then both John and Paul can promote Ted (Paul can promote Ted to level 3).
```

**Type**: `integer`<br />
**Sample Value**: `2`

### minLevelOffsetToDemoteMember

This configuration specifies the required difference in level between a player and the player being demoted (calculated BEFORE the demotion).

If the `minLevelOffsetToDemoteMember` is greater than 1, then only the clan owner can demote someone from the highest available level(s).

Let's look at an example to make things easier:

```
John has a membership level of 5, Paul has a membership level of 4 and
Ted has a membership level of 3.

If the clan has a minLevelOffsetToDemoteMember of 2, that means that only
John can demote Ted, but if that configuration is 1, then both John and
Paul can demote Ted.
```

**Type**: `integer`<br />
**Sample Value**: `2`

### maxMembers

This configuration specifies the maximum number of members a clan can have.

**Type**: `integer`<br />
**Sample Value**: `50`

### maxClansPerPlayer

This configuration specifies the maximum number of clans a player can be a member of.

**Type**: `integer`<br />
**Sample Value**: `1`

### cooldownAfterDeny

Time (in seconds) the player must wait before applying/being invited to a new membership after the last membership application/invite was denied.

**Type**: `integer`<br />
**Sample Value**: `360`

### cooldownAfterDelete

Time (in seconds) the player must wait before applying/being invited to a new membership after the last membership application/invite was deleted.

**Type**: `integer`<br />
**Sample Value**: `720`

### cooldownBeforeInvite

Time (in seconds) a clan member must wait before inviting a member to a new membership after the last membership application/invite was created.

**Type**: `integer`<br />
**Sample Value**: `720`

### cooldownBeforeApply

Time (in seconds) a player must wait before applying for a clan after the last membership application/invite was created.

**Type**: `integer`<br />
**Sample Value**: `480`

### maxPendingInvites

Maximum number of pending invites each player can have withstanding. Set this value to `-1` if your game has no limits on maximum pending invites.

**Type**: `integer`<br />
**Sample Value**: `20`

### clanHookFieldsWhitelist

A comma-separated-values list of properties in the clan's metadata that will trigger the Clan Updated hook upon change.

**Type**: `string`<br />
**Sample Value**: `trophies,country`

### playerHookFieldsWhitelist

A comma-separated-values list of properties in the player's metadata that will trigger the Player Updated hook upon change.

**Type**: `string`<br />
**Sample Value**: `trophies,country`
