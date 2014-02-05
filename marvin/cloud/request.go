package cloud

import (
    "http"
    "io"
    "marvin/store/datastore"
    "marvin/json/jsondata"
    "marvin/store/datatypes"
    "fmt"
)

type Verb uint8

const (
    Invalid Verb = iota
    Register
    
    SetScore
    GetScore
    SetAchievement
    GetAchievements
    GetLeaderBoard
    
    SyncCoins
    DealCoins
    BegCoins
    OfferCoins
    RejectBegCoins

    SendReceiveMessages
    SendTextMessage
    ReceiveTextMessages
    
    UnregisterUser
)

type Command struct {
    verb Verb
}

type cloudResponse struct {
    res string 
}

func (r *cloudResponse) String() string {
    return r.res
}

type Request interface {
    Perform() datatypes.Response
}

type CoinRequest interface {
    Request
}

type UpdateScore interface {
    Request
}

type Store interface {}


type BadRequest struct {
    reason string
}

func (br *BadRequest) String() string {
    return br.reason
}
 
func (br *BadRequest) Perform() datatypes.Response {
    return &cloudResponse{br.reason}
}


type HTTPRequestValidator struct {
}

func (hrv *HTTPRequestValidator) Validate(request *http.Request, json jsondata.JSONString) (cloudReq Request) {
    if jsonMap := jsondata.UnmarshalJSON(json); jsonMap != nil {
        cloudReq = validateRequest(request, jsonMap)
    } else {
        cloudReq = &BadRequest{"{\"reesult\": false, \"answer\": \"bad or unrecognized JSON\"}"}
    }
    return cloudReq
}

type HTTPRequestReader struct {
}

type HTTPResponseWriter struct {
    resWriter http.ResponseWriter
}

func (w *HTTPResponseWriter) Write(res datatypes.Response) {
    io.WriteString(w.resWriter, res.String())
}

func NewHTTPResponseWriter(rw http.ResponseWriter) *HTTPResponseWriter {
    return &(HTTPResponseWriter{rw})
}

func (* HTTPRequestReader) Read(req *http.Request) jsondata.JSONString {
    body := req.Body
    defer body.Close()
    
    var json []byte
    
    // all our messages are at most 1k in size
    jsonChunk := make([]byte, 1024)
    for {
        n, e := body.Read(jsonChunk)
        if n > 8 {
            json = append(json, jsonChunk[:n]...)
        }
        
        if (e != nil || len(json) >= 1024) {
            break
        }
    }

    fmt.Printf("\n[HTTPRequestReader - Received JSON]:\n  %s\n", string(json))
    return json
}

const (
    NoError        int = iota
    OK             int = iota
    BadJSON        int = 1
    InvalidVerb    int = 2
    MissingVerb    int = 4
    UnknownVerb    int = 8
    IncompleteVerb int = 16
    BadParams      int = 32
    DatabaseError  int = 64
)

type RequestHandler func (* jsondata.JSONMap) Request

var reqHandlers map[string]RequestHandler = nil

func InitRequestHandlers() {
    reqHandlers = make(map[string]RequestHandler, 32)
    
    reqHandlers["register"] = validateRegisterRequest, true
    reqHandlers["unregister"] = validateUnregisterRequest, true
    
    reqHandlers["setScore"] = validateSetScore, true
    reqHandlers["getScore"] = validateGetScore, true
    
    reqHandlers["setAchievement"] = validateSetAchievement, true
    reqHandlers["getAchievements"] = validateGetAchievements, true
        
    reqHandlers["requestCoins"] = validateRequestCoins, true
    reqHandlers["offerCoins"] = validateOfferCoins, true
    reqHandlers["syncCoinCounts"] = validateSyncCoinCounts, true
    reqHandlers["getPendingCoinRequests"] = validateGetPendingCoinRequests, true
    reqHandlers["getCoinCount"] = validateGetCoinCountRequest, true

    reqHandlers["getLeaderBoard"] = validateGetLeaderBoard, true

    reqHandlers["sendMessage"] = validateSendMessage, true
    reqHandlers["receiveMessage"] = validateReceiveMessage, true
}

var mapVerbResource = map [string] string {
	"register": "/marvin/registration/", "unregister": "/marvin/unregister/",
	"setScore": "/marvin/scores/", "getScore": "/marvin/scores/",
	"setAchievement": "/marvin/achievements/", "getAchievements": "/marvin/achievements/",
	"requestCoins": "/marvin/coins/", "offerCoins": "/marvin/coins/", "syncCoinCounts": "/marvin/coins/", 
	"getPendingCoinRequests": "/marvin/coins/", "getCoinCount": "/marvin/coins/",
	"getLeaderBoard": "/marvin/leaderboard/",
	"sendMessage": "/marvin/messages/", "receiveMessage": "/marvin/messages/",
}

func isVerbValidForResource(resource string, verb string) bool {
    if path, ok := mapVerbResource[verb]; ok && path == resource {
        return true
    }
    return false
}

func validateRequest(httpReq *http.Request, jsonMap *jsondata.JSONMap) Request {
    var reason string = "missing verb in the request"
    
    if cmd, res := jsonMap.GetString("verb"); res {
        reason = "Invalid or unknown REST-resource and verb combined in the request (" + httpReq.RawURL + " : " + cmd + ")"
        if handler := reqHandlers[cmd]; handler != nil {
            if isVerbValidForResource(httpReq.URL.Path, cmd) {
                return handler(jsonMap)
            }
        }
    }

    msg := "{\"result\": false, \"answer\": \"" + reason + "\"}"
    return &BadRequest{msg}
}

//
// { "verb": "setScore", "hash": <user-id>, "game": <game name>, "score": <unsigned integer> }
// { "result": true, "answer" : "" }
//
func validateSetScore(req *jsondata.JSONMap) Request {
    hash, r1   := req.GetString("hash")
    score, r2  := req.GetUInt("score")
    game, r3   := req.GetString("game")
    
    json := "{\"result\": false, \"answer\": \"missing or invalid arguments in setScore request\"}"
    if (!r1 || !r2 || !r3 || score < 0 || len(hash) < 6 || len(game) < 6) {
        return &BadRequest{json}
    }
    
    return datastore.NewSetScoreRequest(hash, game, score)
}

//
// { "verb": "getScore", "hash": <user-id> , "game": <game-name>}
// { "result": true, "answer" : <score-unsigned-int> }
// { "result": false, "answer" : "invalid user" }
//
func validateGetScore(req *jsondata.JSONMap) Request {
    hash, r1   := req.GetString("hash")
    game, r2   := req.GetString("game")
    
    json := "{\"result\": false, \"answer\": \"missing or invalid arguments in getScore request\"}"
    if (!r1 || !r2 || len(hash) < 6 || len(game) < 6) {
        return &BadRequest{json}
    }
    
    return datastore.NewGetScoreRequest(hash, game)
}

//////

func validateGetAchievements(req *jsondata.JSONMap) Request {
    hash, r1   := req.GetString("hash")
    
    json := "{\"result\": false, \"answer\": \"missing or invalid arguments in getAchievements request\"}"
    if (!r1 || len(hash) < 6) {
        return &BadRequest{json}
    }
    return datastore.NewGetAchievementRequest(hash)
}

func validateSetAchievement(req *jsondata.JSONMap) Request {
    hash, r1   := req.GetString("hash")
    achievement, r2 := req.GetUInt("achievement")
    
    json := "{\"result\": false, \"answer\": \"missing or invalid arguments in setAchievement request\"}"
    if (!r1 || !r2 || len(hash) < 6) {
        return &BadRequest{json}
    }
    return datastore.NewSetAchievementRequest(hash, achievement)
}

//
// { "verb": "getPendingCoinRequests", "hash": <user-id>, "limit": <unsigned integer> }
// { "result": true, "answer" : [zero-or-more{"name": <name>, "count": <unsigned int>}] }
//
func validateGetPendingCoinRequests(req *jsondata.JSONMap) Request {
    hash, r1   := req.GetString("hash")
    limit, r2  := req.GetUInt("limit")
    
    json := "{\"result\": false, \"answer\": \"missing or invalid arguments in getPendingCoinRequests request\"}"
    if (!r1 || !r2 || limit <= 0 || len(hash) < 6) {
        return &BadRequest{json}
    }
    return datastore.NewGetPendingCoinRequests(hash, limit)
}


//
// { "verb": "getCoincount", "name": fb-name }
// { "result": true, "answer" : 6357] }
// { "result": false, "answer" : "not a registered user"] }
//
func validateGetCoinCountRequest(req *jsondata.JSONMap) Request {
    name, r1   := req.GetString("name")
    
    json := "{\"result\": false, \"answer\": \"missing or invalid arguments in getPendingCoinRequests request\"}"
    if (!r1 || len(name) < 6) {
        return &BadRequest{json}
    }
    return datastore.NewGetCoinCount(name)
}


//
// {"verb": "syncCoinCounts", "hash": <user_hash>, "count": <count>}
//
func validateSyncCoinCounts(req *jsondata.JSONMap) Request {
    hash, r1   := req.GetString("hash")
    count, r2  := req.GetUInt("count")
    
    json := "{\"result\": false, \"answer\": \"missing or invalid arguments in syncCoinCount request\"}"
    if (!r1 || !r2 || count <= 0 || len(hash) < 6) {
        return &BadRequest{json}
    }
    return datastore.NewSyncCoinCounts(hash, count)
}

// register: (in) json
//    { "name": "dprasad", "password": "coolVector", 
//      "deviceId":"123456...", "osVersion":"2.2" }
// returns: (out) json 
//    (1) sha-1 key which is the unique id for this person on this device
//    (2) returns additional JSON as getResourceURIs

func validateRegisterRequest(req *jsondata.JSONMap) Request {
    name, r1      := req.GetString("name")
    password, r2  := req.GetString("password")
    device, r3    := req.GetString("deviceId")
    profile, r4   := req.GetString("profile")

    s := "{\"result\": false, \"answer\": \"missing or invalid arguments in the request\"}"
    if (!(r1 && r2 && r3 && r4)) {
        return &BadRequest{s}
    }
    if (len(name) < 6 /*|| len(password) < 2*/ || len(device) < 6) {
        return &BadRequest{s}
    }
    
    displayName, _  := req.GetString("displayName")
    os, _ := req.GetString("os")
    osVer, _ := req.GetString("osVersion")
    density, _ := req.GetString("density")
    portraitX, _ := req.GetUInt("portraitX")
    portraitY, _ := req.GetUInt("portraitY")
    screen, _ := req.GetString("screenSize")
    // TODO: for the moment we have disabled hash generation to avoid JSON issues
    /*hasher := sha1.New()
    hasher.Write([]byte(name + password + device))
    hash := hasher.Sum()*/
    hash := name + "**" + device
    return datastore.NewUserDevice(name, displayName, password, device, hash, profile, os, osVer, portraitX, portraitY, density, screen)
}

// unregister: (in) json
//    { "name": "dprasadm@gmail.com", "hash":"123456..." }
func validateUnregisterRequest(req *jsondata.JSONMap) Request {
    name, r1  := req.GetString("name")
    hash, r2  := req.GetString("hash")

    s := "{\"result\": false, \"answer\": \"missing or invalid arguments in the request\"}"
    if (!r1 || !r2 || len(name) < 6 || len(hash) < 6) {
        return &BadRequest{s}
    }
    return datastore.NewUnregisterRequest(name, hash)
}

//////
func validateRequestCoins(req *jsondata.JSONMap) Request {
    count, r1 := req.GetUInt("count")
    requester, r2 := req.GetString("requester") // hash
    donor, r3 := req.GetString("donor") // fb id
    
    // additionally, we need to record the date and time of this request
    
    if !r1 || !r2 || !r3 || len(requester) < 6 || len(donor) < 6 || count <= 0 {
        jsonRes := "{\"result\": false, \"answer\": \"Missing or Invalid data in requestCoins request\"}"
        return &BadRequest{jsonRes}
    }
    return datastore.NewCoinsRequest(requester, donor, count)
}

///////

func validateOfferCoins(req *jsondata.JSONMap) Request {
    offer, r1 := req.GetUInt("offer")
    countOnDevice, r2 := req.GetUInt("countOnDevice")
    requester, r3 := req.GetString("requester") // fb id
    donor, r4 := req.GetString("donor") // hash 
    
    // additionally, we need to record the date and time of this response
    
    if !r1 || !r2 || !r3 || !r4 || len(donor) < 6 || len(requester) < 6 || offer <= 0  || countOnDevice <= 0 {
        jsonRes := "{\"result\": false, \"answer\": \"Missing or Invalid data in offerCoins request\"}"
        return &BadRequest{jsonRes}
    }
    
    return datastore.NewCoinsOffer(requester, donor, offer, countOnDevice)
}

func validateGetLeaderBoard(req *jsondata.JSONMap) Request {
    hash, r1 := req.GetString("hash")
    
    if !r1 || len(hash) < 6 {
        jsonRes := "{\"result\": false, \"answer\": \"Missing or Invalid data in getLeaderBoard request\"}"
        return &BadRequest{jsonRes}
    }
    
    return datastore.NewLeaderBoardRequest(hash)
}

func validateSendMessage(req *jsondata.JSONMap) Request {
    from, r1 := req.GetString("from") // hash value
    to, r2   := req.GetString("to")   // fb id
    msg, r3  := req.GetString("message")
    
    // additionally, we need to record the date and time of this response
    
    if !r1 || !r2 || !r3 || to == from || len(from) < 6 || len(to) < 6 || len(msg) == 0 {
        jsonRes := "{\"result\": false, \"answer\": \"Missing or Invalid data in sendMessage request\"}"
        return &BadRequest{jsonRes}
    }
    
    return datastore.NewSendMessage(from, to, msg)
}


func validateReceiveMessage(req *jsondata.JSONMap) Request {
    to, r1 := req.GetString("receiver") // hash of the receiver
    limit, _ := req.GetUInt("limit") // ignored, at the moment
    
    // additionally, we need to record the date and time of this response
    
    if (!r1 || len(to) < 6) {
        jsonRes := "{\"result\": false, \"answer\": \"Missing or Invalid data in receiveMessage request\"}"
        return &BadRequest{jsonRes}
    }
    
    return datastore.NewReceiveMessage(to, limit)
}

