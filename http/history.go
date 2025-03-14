package http

import (
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models/rest"
)

func getLifetimesSince(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)

	millisString := r.URL.Query().Get("timestamp")
	if millisString == "" {
		SafeErrorAndExit(werror.ErrBadTimestamp, w)
		return
	}

	millis, err := strconv.ParseInt(millisString, 10, 64)
	if err != nil || millis < 0 {
		SafeErrorAndExit(werror.ErrBadTimestamp, w)
		return
	}

	date := time.UnixMilli(millis)

	lifetimes, err := pack.FileService.GetJournalByTree("USERS").GetLifetimesSince(date)
	if SafeErrorAndExit(err, w) {
		return
	}

	writeJson(w, http.StatusOK, lifetimes)
}

func doFullBackup(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	instance := getInstanceFromCtx(r)

	millisString := r.URL.Query().Get("timestamp")
	if millisString == "" {
		SafeErrorAndExit(werror.ErrBadTimestamp, w)
		return
	}

	millis, err := strconv.ParseInt(millisString, 10, 64)
	if SafeErrorAndExit(err, w) {
		return
	} else if millis < 0 {
		SafeErrorAndExit(werror.ErrBadTimestamp, w)
		return
	}

	since := time.UnixMilli(millis)
	usersJournal := pack.FileService.GetJournalByTree("USERS")
	if usersJournal == nil {
		SafeErrorAndExit(werror.ErrNoJournal, w)
		return
	}

	lts, err := usersJournal.GetLifetimesSince(since)
	if SafeErrorAndExit(err, w) {
		return
	}
	usersItter, err := pack.UserService.GetAll()
	if SafeErrorAndExit(err, w) {
		return
	}
	users := slices.Collect(usersItter)

	instances := pack.InstanceService.GetRemotes()
	instances = append(instances, pack.InstanceService.GetLocal())

	usingKey, err := pack.AccessService.GetApiKey(instance.GetUsingKey())
	if SafeErrorAndExit(err, w) {
		return
	}

	owner := pack.UserService.Get(usingKey.Owner)

	keys, err := pack.AccessService.GetKeysByUser(owner)
	if SafeErrorAndExit(err, w) {
		return
	}

	res := rest.NewBackupInfo(
		lts,
		len(usersJournal.GetAllLifetimes()),
		users,
		instances,
		keys,
	)
	writeJson(w, http.StatusOK, res)
}
