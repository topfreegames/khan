## [5.3.5]
### Changed
- Fix `ApprovedAt` field update.

## [5.3.4]
### Changed
- Fix and improve logs in membership handlers

## [5.3.3]
### Changed
- Add weights in mongo text index

## [5.3.2]
### Changed
- Update `jaeger-client-go` to version `v2.29.1`

## [5.3.1]
### Changed
- Remove encryption entry when don't have encryption key

## [5.3.0]
### Changed
- Update `jaeger` to use official client
- Change `Makefile` to use `go mod download` instead `go mod tidy`

### Removed
- Encryption obligation

## [5.2.2]
### Changed
- Update go.mod

### Removed
- Profiling routes

## [5.1.1]
### Changed
- Fix encrypted players id type (#74) 
- Explicit database creation (#67)
- Add docker-compose for khan (#72)

## [5.1.0]
### Added
- Create script to encrypt player name

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



