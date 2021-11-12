package user

import (
	"fmt"
	"sort"
	"sync"
)

type UserID = int
type Tweet struct {
	Text     string `json:"text"`
	UserName string `json:"user_name"`
}
type User struct {
	ID                     UserID
	Name                   string
	Timeline               []Tweet
	FollowedBy             []UserID
	ReceivedFollowRequests []UserID
	Admin                  bool
	Deleted                bool
	// このユーザーの担当者
	StaffForThisUser UserID

	mutex sync.Mutex
}

var users = map[UserID]*User{}
var usersMutex = sync.Mutex{}

func NewUser(id UserID, name string, admin bool, staff UserID) *User {
	return &User{
		ID:                     id,
		Name:                   name,
		Timeline:               []Tweet{},
		FollowedBy:             []UserID{},
		ReceivedFollowRequests: []UserID{},
		mutex:                  sync.Mutex{},
		Admin:                  admin,
		Deleted:                false,
		StaffForThisUser:       staff,
	}
}

func RegisterUser(name string, admin bool, staff UserID) UserID {
	usersMutex.Lock()
	id := UserID(len(users) + 1)
	users[id] = NewUser(id, name, admin, staff)
	usersMutex.Unlock()
	return id
}

func FindUser(id UserID) *User {
	usersMutex.Lock()
	u, ok := users[id]
	usersMutex.Unlock()
	if ok {
		u.mutex.Lock()
		defer u.mutex.Unlock()
		if !u.Deleted {
			return u
		}
	}
	return nil
}

func FindInArray(userID UserID, arr []UserID) (int, bool) {
	for idx, id := range arr {
		if id == userID {
			return idx, true
		}
	}
	return -1, false
}

func (u *User) SendFollowRequest(toID UserID) {
	toUser := FindUser(toID)
	if toUser == nil {
		return
	}
	toUser.mutex.Lock()
	defer toUser.mutex.Unlock()
	if _, ok := FindInArray(u.ID, toUser.ReceivedFollowRequests); !ok {
		toUser.ReceivedFollowRequests = append(toUser.ReceivedFollowRequests, u.ID)
	}
}

func (u *User) DistributeTweet(tweet Tweet) {
	for _, followerID := range u.FollowedBy {
		follower := FindUser(followerID)
		if follower == nil {
			return
		}
		follower.mutex.Lock()
		follower.Timeline = append(follower.Timeline, tweet)
		follower.mutex.Unlock()
	}
}

func (u *User) Tweet(text string) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	tweet := Tweet{
		Text:     text,
		UserName: u.Name,
	}
	u.Timeline = append(u.Timeline, tweet)
	u.DistributeTweet(tweet)
}

func (u *User) AcceptFollowRequest(fromID UserID) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	idx, received := FindInArray(fromID, u.ReceivedFollowRequests)
	if received {
		// delete follow-request
		u.ReceivedFollowRequests = append(u.ReceivedFollowRequests[:idx], u.ReceivedFollowRequests[idx+1:]...)
		u.FollowedBy = append(u.FollowedBy, fromID)
		// prevent deadlocking
		sort.Slice(u.FollowedBy, func(l, r int) bool { return u.FollowedBy[l] < u.FollowedBy[r] })
		return nil
	}
	return fmt.Errorf("there is no corresponding follow-request")
}
