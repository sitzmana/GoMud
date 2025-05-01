package web

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
	"github.com/GoMudEngine/GoMud/internal/users"
)

// adminFunc is the signature for an admin command handler.
type adminFunc func(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error)

// adminCommands maps command names to their handler functions.
var adminCommands = map[string]adminFunc{
	`aid`:         usercommands.Aid,
	`alias`:       usercommands.Alias,
	`appraise`:    usercommands.Appraise,
	`ask`:         usercommands.Ask,
	`attack`:      usercommands.Attack,
	`backstab`:    usercommands.Backstab,
	`badcommands`: usercommands.BadCommands, // Admin only
	`biome`:       usercommands.Biome,
	`broadcast`:   usercommands.Broadcast,
	`bury`:        usercommands.Bury,
	`character`:   usercommands.Character,
	`tackle`:      usercommands.Tackle,
	`bank`:        usercommands.Bank,
	`break`:       usercommands.Break,
	`build`:       usercommands.Build, // Admin only
	`buff`:        usercommands.Buff,  // Admin only
	`bump`:        usercommands.Bump,
	`buy`:         usercommands.Buy,
	`cast`:        usercommands.Cast,
	`cooldowns`:   usercommands.Cooldowns,
	`command`:     usercommands.Command, // Admin only
	`conditions`:  usercommands.Conditions,
	`consider`:    usercommands.Consider,
	`deafen`:      usercommands.Deafen, // Admin only
	`default`:     usercommands.Default,
	`disarm`:      usercommands.Disarm,
	`drop`:        usercommands.Drop,
	`drink`:       usercommands.Drink,
	`eat`:         usercommands.Eat,
	`emote`:       usercommands.Emote,
	`enchant`:     usercommands.Enchant,
	`experience`:  usercommands.Experience,
	`equip`:       usercommands.Equip,
	`flee`:        usercommands.Flee,
	`gearup`:      usercommands.Gearup,
	`get`:         usercommands.Get,
	`give`:        usercommands.Give,
	`go`:          usercommands.Go,
	`grant`:       usercommands.Grant, // Admin only
	`help`:        usercommands.Help,
	`keyring`:     usercommands.KeyRing,
	`killstats`:   usercommands.Killstats,
	`history`:     usercommands.History,
	`inbox`:       usercommands.Inbox,
	`inspect`:     usercommands.Inspect,
	`inventory`:   usercommands.Inventory,
	`item`:        usercommands.Item, // Admin only
	`jobs`:        usercommands.Jobs,
	`list`:        usercommands.List,
	`locate`:      usercommands.Locate, // Admin only
	`lock`:        usercommands.Lock,
	`look`:        usercommands.Look,
	`map`:         usercommands.Map,
	`macros`:      usercommands.Macros,
	`mob`:         usercommands.Mob,    // Admin only
	`modify`:      usercommands.Modify, // Admin only
	`motd`:        usercommands.Motd,
	`mudmail`:     usercommands.Mudmail, // Admin only
	`mute`:        usercommands.Mute,
	`noop`:        usercommands.Noop,
	`offer`:       usercommands.Offer,
	`online`:      usercommands.Online,
	`party`:       usercommands.Party,
	`password`:    usercommands.Password,
	`paz`:         usercommands.Paz, // Admin only
	`peep`:        usercommands.Peep,
	`pet`:         usercommands.Pet,
	`picklock`:    usercommands.Picklock,
	`pickpocket`:  usercommands.Pickpocket,
	`prepare`:     usercommands.Prepare, // Admin only
	`portal`:      usercommands.Portal,
	`pray`:        usercommands.Pray,
	`print`:       usercommands.Print,
	`printline`:   usercommands.PrintLine,
	`put`:         usercommands.Put,
	`pvp`:         usercommands.Pvp,
	`quests`:      usercommands.Quests,
	`quit`:        usercommands.Quit,
	`questtoken`:  usercommands.QuestToken, // Admin only
	`rank`:        usercommands.Rank,
	`read`:        usercommands.Read,
	`recover`:     usercommands.Recover,
	`reload`:      usercommands.Reload, // Admin only
	`remove`:      usercommands.Remove,
	`rename`:      usercommands.Rename,     // Admin only
	`redescribe`:  usercommands.Redescribe, // Admin only
	`room`:        usercommands.Room,       // Admin only
	`save`:        usercommands.Save,
	`say`:         usercommands.Say,
	`scribe`:      usercommands.Scribe,
	`search`:      usercommands.Search,
	`sell`:        usercommands.Sell,
	`server`:      usercommands.Server, // Admin only
	`set`:         usercommands.Set,
	`share`:       usercommands.Share,
	`shoot`:       usercommands.Shoot,
	`shout`:       usercommands.Shout,
	`show`:        usercommands.Show,
	`skills`:      usercommands.Skills,
	`skillset`:    usercommands.Skillset, // Admin only
	`sneak`:       usercommands.Sneak,
	`spawn`:       usercommands.Spawn, // Admin only
	`spell`:       usercommands.Spell, // Admin only
	`spells`:      usercommands.Spells,
	`stash`:       usercommands.Stash,
	`status`:      usercommands.Status,
	`storage`:     usercommands.Storage,
	`suicide`:     usercommands.Suicide,
	`syslogs`:     usercommands.SysLogs, // Admin only
	`tame`:        usercommands.Tame,
	`teleport`:    usercommands.Teleport, // Admin only
	`throw`:       usercommands.Throw,
	`track`:       usercommands.Track,
	`trash`:       usercommands.Trash,
	`train`:       usercommands.Train,
	`unenchant`:   usercommands.Unenchant,
	`uncurse`:     usercommands.Uncurse,
	`unlock`:      usercommands.Unlock,
	`undeafen`:    usercommands.UnDeafen, // Admin only
	`unmute`:      usercommands.UnMute,   // Admin only
	`use`:         usercommands.Use,
	`dual-wield`:  usercommands.DualWield,
	`whisper`:     usercommands.Whisper,
	`who`:         usercommands.Who,
	`zap`:         usercommands.Zap,  // Admin only
	`zone`:        usercommands.Zone, // Admin only
}

// executeAdminCommand runs the handler for the given command string.
func executeAdminCommand(commandLine string, userCtx *users.UserRecord, roomCtx *rooms.Room, rest string) (bool, error) {
	parts := strings.Fields(commandLine)
	if len(parts) == 0 {
		return false, fmt.Errorf("no command provided")
	}
	name := parts[0]

	if len(parts) > 1 {
		rest = strings.Join(parts[1:], " ")
	}

	if handler, ok := adminCommands[name]; ok {
		return handler(rest, userCtx, roomCtx, events.EventFlag(0))
	}
	return false, fmt.Errorf("unknown admin command %q", name)
}
