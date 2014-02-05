package datatypes

// date registered is added while inserting the record
type User struct {
    Name        string
    DisplayName string
    Password    string
    DeviceId    string
    Hash        string     //bson.Binary
    Profile     string     // "fb" - facebook, "tw" - twitter, "99" - us!
    Score       int
    Borrowed    int
    Earned      int
    Achievements int //[]string
}

// we may want to add location and other information
type Device struct {
    DeviceId    string
    OS          string
    OSVer       string
    PortraitX   int
    PortraitY   int
    Density     string
    ScreenSize  string
}

type CoinsRequest struct {
    RequesterHash   string /* hash id */
    Requester   string /* requester's fb id */
    DonorHash   string
    Donor       string /* fb id */
    Ask         int    /* # Count requested coins */
}

type CoinsOffer struct {
    RequesterHash   string /* hash id */
    Requester   string /* requester's fb id */
    DonorHash   string
    Donor       string /* fb id */
    Offer       int    /* # Count requested coins */
    EarnedOnDevice     int
}

// Response must return a JSON encoded string
type Response interface {
    String() string
}

type GenericResponse struct {
    Answer string
}


func (r *GenericResponse) String() string {
    return r.Answer
}

