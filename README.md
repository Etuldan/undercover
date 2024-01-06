# Undercover

## How to build

### Manual
Move to src/websocket-server directory 
- `go mod download`
- `go build`

### Using Docker
`docker build -t undercover-server .`

## Client communication with the Server
Each response from the server is in json format.
The server should respond to every command, and should them message to each client when a game state changed.

### Models
#### Command
##### Message
##### Enums
- CommandCode

| Code | Value |
|---|---|
| Host | host |
| Start | start |
| Join | join |
| Kick | kick |
| Play | play |
| Leave | leave |
| Status | status |

#### Response
##### Message
```json
{
   "error":{
      "code": errorCode,
      "message": errorMessage
   },
   "info":{
      "message": infoMessage,
      "action": actionCode,
      "initiator": playerObject
   },
   "gameInfo":{
      "gameId": gameGuid,
      "players": playerArrayObject,
      "turn": gameTurn,
      "action": gameAction
   },
}
```

| Variable | Type | Meaning |
|---|---|---|
| errorCode | integer | If non 0, an error occurred. See enum Error |
| errorMessage | string | Human readable error message, in case of error |
| infoMessage | string | Human readable info message |
| actionCode | integer | If non 0, see enum Action |
| initator | object | Information of a player who done an specific action |
| playerArrayObject | object | Informations of all Players in the game |
| gameGuid | string | Unique Id of the game, GuidFormat |
| gameTurn | integer | Determine the current player to play. See playerPosition |
| gameAction | integer | If non 0, see enum Action |

##### Player
```json
{
   "nickname": playerNickname,
   "rank": playerRank,
   "position": playerPosition,
   "eliminated": playerEliminated
}
```

| Variable | Type | Meaning |
|---|---|---|
| playerNickname | string | Nickname of a player |
| playerRank | integer | Rank of the Player, see enum Rank |
| playerPosition | integer | Position, as order to play, of the Player |
| playerEliminated | boolean | Boolean to describe is the Player is eliminated |

##### Enums
- Error

| Code | Value |
|---|---|
| NoError | 0 |
| GameNotFound | 1 |
| GameNotAvailable | 2 |
| NicknameNotAvailable | 3 |
| IncorrectGameState | 4 |
| PlayerNotFound | 5 |
| InsufficientPermission | 6 |
| NotYourTurn | 7 |

- Action

| Code | Value |
|---|---|
| NoAction | 0 |
| WriteDown | 1 |
| Vote | 2 |
| Voted | 3 |
| Eliminated | 4 |
| DisplayWord | 5 |
| WhiteGuess | 6 |
| Winner | 7 |
| Closed | 8 |

- GameAction

| Code | Value |
|---|---|
| NoGameAction | 0 |
| MrWhiteGuessAttempt | 1 |

- Rank

| Code | Value |
|---|---|
| NoRank | 0 |
| Host | 1 |
| Guest | 2 |

- Role

| Code | Value |
|---|---|
| NoRole | 0 |
| Undercover | 1 |
| White | 2 |
| Civilian | 3 |