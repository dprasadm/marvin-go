package cloud

import "crypto/sha1"
import "marvin/json/jsondata"
import "fmt"

/*
type Verb uint8

const (
    Invalid Verb = iota
    Register
    SignIn
    CheckUpdates
    GetResourceURIs
    Sync
    SyncMetrics
    BegCoins
    OfferCoins
    RejectBegCoins
    GetLeaderBoard
    VerifyOnlineStatus
    SendTextMessage
    ReceiveTextMessages
        
)
*/

type Error uint8
const (
    InvalidVerb Error = 2
    MissingVerb Error = 4
)

type Response struct {
}

type RequestHandler func (* jsondata.JSONMap) *Response

var reqHandlers map[string]RequestHandler = nil

func InitRequestHandlers() {
    reqHandlers = make(map[string]RequestHandler, 32)
    
    reqHandlers["register"] = ValidateRegisterRequest, true
    reqHandlers["signin"] = signin, true
    
    reqHandlers["checkUpdates"] = checkUpdates, true
    reqHandlers["getResourceURIs"] = getResourceURIs, true
    
    reqHandlers["receiveTextMessage"] = receiveTextMessages, true
}

func ValidateRequest(req *jsondata.JSONMap) *Response {
    err := MissingVerb
    
    if cmd, res := req.GetString("verb"); res {
        err = InvalidVerb
        if handler := reqHandlers[cmd]; handler != nil {
            return handler(req)
        }
    }
    panic(err)
}


// register: (in) json
//    { "name": "dprasad", "password": "coolVector", 
//      "deviceId":"123456...", "osVersion":"2.2" }
// returns: (out) json 
//    (1) sha-1 key which is the unique id for this person on this device
//    (2) returns additional JSON as getResourceURIs

func ValidateRegisterRequest(req *jsondata.JSONMap) *Response {
    name, r1      := req.GetString("name")
    password, r2  := req.GetString("password")
    device, r3    := req.GetString("deviceId")

    if (r1 && r2 && r3) {
        osVer, _ := req.GetString("osVersion")
        hasher := sha1.New()
        hasher.Write([]byte(name + password + device))
        fmt.Printf("osVer: %v, hash size: %v, hash sum: %v\n", 
                    osVer, hasher.Size(), hasher.Sum())
        return nil
    }
    
    panic("Ill-formed JSON for register request")
}

func signin(req *jsondata.JSONMap) *Response {
    return nil
}

func getUID(req *jsondata.JSONMap) string {
    if uid, r := req.GetString("uid"); r {
        return uid
    }
    panic("Fatal Error: missing UID in the request")
}

// screen sizes
const (
    InvalidScreen int = iota
    Small
    Normal
    Large
    XtraLarge
)
func screenSize(s string) int {
    switch s {
    case "small"  : return Small
    case "normal" : return Normal
    case "large"  : return Large
    case "xlarge" : return XtraLarge
    }
    return InvalidScreen
}

const (
    InvalidDensity int = iota
    LDPI 
    MDPI 
    HDPI 
    XDPI
)

func screenDensity(d string) int {
    switch d {
    case "ldpi" : return LDPI
    case "mdpi" : return MDPI
    case "hdpi" : return HDPI
    case "xdpi" : return XDPI
    }
    return InvalidDensity
}

type CheckUpdateRequestParams struct {
    uid string
    verApp, verRes string
    portraitX, portraitY int
    density, screenSize int
}

func parseUpdateParams(req* jsondata.JSONMap) *CheckUpdateRequestParams {
    uid := getUID(req)
    verApp, r1 := req.GetString("verApp")
    verRes, r2 := req.GetString("verRes")
    if (!r1 || !r2) { return nil }
    
    portraitX := 0
    portraitY := 0
    density := 0
    size := InvalidScreen
    
    portraitX, _ = req.GetUInt("portraitX")
    portraitY, _ = req.GetUInt("portraitY")
    if (portraitX == 0 || portraitY == 0) {
        return nil
    }
    
    d, _ := req.GetString("density")
    density = screenDensity(d)
    d, _ = req.GetString("screenSize")
    size = screenSize(d)
    
    p := CheckUpdateRequestParams{uid, verApp, verRes, 
                                  portraitX, portraitY, 
                                  density, size}
    return &p
    
}

func checkUpdates(req* jsondata.JSONMap) (r *Response) {
    r = nil
    if p:= parseUpdateParams(req); p != nil {
        r = doCheckUpdates(p)
    }
    
    return r
}

func doCheckUpdates(p *CheckUpdateRequestParams) (r *Response) {
    r = nil
    fmt.Printf("%v, %v, %v, %d, %d, %d, %d\n", p.uid, p.verApp, p.verRes, 
                                   p.portraitX, p.portraitY,
                                   p.density, p.screenSize)
    
    return r
}


func getResourceURIs(req* jsondata.JSONMap) (r *Response) {
    r = nil
    if p:= parseUpdateParams(req); p != nil {
        r = doGetResourceURIs(p)
    }
    
    return r
}

func doGetResourceURIs(p *CheckUpdateRequestParams) (r *Response) {
    r = nil
    fmt.Printf("%v, %v, %v, %d, %d, %d, %d\n", p.uid, p.verApp, p.verRes, 
                                   p.portraitX, p.portraitY,
                                   p.density, p.screenSize)
    return r
}

func receiveTextMessages(req *jsondata.JSONMap) *Response {
    return nil
}

