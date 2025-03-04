package werror

import "errors"

var ErrNoMedia = ClientSafeErr{
	safeErr:    errors.New("no media found"),
	statusCode: 404,
}

var ErrNoExiftool = errors.New("exiftool not initialized")
var ErrNoCache = errors.New("could not find or generate requested media cache file")

var ErrMediaNil = errors.New("media is nil")
var ErrMediaBadMime = errors.New("media has missing or unrecognized mime type")
var ErrMediaNotVideo = errors.New("media is not a video type")
var ErrMediaNoId = errors.New("media has no contentId")
var ErrMediaNoDimensions = errors.New("media has a missing width or height dimension")
var ErrMediaNoPages = errors.New("media must have a page count of at least 1")
var ErrMediaNoFiles = errors.New("media cannot be added with no file ids specified")
var ErrMediaAlreadyExists = errors.New("media with given contentId already exists")
var ErrMediaNoDuration = errors.New("media of video type must have a duration")
var ErrMediaHasDuration = errors.New("media of non-video type cannot have a duration")
