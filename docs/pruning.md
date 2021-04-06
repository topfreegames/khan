Pruning Stale Data
==================

Depending on your usage of Khan, you will accumulate some stale data. Some examples include:

* Pending Applications to a clan;
* Pending Invitations to a clan;
* Deleted Memberships (member left or was banned);
* Denied Memberships.

While there's nothing wrong with keeping this data in the Data Store, it will slow Khan down considerably depending on your usage of it.

## Gotchas

One good example of how this sort of stale data can go wrong is when building a clan suggestion for players. If you always suggest the same clans, eventually those clans will have thousands of unfulfilled applications.

That way, anytime someone requests info for these clans, Khan will have a hard time to fulfill that request.

## Pruning Stale Data

Khan has a `prune` command built-in, designed for the purpose of keeping your data balanced. It looks for some configuration keys in your games, and then decides on what data should be deleted.

### WARNING

This command performs a **HARD** delete on the memberships row and can't be undone. Please ensure you have frequent backups of your data store before applying pruning.

## Configuring Games to be Pruned

Configuring a game to be pruned is as easy as including some keys in the game's metadata property:

* `pendingApplicationsExpiration`: the number of **SECONDS** to wait before deleting a pending application;
* `pendingInvitesExpiration`: the number of **SECONDS** to wait before deleting a pending invitation;
* `deniedMembershipsExpiration`: the number of **SECONDS** to wait before deleting a denied membership;
* `deletedMembershipsExpiration`: the number of **SECONDS** to wait before deleting a deleted membership (either the member left or was banned).

**PLEASE** take note that all the expirations are in **SECONDS**. The timestamp used to compare the expiration to is the `updated_at` field of memberships.

Khan will delete any membership that meets one of the criteria above **AND** has an `updated_at` timestamp older than the relevant configuration subtracted in seconds from NOW.

### NOTICE

If you want a game to be pruned, **ALL** expiration keys **MUST** be set. Otherwise, Khan will ignore that game as far as pruning goes.

## Periodically Running Pruning

Khan's command line for pruning is:

```
$ khan prune -c /path/to/config.yaml
```

Khan will use the connection details in your specified config file. Double-check the config file being used to ensure that you won't lose any unwanted information.

## Pruning with a Container

Since Khan has container offers, you can also use a container for running pruning in any PaaS that supports Docker containers.

In order to use it, you need to configure these environment variables in the container:

* `KHAN_POSTGRES_HOST` - PostgreSQL to prune hostname;
* `KHAN_POSTGRES_PORT` - PostgreSQL to prune port;
* `KHAN_POSTGRES_USER` - PostgreSQL to prune username;
* `KHAN_POSTGRES_PASSWORD` - PostgreSQL to prune password;
* `KHAN_POSTGRES_DBNAME` - PostgreSQL to prune database name;
* `KHAN_PRUNING_SLEEP` - Number of seconds to sleep between pruning operations. Defaults to 3600.

The image can be found at our [official Docker Hub repository](https://hub.docker.com/r/tfgco/khan-prune/).
