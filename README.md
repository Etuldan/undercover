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
TODO

#### Response
##### Message
```json
{
   "error":{
      "code": errorCode,
      "message":errorMessage
   },
   "info":{
      "message":infoMessage,
      "action": actionCode,
      "initiator": playerObject
   },
   "gameInfo":{
      "gameId":gameGuid,
      "players": playerArrayObject,
      "turn": gameTurn,
      "action":gameAction
   },
}
```

| Variable | Type | Meaning |
-----------------------------
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
   "nickname":playerNickname,
   "rank":playerRank,
   "position":playerPosition,
   "eliminated": playerEliminated
}
```

| Variable | Type | Meaning |
-----------------------------
| playerNickname | string | Nickname of a player |
| playerRank | integer | Rank of the Player, see enum Rank |
| playerPosition | integer | Position, as order to play, of the Player |
| playerEliminated | boolean | Boolean to describe is the Player is eliminated |

##### Enums
- Error
NoError
GameNotFound
GameNotAvailable
NicknameNotAvailable
IncorrectGameState
PlayerNotFound
InsufficientPermission
NotYourTurn

- Action
NoAction
WriteDown
Vote
Voted
Eliminated
DisplayWord
WhiteGuess
Winner
Closed

- GameAction
NoGameAction
MrWhiteGuessAttempt

- Rank
NoRank
Host
Guest

- Role
NoRole
Undercover
White
Civilian