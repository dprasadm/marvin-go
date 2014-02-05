package datastore

import (
    "fmt"
    "marvin/store/datatypes"
    "launchpad.net/gobson/bson"
    "launchpad.net/mgo"
    "json"
    "strconv"
)

// FIXME: this has to be passed as a configuration parameter to Marvin
const URL_127_0_0_1            =    "127.0.0.1"
const GUSTO_DB_NAME            =    "Gusto"
const REGISTERED_USERS         =    "RegisteredUsers"
const REGISTERED_DEVICES       =    "RegisteredDevices"
const PENDING_COIN_REQUESTS    =    "OpenCoinRequests"
const GAME_SCORES              =    "GameScores"
const TEXT_MESSAGES            =    "TextMessages"

type DBSession struct {
    url string
    * mgo.Session
}

func NewSession() (session *DBSession) {
    session = nil
    if s, err := mgo.Mongo(URL_127_0_0_1); err == nil {
        s.SetSafe(&mgo.Safe{})
        session = &DBSession{URL_127_0_0_1, s}
    }
    
    return session
}

type UserDevice struct {
    datatypes.User
    datatypes.Device
}

func NewUserDevice(name string, displayName string, password string, device string, hash string, profile string, os string, osVer string, portraitX int, portraitY int, density string, screen string) *UserDevice {
    //var achievements []string
    return &UserDevice{datatypes.User{name, displayName, password, device, hash, profile, 0, 0, 0, 0}, datatypes.Device{device, os, osVer, portraitX, portraitY, density, screen}}
}

func (ud *UserDevice) Perform() datatypes.Response {
    //fmt.Println("Registration::Perform --> ", reg.Name, reg.Password)
    jsonRes := "{\"result\": false, \"answer\": \"User registration attempt failed\" }"
    
    dataStore := NewSession()
    defer dataStore.Close()
    
    // on success, answer indicates the hash value. 
    // on failure, answer contains the error message 
    answer, status := dataStore.registerNewUser(ud);
    if status == true {
       jsonRes = "{\"result\": true, \"answer\": " + answer + "}"
    }
    
    return &datatypes.GenericResponse{jsonRes}
}


func (session *DBSession) registerUser(user *datatypes.User) (r bool) {
    r = false
    c := session.DB(GUSTO_DB_NAME).C(REGISTERED_USERS);
    query := c.Find(bson.M{"hash": user.Hash})
    count, err := query.Count()
    if err == nil && count == 0 { /* no hash duplicates */
        query = c.Find(bson.M{"name": user.Name})
        count, err = query.Count()
        if err == nil && count == 0 { /* no name duplicates */
	        if err = c.Insert(user); err == nil {
	            r = true  /* everything's fine! */
	        }
	    }
    }
    return r
}

// we allow multiple registrations for the same device
func (session *DBSession) registerDevice(device *datatypes.Device) (r bool) {
    r = true
    c := session.DB(GUSTO_DB_NAME).C(REGISTERED_DEVICES)
    query := c.Find(bson.M{"deviceid":device.DeviceId})
    count, err := query.Count()
     /* We DO NOT upset an existing entry (if count > 0) */
    if err == nil && count == 0 {
        if e := c.Insert(device); e == nil {
            r = true
        } else {
            r = false
        }
    }
    
    return r
}

func (s *DBSession) registerNewUser(ud *UserDevice) (answer string, r bool) {
    r = false
    answer = "\"Generic DataStore Error\""
    if s.registerUser(&ud.User) && s.registerDevice(&ud.Device) {
        r = true
        answer = fmt.Sprintf("\"%s\"", string(ud.Hash))
    }
    
    return answer, r
}

///////
type UnregisterRequest struct {
    name string
    hash string
}

func NewUnregisterRequest(name string, hash string) *UnregisterRequest {
    return &UnregisterRequest{name, hash}
}

func (req *UnregisterRequest) Perform() datatypes.Response {
    jsonRes := "{\"result\": false, \"answer\": \"wrong or invalid arguments in the request\" }"
    
    ds := NewSession()
    defer ds.Close()
    
    if !ds.validUserHash(req.hash) {
        return &datatypes.GenericResponse{jsonRes}
    }
    
    jsonRes = "{\"result\": true, \"answer\": \"Generic datastore error\" }"
   
    users := ds.DB(GUSTO_DB_NAME).C(REGISTERED_USERS);
    user := bson.M{"hash": req.hash, "name": req.name}
    n, err := users.Find(user).Count()
    if err == nil && n == 1 {
        users.Remove(user)
        jsonRes = "{\"result\": true, \"answer\": \"\" }"
        fmt.Printf("User %s unregistered!", req.name)
    }
    
    return &datatypes.GenericResponse{jsonRes}
}


//////
type GetScoreRequest struct {
    hash string
    game string
}

type SetScoreRequest struct {
    hash string
    game string
    score int
}

func NewGetScoreRequest(hash string, game string) *GetScoreRequest {
   return &GetScoreRequest{hash, game}
}

func NewSetScoreRequest(hash string, game string, score int) *SetScoreRequest {
   return &SetScoreRequest{hash, game, score}
}

func (req *GetScoreRequest) Perform() datatypes.Response {
    jsonRes := "{\"result\": false, \"answer\": \"Not a registered user\"}"
    
    ds := NewSession()
    defer ds.Close()

    if !ds.validUserHash(req.hash) {
        return &datatypes.GenericResponse{jsonRes}
    }
    
    scores := ds.DB(GUSTO_DB_NAME).C(GAME_SCORES);
    var gameScore interface{}
    err := scores.Find(bson.M{"hash": req.hash, "game": req.game}).
                     Select(bson.M{"score": 1, "_id": 0}).One(&gameScore)
                     
    jsonRes = "{\"result\": false, \"answer\": \"Unknown game name in the request\"}"
    if  err == nil {
        if data, ok := gameScore.(bson.M); ok {
            score := data["score"].(int)
            jsonRes = "{\"result\": true, \"answer\": " + strconv.Itoa(score) + "}"
        }
    }
    
    return &datatypes.GenericResponse{jsonRes}
}

func (req *SetScoreRequest) Perform() datatypes.Response {
    jsonRes := "{\"result\": false, \"answer\": \"Not a registered user\"}"
    
    ds := NewSession()
    defer ds.Close()

    if !ds.validUserHash(req.hash) {
        return &datatypes.GenericResponse{jsonRes}
    }
    
    jsonRes = "{\"result\": false, \"answer\": \"Unknown error\"}"
    scorees := ds.DB(GUSTO_DB_NAME).C(GAME_SCORES);
    change := mgo.Change{Update: bson.M{"$set": bson.M{"score": req.score}}, Upsert: true}
    
    var result interface{}
    err := scorees.Find(bson.M{"hash": req.hash, "game": req.game}).Modify(change, &result)
    if err == nil {
        jsonRes = "{\"result\": true, \"answer\": \"\"}"
    }
    
    return &datatypes.GenericResponse{jsonRes}
}

//////
type GetAchievementsRequest struct {
    hash string
}

func NewGetAchievementRequest(hash string) *GetAchievementsRequest {
    return &GetAchievementsRequest{hash}
}

type SetAchievementRequest struct {
    hash string
    achievement int
}

func NewSetAchievementRequest(hash string, achieved int) *SetAchievementRequest {
    return &SetAchievementRequest{hash, achieved}
}

func (req *GetAchievementsRequest) Perform() datatypes.Response {
    ds := NewSession()
    defer ds.Close()
    
    isHash := false
    isName := ds.validUser(req.hash)
    if !isName {
        isHash = ds.validUserHash(req.hash)
    }
    if !isName || !isHash {
        jsonRes := "{\"result\": false, \"answer\": \"Not a registered user\"}"
        return &datatypes.GenericResponse{jsonRes}
    }
    
    users := ds.DB(GUSTO_DB_NAME).C(REGISTERED_USERS);
    var result interface{}
    query := users.Find(bson.M{"name": req.hash})
    if (isHash) {
        query = users.Find(bson.M{"hash": req.hash})
    }
    err := query.Select(bson.M{"achievements": 1, "_id": 0}).One(&result)
    jsonRes := "{\"result\": false , \"answer\": \"Unknown error\"}"
    if err == nil {
        if mapp, ok := result.(bson.M); ok {
            achievements := mapp["achievements"]
            fmt.Printf("    GetAchievementsRequest: %v\n", achievements)
            if data, err := json.Marshal(achievements); err == nil {
                jsonRes = "{\"result\": true , \"answer\":" + string(data) + "}"
            }
        }
    }
    
    fmt.Printf("    %s\n", jsonRes)
    return &datatypes.GenericResponse{jsonRes}
}
    
func (req *SetAchievementRequest) Perform() datatypes.Response {
    jsonRes := "{\"result\": false, \"answer\": \"Not a registered user\"}"
    
    ds := NewSession()
    defer ds.Close()

    if !ds.validUserHash(req.hash) {
        return &datatypes.GenericResponse{jsonRes}
    }
    
    users := ds.DB(GUSTO_DB_NAME).C(REGISTERED_USERS);
    user := users.Find(bson.M{"hash": req.hash})
    
    change := mgo.Change{Update: bson.M{"$set": bson.M{"achievements": 
                                                             req.achievement}}}
    var doc interface{}
    if user.Modify(change, &doc) == nil {
        jsonRes = "{\"result\": true, \"answer\": \"set achievement succeeded\"}"
    } else {
        jsonRes = "{\"result\": false, \"answer\": \"Unknown error\"}"
    }
    
    return &datatypes.GenericResponse{jsonRes}
}

/////
type GetCoinCountRequest struct {
    name string
}

func NewGetCoinCount(name string) *GetCoinCountRequest {
    return &GetCoinCountRequest{name}
}

func (req *GetCoinCountRequest) Perform() datatypes.Response {
    jsonRes := "{\"result\": false, \"answer\": \"Not a registered user\"}"
    
    ds := NewSession()
    defer ds.Close()

    if !ds.validUser(req.name) {
        return &datatypes.GenericResponse{jsonRes}
    }
    
    users := ds.DB(GUSTO_DB_NAME).C(REGISTERED_USERS);
    query := users.Find(bson.M{"name": req.name}).Select(bson.M{"borrowed": 1, "earned": 1, "_id": 0})
    
    var result interface{}
    
    jsonRes = "{\"result\": false, \"answer\": 0 }"
    if query.One(&result) == nil {
        // coin count = borrowed + earned
        if data, ok := result.(bson.M); ok {
            borrowed := data["borrowed"].(int)
            earned := data["earned"].(int)
            count := borrowed + earned
            jsonRes = "{\"result\": true, \"answer\": " + strconv.Itoa(count) + "}"
        }
    }
    
    return &datatypes.GenericResponse{jsonRes}
}

//////
type GetPendingCoinRequest struct {
    hash string
    limit int
}

func NewGetPendingCoinRequests(hash string, limit int) *GetPendingCoinRequest {
    return &GetPendingCoinRequest{hash, limit}
}

//
// { "verb": "getPendingCoinRequests", "hash": <user-id>, "limit": <unsigned integer> }
// { "result": true, "answer" : [zero-or-more{"name": <name>, "count": <unsigned int>}] }
//
func (req *GetPendingCoinRequest) Perform() datatypes.Response {
    //jsonRes := "{\"result\": false, \"answer\": []"}"
    
    ds := NewSession()
    defer ds.Close()

    if !ds.validUserHash(req.hash) {
        jsonRes := "{\"result\": false, \"answer\": []}"
        return &datatypes.GenericResponse{jsonRes}
    }
    
    coinRequests := ds.DB(GUSTO_DB_NAME).C(PENDING_COIN_REQUESTS);
    requestsForDonor := coinRequests.Find(bson.M{"donorhash": req.hash})
    //query := requests.Find(bson.M{"donorhash": req.hash}).
    query := requestsForDonor.
                         Select(bson.M{"_id": 0, "requester": 1, "ask" : 1})
    count, err := query.Count()
    if err == nil && count == 0 {
        jsonRes := "{\"result\": true, \"answer\": []}"
        return &datatypes.GenericResponse{jsonRes}
    }
    fmt.Printf("GetPendingCoinRequests: %d\n", count)
    
    mapRes := "[ "
    if iter, err := query.Iter(); err == nil {
        var result interface{}
        firstItem := true
        
        for {
            result = nil
            if err = iter.Next(&result); err != nil {
                break
            }
            if data, err := json.Marshal(result); err == nil {
                if !firstItem { mapRes += ", " }
                mapRes += string(data)
                firstItem = false
            }
        }
        // DO NOT DO THE FOLLOWING, STUPID!! 
        // remove the requests from the collection
        coinRequests.RemoveAll(bson.M{"donorhash": req.hash})
    }
    mapRes += " ]"
    
    jsonRes := "{\"result\": true, \"answer\":" + mapRes + "}"
    
    fmt.Printf("    %s\n", jsonRes)
    return &datatypes.GenericResponse{jsonRes}
}

///////
type SyncCoinCounts struct {
    hash string
    count int
}

func NewSyncCoinCounts(hash string, count int) *SyncCoinCounts {
    return &SyncCoinCounts{hash, count}
}

func (req *SyncCoinCounts) Perform() datatypes.Response {
    jsonRes := "{\"result\": false, \"answer\": \"Not a registered user\"}"

    ds := NewSession()
    defer ds.Close()

    users := ds.DB(GUSTO_DB_NAME).C(REGISTERED_USERS);
    query := users.Find(bson.M{"hash": req.hash})
    var user interface{}
    
    if query.One(&user) == nil {
        // new coin count = borrowed + req.count
        if data, ok := user.(bson.M); ok {
            jsonRes = "{\"result\": false, \"answer\": \"Generic Error\"}"
            borrowed := data["borrowed"].(int)
            earned := borrowed + req.count
            fmt.Printf("SyncCoinCounts: replacing %d with %d earned coins\n", data["earned"].(int), earned)            
            change := mgo.Change{Update: bson.M{"$set": bson.M{"borrowed": 0, "earned": earned}}}
            var doc interface{}
            if users.Find(bson.M{"hash": req.hash}).Modify(change, &doc) == nil {
                jsonRes = "{\"result\": true, \"newCount\": " + strconv.Itoa(earned) + ", \"answer\": \"\"}"
            }
        }
    }
    
    return &datatypes.GenericResponse{jsonRes}
}

type CoinsAsk struct {
    datatypes.CoinsRequest
}

func NewCoinsRequest(requester /* hash */string, donor string, ask int) *CoinsAsk {
    return &CoinsAsk{datatypes.CoinsRequest{requester, "", "", donor, ask}}
}

func (req *CoinsAsk) Perform() datatypes.Response {
    jsonRes := "{\"result\": false, \"answer\": \"\"}"

    dataStore := NewSession()
    defer dataStore.Close()

    fmt.Printf("Coin Reqester(Hash): %v, donor: %s\n", req.RequesterHash, req.Donor)
    
    if !(dataStore.validUser(req.Donor) && dataStore.validUserHash(req.RequesterHash)) {
        jsonRes = "{\"result\": false, \"answer\": \"Not a registered requester or recipient in the request\"}"
        return &datatypes.GenericResponse{jsonRes}
    }
    
    req.Requester = dataStore.UserNameFromHash(req.RequesterHash)
    req.DonorHash = dataStore.HashFromUserName(req.Donor)
    
    c := dataStore.DB(GUSTO_DB_NAME).C(PENDING_COIN_REQUESTS);
    query := c.Find(bson.M{"requesterhash": req.RequesterHash, "donor": req.Donor})
    count, err := query.Count()
    if err == nil && count == 0 { /* no old, outstanding requests */
    /* TODO: instead of inserting, should we consider updating the count? */
        if e := c.Insert(req.CoinsRequest); e == nil {
            fmt.Printf("Coin Request sent successfully\n")
            jsonRes = "{\"result\": true, \"answer\": \"coin request sent successfully\"}"
        } else {
            fmt.Printf("attempt to insert coin request failed!")
            jsonRes = "{\"result\": false, \"answer\": \"Failed to insert Coin Request.\"}"
        }
    } else {
        fmt.Printf("Coin Request rejected because there are pending requests\n")
        jsonRes = "{\"result\": false, \"answer\": \"Request rejected. Pending coin requests. \"}"
    }
    
    return &datatypes.GenericResponse{jsonRes}
}

type CoinsOffer struct {
    datatypes.CoinsOffer
}

func NewCoinsOffer(requester string, donorHash /* hash */string, offer int, earnedOnDevice int) *CoinsOffer {
    return &CoinsOffer{datatypes.CoinsOffer{"", requester, donorHash, "", offer, earnedOnDevice}}
}

// check if donor has enough coins in his account
func HaveAdequaueCoins(ask int, hash string, donor string, ds *DBSession) (bool, int, int) {
    haveAdequateCoins, coinsEarned, coinsBorrowed := false, 0, 0

    var user interface{}
    coinCounts := bson.M{"borrowed": 1, "earned": 1, "_id": 0}
    users := ds.DB(GUSTO_DB_NAME).C(REGISTERED_USERS)
    users.Find(bson.M{"hash": hash}).Select(coinCounts).One(&user)
    
    if r, ok := user.(bson.M); ok {
        totalCoins := r["earned"].(int) + r["borrowed"].(int)
        if totalCoins > ask { //NOTE: we fail when (totalCoins <= ask)
            haveAdequateCoins = true
            coinsEarned = (totalCoins - ask)
            coinsBorrowed = 0
        }
    }
    return haveAdequateCoins, coinsEarned, coinsBorrowed
}

func (req *CoinsOffer) Perform() datatypes.Response {
    ds := NewSession()
    defer ds.Close()

    if !(ds.validUserHash(req.DonorHash) && ds.validUser(req.Requester)) {
        jsonRes := "{\"result\": false, \"answer\": \"Unknown users specified in the request\"}"
        return &datatypes.GenericResponse{jsonRes}
    }

    donorName := ds.UserNameFromHash(req.DonorHash)
    requesterHash := ds.HashFromUserName(req.Requester)
    
    // first sync the "earned" coin count with that on the device
    var result interface{} = nil
    users := ds.DB(GUSTO_DB_NAME).C(REGISTERED_USERS)
    changeEarned := mgo.Change{Update: bson.M{"$set": bson.M{"earned": req.EarnedOnDevice}}}
    users.Find(bson.M{"name": donorName, "hash": req.DonorHash}).Modify(changeEarned, &result)
    
    // next see if this user can really offer these many coins
    haveCoins, updatedEarned, updatedBorrowed := HaveAdequaueCoins(req.Offer, req.DonorHash, donorName, ds)
    
    jsonRes := fmt.Sprintf("{\"result\": false, \"requested\": %d, \"available\": %d, \"answer\": \"not enough coins\"}", req.Offer, updatedEarned + updatedBorrowed)

    // well, transfer coins!    
    if (haveCoins) {
        // remove the coin request 
        coinRequests := ds.DB(GUSTO_DB_NAME).C(PENDING_COIN_REQUESTS)
        coinRequests.Remove(bson.M{"requester": req.Requester, "donorhash": req.DonorHash})
        
        // credit it to the "borrow" count on the receiver. DO NOT change "earned" now. 
        credit := mgo.Change{Update: bson.M{"$inc": bson.M{"borrowed": req.Offer}}}
        // set "borrowed" to Zero and keep "earned" honest for the donor!
        debit := mgo.Change{Update: bson.M{"$set": bson.M{"earned": updatedEarned, "borrowed": updatedBorrowed}}}
        
        result = nil // THIS IS ABSOLUTELY REQUIRED!!
        err := users.Find(bson.M{"name": req.Requester, "hash": requesterHash}).Modify(credit, &result)
                            
        result = nil // THIS IS ABSOLUTELY REQUIRED!!
        err = users.Find(bson.M{"name": donorName, "hash": req.DonorHash}).Modify(debit, &result)
       
        if err == nil {
            jsonRes = fmt.Sprintf("{\"result\": true, \"offered\": %d, \"newCount\": %d, \"answer\": \"\" }", req.Offer, updatedEarned)
        }
    }
    
    return &datatypes.GenericResponse{jsonRes}
}

func (ds *DBSession) validUser(name string) bool {
    res := false
    c := ds.DB(GUSTO_DB_NAME).C(REGISTERED_USERS)
    query := c.Find(bson.M{"name":name})
    if n, e := query.Count(); e == nil && n == 1 { 
        res = true 
    } else {
        fmt.Printf("Invalid User request : %s\n", name)
    }
    return res
}

func (ds *DBSession) validUserHash(hash string) bool {
    res := false
    c := ds.DB(GUSTO_DB_NAME).C(REGISTERED_USERS)
    query := c.Find(bson.M{"hash": hash})
    if n, e := query.Count(); e == nil && n == 1 { 
        res = true 
    } else {
        fmt.Printf("Invalid user hash in JSON request: %v", hash)
    }
    
    return res
}

func (ds *DBSession) HashFromUserName(name string) (hash string) {
    var record interface{}
    hash = ""
    
    c := ds.DB(GUSTO_DB_NAME).C(REGISTERED_USERS)
    query := c.Find(bson.M{"name": name})
    if n, e := query.Count(); e == nil && n == 1 { 
        err := query.Select(bson.M{"hash": 1, "_id": 0}).Distinct("hash", &record)
        if err == nil {
            if r, ok := record.([]interface{}); ok {
	            if data, e := json.Marshal(r[0]); e == nil && len(data) > 2 {
	                hash = string(data[1:len(data)-1])
	                fmt.Printf("HashFromUserName: %s --> %s\n", hash, name)
	            }
	        }
        }
    } else {
        fmt.Printf("Invalid user hash in JSON request: %v", hash)
    }
    
    return hash
}


func (ds *DBSession) UserNameFromHash(hash string) (name string) {
    var record interface{}
    name = ""
    
    c := ds.DB(GUSTO_DB_NAME).C(REGISTERED_USERS)
    query := c.Find(bson.M{"hash": hash})
    if n, e := query.Count(); e == nil && n == 1 {
        err := query.Select(bson.M{"name": 1, "_id": 0}).Distinct("name", &record)
        if err == nil {
            if r, ok := record.([]interface{}); ok {
	            if data, e := json.Marshal(r[0]); e == nil && len(data) > 2 {
	                name = string(data[1:len(data)-1])
	                fmt.Printf("UserNameFromHash: %s --> %s\n", hash, name)
	            } else {
	                fmt.Printf("UserNameFromHash[Failed 3]: %d, %v\n", n, e)
	            }
	        }
        } else {
            fmt.Printf("UserNameFromHash[Failed 2]: %d, %v\n", n, err)
        }
    } else {
        fmt.Printf("UserNameFromHash[Failed 1]: %d, %v\n", n, e)
    }
    
    if name == "" {
        fmt.Printf("UserNameFromHash Error: user with provided hash doesn't exist. %s\n", hash)
    }
    
    return name
}

/////////

type sendMessage struct {
    //datatypes.ShortTextMessage
    from string
    to string
    text   string
}

type receiveMessage struct {
    to string
    limit int
}

func NewSendMessage(from string, to string, msg string) *sendMessage {
    //return &SendMessage{datatypes.ShortMessage{from, to, msg}}
    return &sendMessage{from, to, msg}
}

func NewReceiveMessage(to string, limit int) *receiveMessage {
    return &receiveMessage{to, limit}
}

func (req *sendMessage) Perform() datatypes.Response {
    ds := NewSession()
    defer ds.Close()

    jsonRes := "{\"result\": false, \"answer\": \"\"}"
    if (!ds.validUserHash(req.from) || !ds.validUser(req.to) || req.from == req.to) {
        jsonRes = "{\"result\": false, \"answer\": \"Invalid user specified in sendMessage request\"}"
        return &datatypes.GenericResponse{jsonRes}
    }

    messages := ds.DB(GUSTO_DB_NAME).C(TEXT_MESSAGES)
    senderName := ds.UserNameFromHash(req.from)
    if messages.Insert(bson.M{"from": senderName, "to": req.to, "msg": req.text}) == nil {
        jsonRes = "{\"result\": true, \"answer\": \"message sent\"}"
    } else {
        jsonRes = "{\"result\": false, \"answer\": \"Unknown error while attempting to send the message\"}"
    }
    return &datatypes.GenericResponse{jsonRes}
}

func (req *receiveMessage) Perform() datatypes.Response {
    jsonRes := "{\"result\": false}"

    ds := NewSession()
    defer ds.Close()

    if (!ds.validUserHash(req.to)) {
        jsonRes = "{\"result\": false, \"answer\": \"Invalid user specified in receiveMessage request\"}"
        return &datatypes.GenericResponse{jsonRes}
    }

    messageCollection := ds.DB(GUSTO_DB_NAME).C(TEXT_MESSAGES)
    receiver := ds.UserNameFromHash(req.to)
    // The "_id" field is included by default. We must exclude it specifically.
    query := messageCollection.Find(bson.M{"to" : receiver}).
                               Select(bson.M{"from": 1, "msg": 1,"_id": 0})
    if c, err := query.Count(); err != nil || c == 0 {
        jsonRes = "{\"result\": true, \"answer\": []}"
        return &datatypes.GenericResponse{jsonRes}
    }

    mapRes := "[ "
    if iter, err := query.Iter(); err == nil {
        var result interface{}
        firstItem := true
        
        for {
            result = nil
            if err = iter.Next(&result); err != nil {
                break
            }
            if data, err := json.Marshal(result); err == nil {
                if !firstItem { mapRes += ", " }
                mapRes += string(data)
                firstItem = false
            }
        }
        
        // remove the messages from the collection
        messageCollection.RemoveAll(bson.M{"to": receiver})
    }
    mapRes += " ]"
    
    jsonRes = "{\"result\": true, \"answer\":" + mapRes + "}"
    return &datatypes.GenericResponse{jsonRes}
}

type leaderBoardRequest struct {
    hash string
}

func NewLeaderBoardRequest(hash string) *leaderBoardRequest {
    return &leaderBoardRequest{hash}
}

func (req *leaderBoardRequest) Perform() datatypes.Response {
    ds := NewSession()
    defer ds.Close()

    if (!ds.validUserHash(req.hash)) {
        jsonRes := "{\"result\": false, \"answer\": \"Invalid user specified in getLeaderBoard request\"}"
        return &datatypes.GenericResponse{jsonRes}
    }

    gameScores := ds.DB(GUSTO_DB_NAME).C(GAME_SCORES)
    
    var games []string
    if err := gameScores.Find(nil).Distinct("game", &games); err != nil {
        jsonRes := "{\"result\": false, \"answer\": \"Generic datastore error\"}"
        return &datatypes.GenericResponse{jsonRes}
    }
    
    mainResultStr := "true"
    mainAnswerStr := "[ "
    firstGameRecord := true
    strGamePlayerScore := ""
    //
    // {result: true, answer: [{"game": "Boondh", scores: [{"player": "siddu", "score": 80}, {"player": "shamanth", "score": 300}, ..., ] }, {"game": "Three Monkeys", "scores": [{}, {}, {}] }]
    
    for i := 0; i < len(games); i++ {
        //fmt.Printf("collecting top 5 scorers for [%s]\n", games[i])
        
        queryTopFiveScores := gameScores.Find(bson.M{"game": games[i]}).
                                      Sort(bson.M{"score": -1}).Limit(5).
                                          Select(bson.M{"hash": 1, "score": 1, "_id": 0})
        
        mapPlayerScore := "[ "
        strPlayerScoreRecord := ""
        
        if iter, err := queryTopFiveScores.Iter(); iter != nil && err == nil {
            var result interface{}
            firsPlayerScoreRecord := true
            for {
                result = nil
                if err = iter.Next(&result); err != nil {
                    break
                }
                if mapp, ok := result.(bson.M); ok {
                    score, r1 := mapp["score"].(int)
                    hash, r2 := mapp["hash"].(string)
                    playerName := ds.UserNameFromHash(hash) // can be "" if user has unregistered!
                    if r1 && r2 && playerName != "" {
                        if !firsPlayerScoreRecord { strPlayerScoreRecord += ", " }
                        strPlayerScoreRecord += "{ \"player\": \"" + 
                                                   playerName + 
                                                   "\", \"score\": " + 
                                                   strconv.Itoa(score) + "}"
                        firsPlayerScoreRecord = false
                    }
                }
            }
            //fmt.Printf("\nPlayerScore: %s\n", strPlayerScoreRecord)
        }
        
        if !firstGameRecord { strGamePlayerScore += ", " }
        mapPlayerScore += (strPlayerScoreRecord + " ]")
        strGamePlayerScore += "{\"game\": \"" + games[i] + "\", \"scores\": " + mapPlayerScore + "}"
        firstGameRecord = false
        
        fmt.Printf("\nGamePlayerScore: %s\n", strGamePlayerScore)
    }
    
    mainAnswerStr += strGamePlayerScore + " ]"
    mainJsonRes := "{\"result\": " + mainResultStr + ", \"answer\":" + mainAnswerStr + "}"
    
    fmt.Printf("\nGetLeaderBoard -->: %s\n", mainJsonRes)
    
    return &datatypes.GenericResponse{mainJsonRes}
}

