package follow

import (
	"embed"
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/parties"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

var (

	//////////////////////////////////////////////////////////////////////
	// NOTE: The below //go:embed directive is important!
	// It embeds the relative path into the var below it.
	//////////////////////////////////////////////////////////////////////

	//go:embed files/*
	files embed.FS
)

// ////////////////////////////////////////////////////////////////////
// NOTE: The init function in Go is a special function that is
// automatically executed before the main function within a package.
// It is used to initialize variables, set up configurations, or
// perform any other setup tasks that need to be done before the
// program starts running.
// ////////////////////////////////////////////////////////////////////
func init() {
	//
	// We can use all functions only, but this demonstrates
	// how to use a struct
	//
	f := FollowModule{
		plug:      plugins.New(`follow`, `1.0`),
		followed:  make(map[followId][]followId),
		followers: make(map[followId]followId),
	}

	//
	// Add the embedded filesystem
	//
	if err := f.plug.AttachFileSystem(files); err != nil {
		panic(err)
	}
	//
	// Register any user/mob commands
	//
	f.plug.AddUserCommand(`follow`, f.followUserCommand, true, false)
	f.plug.AddMobCommand(`follow`, f.followMobCommand, true)

	events.RegisterListener(events.RoomChange{}, f.roomChangeHandler)
	events.RegisterListener(events.PlayerDespawn{}, f.playerDespawnHandler)
	events.RegisterListener(events.MobIdle{}, f.idleMobHandler, events.First)
	events.RegisterListener(events.PartyUpdated{}, f.onPartyChange)
}

//////////////////////////////////////////////////////////////////////
// NOTE: What follows is all custom code. For this module.
//////////////////////////////////////////////////////////////////////

type followId struct {
	userId        int
	mobInstanceId int
}

// Using a struct gives a way to store longer term data.
type FollowModule struct {
	// Keep a reference to the plugin when we create it so that we can call ReadBytes() and WriteBytes() on it.
	plug *plugins.Plugin

	followed  map[followId][]followId // key => who's followed. value ([]followId{}) => who's following them
	followers map[followId]followId   // key => who's following someone. value => who's being followed
}

// Get all followeres attached to a target
func (f *FollowModule) isFollowing(followCheck followId) bool {
	_, ok := f.followers[followCheck]
	return ok
}

// Get all followeres attached to a target
func (f *FollowModule) getFollowers(followTarget followId) []followId {

	if _, ok := f.followed[followTarget]; !ok {
		return []followId{}
	}

	followerResults := make([]followId, len(f.followed[followTarget]))
	copy(followerResults, f.followed[followTarget])

	return followerResults
}

// Add a single follower to a target
func (f *FollowModule) startFollow(followTarget followId, followSource followId) {

	// Make sure they no longer follow whoever they were before.
	f.stopFollowing(followSource)

	f.followers[followSource] = followTarget
	if _, ok := f.followed[followTarget]; !ok {
		f.followed[followTarget] = []followId{}
	}

	f.followed[followTarget] = append(f.followed[followTarget], followSource)
}

// Remove a single follower from whoever they are following (if any)
func (f *FollowModule) stopFollowing(followSource followId) followId {

	wasFollowing := followId{}

	if followTarget, ok := f.followers[followSource]; ok {
		delete(f.followers, followSource)

		wasFollowing = followTarget

		for idx, fId := range f.followed[followTarget] {
			if fId == followSource {
				f.followed[followTarget] = append(f.followed[followTarget][0:idx], f.followed[followTarget][idx+1:]...)

				if len(f.followed[followTarget]) == 0 {
					delete(f.followed, followTarget)
				}

				break
			}
		}
	}

	return wasFollowing
}

// Remove all followers from a target
func (f *FollowModule) loseFollowers(followTarget followId) []followId {
	allFollowers := f.getFollowers(followTarget)
	for _, followSource := range allFollowers {
		f.stopFollowing(followSource)
	}
	return allFollowers
}

//
// Event Handlers
//

// If players make changes (into/out of party)
// Just make sure they aren't following anyone.
// This is just basic cleanup/precaution
func (f *FollowModule) onPartyChange(e events.Event) events.ListenerReturn {

	evt := e.(events.PartyUpdated)

	for _, uId := range evt.UserIds {
		f.stopFollowing(followId{userId: uId})
	}

	return events.Continue
}

// Interrupt the idle action of mobs if they are currently following someone.
func (f *FollowModule) idleMobHandler(e events.Event) events.ListenerReturn {
	evt := e.(events.MobIdle)

	if f.isFollowing(followId{mobInstanceId: evt.MobInstanceId}) {
		return events.Cancel
	}

	return events.Continue
}

func (f *FollowModule) roomChangeHandler(e events.Event) events.ListenerReturn {
	evt := e.(events.RoomChange)

	moverId := followId{userId: evt.UserId, mobInstanceId: evt.MobInstanceId}

	allFollowers := f.getFollowers(moverId)
	if len(allFollowers) == 0 {
		return events.Continue
	}

	fromRoom := rooms.LoadRoom(evt.FromRoomId)
	if fromRoom == nil {
		return events.Continue
	}

	followExitName := ``
	for exitName, exitInfo := range fromRoom.Exits {
		if exitInfo.RoomId == evt.ToRoomId {
			followExitName = exitName
			break
		}
	}

	if followExitName == `` {
		for exitName, exitInfo := range fromRoom.ExitsTemp {
			if exitInfo.RoomId == evt.ToRoomId {
				followExitName = exitName
				break
			}
		}
	}

	// The exit they went through is gone/missing? (Teleported?)
	// End the follow
	if followExitName == `` {
		if evt.UserId > 0 {
			if user := users.GetByUserId(evt.UserId); user != nil {
				user.Command(`follow lose`)
			}
		}
	} else {

		for _, fId := range allFollowers {

			if fId.mobInstanceId > 0 {

				if mob := mobs.GetInstance(fId.mobInstanceId); mob != nil {
					if fromRoom.RoomId == mob.Character.RoomId {
						mob.Command(followExitName, .25)
						continue
					}

					mob.Command(`follow stop`)
				}
				f.stopFollowing(fId)

			} else if fId.userId > 0 {

				if user := users.GetByUserId(fId.userId); user != nil {
					if fromRoom.RoomId == user.Character.RoomId {
						user.Command(followExitName, .25)
						continue
					}

					user.Command(`follow stop`)
				}
				f.stopFollowing(fId)

			}

		}

	}

	return events.Continue
}

func (f *FollowModule) playerDespawnHandler(e events.Event) events.ListenerReturn {
	// Don't really care about the event data for this
	evt, typeOk := e.(events.PlayerDespawn)
	if !typeOk {
		return events.Cancel
	}

	f.loseFollowers(followId{userId: evt.UserId})

	return events.Continue
}

//
// Commands
//

func (f *FollowModule) followUserCommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if rest == "" {
		user.SendText(`Follow whom? Try <ansi fg="command">help command</ansi>`)
		return true, nil
	}

	if parties.Get(user.UserId) != nil {
		user.SendText(`You can't use this command while in a party.`)
		return true, nil
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	followTargetName := args[0]
	followAction := `follow`

	if rest == `stop` || rest == `lose` {
		followAction = rest
		followTargetName = ``
	}

	userId, mobInstId := 0, 0
	if len(followTargetName) > 0 {
		userId, mobInstId = room.FindByName(followTargetName)
	}

	followCommandTarget := followId{userId: userId, mobInstanceId: mobInstId}
	followCommandSource := followId{userId: user.UserId}

	if followCommandTarget.userId == followCommandSource.userId {
		user.SendText(`You can't target yourself.`)
		return true, nil
	}

	// Lose any followers
	if followAction == `lose` {

		if lostFollowers := f.loseFollowers(followCommandSource); len(lostFollowers) > 0 {

			// Tell all the followers they
			for _, fId := range lostFollowers {
				if fId.userId == 0 {
					continue
				}

				if followerUser := users.GetByUserId(fId.userId); followerUser != nil {
					followerUser.SendText(fmt.Sprintf(`You are no longer following <ansi fg="username">%s</ansi>.`, user.Character.Name))
				}
			}

		}

		user.SendText(fmt.Sprintf(`Nobody is following you.`))

		return true, nil
	}

	// Stop following someone?
	if followAction == `stop` {

		wasFollowing := f.stopFollowing(followCommandSource)

		if wasFollowing.userId > 0 {

			if followUser := users.GetByUserId(wasFollowing.userId); followUser != nil {
				followUser.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> stopped following you.`, followUser.Character.Name))
				user.SendText(fmt.Sprintf(`You are no longer following <ansi fg="username">%s</ansi>.`, followUser.Character.Name))
				return true, nil
			}

		}

		if wasFollowing.mobInstanceId > 0 {

			if followMob := mobs.GetInstance(wasFollowing.mobInstanceId); followMob != nil {
				user.SendText(fmt.Sprintf(`You are no longer following <ansi fg="mobname">%s</ansi>.`, followMob.Character.Name))
				return true, nil
			}

		}

		user.SendText(`You aren't following anyone.`)

		return true, nil
	}

	// Default behavior is follow
	if followCommandTarget.userId > 0 {

		f.startFollow(followCommandTarget, followCommandSource)

		targetUser := users.GetByUserId(followCommandTarget.userId)

		user.SendText(fmt.Sprintf(`You start following <ansi fg="username">%s</ansi>.`, targetUser.Character.Name))

		targetUser.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is following you.`, user.Character.Name))

		return true, nil
	}

	if followCommandTarget.mobInstanceId > 0 {

		targetMob := mobs.GetInstance(followCommandTarget.mobInstanceId)

		if targetMob.HatesAlignment(user.Character.Alignment) {
			user.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> won't let you follow them.`, targetMob.Character.Name))
		} else {
			f.startFollow(followCommandTarget, followCommandSource)

			user.SendText(fmt.Sprintf(`You start following <ansi fg="mobname">%s</ansi>.`, targetMob.Character.Name))
		}

		return true, nil
	}

	user.SendText(`Follow whom?`)

	return true, nil
}

func (f *FollowModule) followMobCommand(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	if rest == "" {
		return true, nil
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	followTargetName := args[0]
	followAction := `follow`

	if rest == `stop` || rest == `lose` {
		followAction = rest
		followTargetName = ``
	}

	userId, mobInstId := 0, 0
	if len(followTargetName) > 0 {
		userId, mobInstId = room.FindByName(followTargetName)
	}

	followCommandTarget := followId{userId: userId, mobInstanceId: mobInstId}
	followCommandSource := followId{mobInstanceId: mob.InstanceId}

	if followCommandTarget.mobInstanceId == followCommandSource.mobInstanceId {
		return true, nil
	}

	// Lose any followers
	if followAction == `lose` {

		if lostFollowers := f.loseFollowers(followCommandSource); len(lostFollowers) > 0 {

			// Tell all the followers they
			for _, fId := range lostFollowers {
				if fId.userId == 0 {
					continue
				}

				if followerUser := users.GetByUserId(fId.userId); followerUser != nil {
					followerUser.SendText(fmt.Sprintf(`You are no longer following <ansi fg="mobname">%s</ansi>.`, mob.Character.Name))
				}
			}

			return true, nil
		}

		return true, nil
	}

	// Stop following someone?
	if followAction == `stop` {

		wasFollowing := f.stopFollowing(followCommandSource)

		if wasFollowing.userId > 0 {

			if followUser := users.GetByUserId(wasFollowing.userId); followUser != nil {
				followUser.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> stopped following you.`, followUser.Character.Name))
				return true, nil
			}

		}

		return true, nil
	}

	// Default behavior is follow

	// If they are on a path, clear it. The follow takes priority.
	mob.Path.Clear()

	if followCommandTarget.userId > 0 {

		f.startFollow(followCommandTarget, followCommandSource)

		targetUser := users.GetByUserId(followCommandTarget.userId)

		targetUser.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is following you.`, mob.Character.Name))

		return true, nil
	}

	if followCommandTarget.mobInstanceId > 0 {

		f.startFollow(followCommandTarget, followCommandSource)

		return true, nil
	}

	return false, nil
}
