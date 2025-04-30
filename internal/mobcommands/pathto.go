package mobcommands

import (
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/mapper"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Pathto(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	// If only going home, check whether a path home was already tried and marked as impossible.
	if rest == `home` {
		cantGoHome := mob.GetTempData(`home-impossible`)
		if cantGoHome != nil && cantGoHome.(bool) == true {
			return true, nil
		}
	}

	toRoomIds := []int{}

	for _, roomIdStr := range util.SplitButRespectQuotes(strings.ToLower(rest)) {

		if roomIdStr == `home` {
			toRoomIds = append(toRoomIds, mob.HomeRoomId)
			continue
		}

		if roomIdInt, err := strconv.Atoi(roomIdStr); err == nil {
			toRoomIds = append(toRoomIds, roomIdInt)
		}
	}

	if len(toRoomIds) < 1 || toRoomIds[0] == mob.Character.RoomId {
		return true, nil
	}

	path, err := mapper.GetPath(mob.Character.RoomId, toRoomIds...)
	if err != nil {
		if rest == `home` {
			mob.SetTempData(`home-impossible`, true)
			mob.Character.SetAdjective(`lost`, true)
		}
		return false, err
	}

	if rest == `home` && len(path) == 0 {
		mob.SetTempData(`home-impossible`, true)
		mob.Character.SetAdjective(`lost`, true)
		return true, nil
	}

	newPath := make([]mobs.PathRoom, len(path))

	// Copy everything over
	for idx, p := range path {
		newPath[idx] = p
	}

	mob.Path.SetPath(newPath)

	return true, nil
}
