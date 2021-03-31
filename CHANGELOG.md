## [5.0.1]
### Added
- Transactions on write player routes

## [5.0.0]
### Added
- Encryption to write player models functions `CreatePlayer` and `UpdatePlayer`
- Create `encrypted_players` table to trace the encryption proccess and support the future `encryption_script`

## [4.4.0]
### Added
- Go modules and `go.mod`, `go.sum`
- `util.security.go`
- Github actions workflow
- `EncryptionKey` to `api.App`
- Encryption to models function `Serialize` of `clanDetailsDAO`
- Encryption to models function `Serialize` of `PlayerDetailsDAO`
- Encryptiion to models functions `GetPlayerByID`, `GetPlayerByPublicID`, `GetPlayerDetails` and `GetClanDetails`

## Changed
- Update docker image base image to `1.15.2-alpine`
- Apply go modules changes on Makefile
- Centralize on `models/player.go` file the player serialization in favor of easiness encryption implementation

## Removed
- `Gopkg.toml` and `Gopkg.lock`



