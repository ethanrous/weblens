package file

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	tag_model "github.com/ethanrous/weblens/models/tag"
	"github.com/ethanrous/weblens/modules/netwrk"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var hexColorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{3,8}$`)

func isValidHexColor(color string) bool {
	return hexColorRegex.MatchString(color)
}

type createTagParams struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type updateTagParams struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type fileIDsParams struct {
	FileIDs []string `json:"fileIDs"`
}

// GetUserTags godoc
//
//	@ID			GetUserTags
//
//	@Security	SessionAuth
//
//	@Summary	Get all tags for the authenticated user
//	@Tags		Tags
//	@Produce	json
//	@Success	200	{array}	tag_model.Tag	"User's tags"
//	@Failure	401
//	@Failure	500
//	@Router		/tags [get]
func GetUserTags(ctx context_service.RequestContext) {
	tags, err := tag_model.GetTagsByOwner(ctx, ctx.Requester.GetUsername())
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	if tags == nil {
		tags = []*tag_model.Tag{}
	}

	ctx.JSON(http.StatusOK, tags)
}

// CreateTag godoc
//
//	@ID			CreateTag
//
//	@Security	SessionAuth
//
//	@Summary	Create a new tag
//	@Tags		Tags
//	@Accept		json
//	@Produce	json
//	@Param		request	body		createTagParams	true	"Create tag request body"
//	@Success	201		{object}	tag_model.Tag	"Created tag"
//	@Failure	400
//	@Failure	401
//	@Failure	409
//	@Failure	500
//	@Router		/tags [post]
func CreateTag(ctx context_service.RequestContext) {
	params, err := netwrk.ReadRequestBody[createTagParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	}

	if strings.TrimSpace(params.Name) == "" {
		ctx.Error(http.StatusBadRequest, wlerrors.New("tag name is required"))

		return
	}

	if strings.TrimSpace(params.Color) == "" {
		ctx.Error(http.StatusBadRequest, wlerrors.New("tag color is required"))

		return
	}

	if len(strings.TrimSpace(params.Name)) > 255 {
		ctx.Error(http.StatusBadRequest, wlerrors.New("tag name must be 255 characters or fewer"))

		return
	}

	if !isValidHexColor(strings.TrimSpace(params.Color)) {
		ctx.Error(http.StatusBadRequest, wlerrors.New("tag color must be a valid hex color (e.g. #ff0000)"))

		return
	}

	tag, err := tag_model.CreateTag(ctx, strings.TrimSpace(params.Name), strings.TrimSpace(params.Color), ctx.Requester.GetUsername())
	if err != nil {
		if db.IsAlreadyExists(err) {
			ctx.Error(http.StatusConflict, wlerrors.New("a tag with that name already exists"))

			return
		}

		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.JSON(http.StatusCreated, tag)
}

// GetTag godoc
//
//	@ID			GetTag
//
//	@Security	SessionAuth
//
//	@Summary	Get a tag by ID
//	@Tags		Tags
//	@Produce	json
//	@Param		tagID	path		string			true	"Tag ID"
//	@Success	200		{object}	tag_model.Tag	"Tag"
//	@Failure	400
//	@Failure	401
//	@Failure	403
//	@Failure	404
//	@Router		/tags/{tagID} [get]
func GetTag(ctx context_service.RequestContext) {
	tag, err := getTagFromPath(ctx)
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, tag)
}

// UpdateTag godoc
//
//	@ID			UpdateTag
//
//	@Security	SessionAuth
//
//	@Summary	Update a tag's name and/or color
//	@Tags		Tags
//	@Accept		json
//	@Param		tagID	path	string			true	"Tag ID"
//	@Param		request	body	updateTagParams	true	"Update tag request body"
//	@Success	200
//	@Failure	400
//	@Failure	401
//	@Failure	403
//	@Failure	404
//	@Failure	500
//	@Router		/tags/{tagID} [patch]
func UpdateTag(ctx context_service.RequestContext) {
	tag, err := getTagFromPath(ctx)
	if err != nil {
		return
	}

	params, err := netwrk.ReadRequestBody[updateTagParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	}

	if params.Name == "" && params.Color == "" {
		ctx.Error(http.StatusBadRequest, wlerrors.New("at least one of name or color is required"))

		return
	}

	if params.Name != "" && len(strings.TrimSpace(params.Name)) > 255 {
		ctx.Error(http.StatusBadRequest, wlerrors.New("tag name must be 255 characters or fewer"))

		return
	}

	if params.Color != "" && !isValidHexColor(strings.TrimSpace(params.Color)) {
		ctx.Error(http.StatusBadRequest, wlerrors.New("tag color must be a valid hex color (e.g. #ff0000)"))

		return
	}

	err = tag_model.UpdateTag(ctx, tag.TagID, strings.TrimSpace(params.Name), strings.TrimSpace(params.Color))
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.JSON(http.StatusOK, nil)
}

// DeleteTag godoc
//
//	@ID			DeleteTag
//
//	@Security	SessionAuth
//
//	@Summary	Delete a tag
//	@Tags		Tags
//	@Param		tagID	path	string	true	"Tag ID"
//	@Success	200
//	@Failure	400
//	@Failure	401
//	@Failure	403
//	@Failure	404
//	@Failure	500
//	@Router		/tags/{tagID} [delete]
func DeleteTag(ctx context_service.RequestContext) {
	tag, err := getTagFromPath(ctx)
	if err != nil {
		return
	}

	err = tag_model.DeleteTag(ctx, tag.TagID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.JSON(http.StatusOK, nil)
}

// AddFilesToTag godoc
//
//	@ID			AddFilesToTag
//
//	@Security	SessionAuth
//
//	@Summary	Add files to a tag
//	@Tags		Tags
//	@Accept		json
//	@Param		tagID	path	string			true	"Tag ID"
//	@Param		request	body	fileIDsParams	true	"File IDs to add"
//	@Success	200
//	@Failure	400
//	@Failure	401
//	@Failure	403
//	@Failure	404
//	@Failure	500
//	@Router		/tags/{tagID}/files [post]
func AddFilesToTag(ctx context_service.RequestContext) {
	tag, err := getTagFromPath(ctx)
	if err != nil {
		return
	}

	params, err := netwrk.ReadRequestBody[fileIDsParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	}

	if len(params.FileIDs) == 0 {
		ctx.Error(http.StatusBadRequest, wlerrors.New("fileIDs is required"))

		return
	}

	for _, id := range params.FileIDs {
		if strings.TrimSpace(id) == "" {
			ctx.Error(http.StatusBadRequest, wlerrors.New("fileIDs must not contain empty values"))

			return
		}
	}

	err = tag_model.AddFilesToTag(ctx, tag.TagID, params.FileIDs)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.JSON(http.StatusOK, nil)
}

// RemoveFilesFromTag godoc
//
//	@ID			RemoveFilesFromTag
//
//	@Security	SessionAuth
//
//	@Summary	Remove files from a tag
//	@Tags		Tags
//	@Accept		json
//	@Param		tagID	path	string			true	"Tag ID"
//	@Param		request	body	fileIDsParams	true	"File IDs to remove"
//	@Success	200
//	@Failure	400
//	@Failure	401
//	@Failure	403
//	@Failure	404
//	@Failure	500
//	@Router		/tags/{tagID}/files [delete]
func RemoveFilesFromTag(ctx context_service.RequestContext) {
	tag, err := getTagFromPath(ctx)
	if err != nil {
		return
	}

	params, err := netwrk.ReadRequestBody[fileIDsParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	}

	if len(params.FileIDs) == 0 {
		ctx.Error(http.StatusBadRequest, wlerrors.New("fileIDs is required"))

		return
	}

	for _, id := range params.FileIDs {
		if strings.TrimSpace(id) == "" {
			ctx.Error(http.StatusBadRequest, wlerrors.New("fileIDs must not contain empty values"))

			return
		}
	}

	err = tag_model.RemoveFilesFromTag(ctx, tag.TagID, params.FileIDs)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.JSON(http.StatusOK, nil)
}

// GetTagsForFile godoc
//
//	@ID			GetTagsForFile
//
//	@Security	SessionAuth
//
//	@Summary	Get tags for a file
//	@Tags		Tags
//	@Produce	json
//	@Param		fileID	path	string			true	"File ID"
//	@Success	200		{array}	tag_model.Tag	"Tags containing the file"
//	@Failure	400
//	@Failure	401
//	@Failure	500
//	@Router		/tags/file/{fileID} [get]
func GetTagsForFile(ctx context_service.RequestContext) {
	fileID := ctx.Path("fileID")
	if fileID == "" {
		ctx.Error(http.StatusBadRequest, wlerrors.New("fileID is required"))

		return
	}

	tags, err := tag_model.GetTagsForFile(ctx, fileID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	// Filter to only the requesting user's tags
	owner := ctx.Requester.GetUsername()

	owned := make([]*tag_model.Tag, 0, len(tags))

	for _, t := range tags {
		if t.Owner == owner {
			owned = append(owned, t)
		}
	}

	ctx.JSON(http.StatusOK, owned)
}

// GetFilesByTag godoc
//
//	@ID			GetFilesByTag
//
//	@Security	SessionAuth
//
//	@Summary	Get all files in a tag
//	@Tags		Tags
//	@Produce	json
//	@Param		tagID	path		string				true	"Tag ID"
//	@Success	200		{object}	wlstructs.FilesInfo	"Files in the tag"
//	@Failure	400
//	@Failure	401
//	@Failure	403
//	@Failure	404
//	@Failure	500
//	@Router		/tags/{tagID}/files [get]
func GetFilesByTag(ctx context_service.RequestContext) {
	tag, err := getTagFromPath(ctx)
	if err != nil {
		return
	}

	files := make([]*file_model.WeblensFileImpl, 0, len(tag.FileIDs))

	for _, fileID := range tag.FileIDs {
		file, err := ctx.FileService.GetFileByID(ctx, fileID)
		if err != nil {
			if wlerrors.Is(err, file_model.ErrFileNotFound) {
				// Stale file ID in tag — skip silently
				continue
			}

			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		// Verify the requester can access this file
		if _, err = auth.CanUserAccessFile(ctx, ctx.Requester, file, ctx.Share); err != nil {
			continue
		}

		files = append(files, file)
	}

	resp, err := filesInfoFromFiles(ctx, files)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// getTagFromPath extracts the tag ID from the URL path, fetches the tag,
// and verifies ownership. Returns nil tag and reports error to client on failure.
func getTagFromPath(ctx context_service.RequestContext) (*tag_model.Tag, error) {
	tagIDStr := ctx.Path("tagID")
	if tagIDStr == "" {
		ctx.Error(http.StatusBadRequest, wlerrors.New("tagID is required"))

		return nil, wlerrors.New("tagID is required")
	}

	tagID, err := primitive.ObjectIDFromHex(tagIDStr)
	if err != nil {
		ctx.Error(http.StatusBadRequest, wlerrors.New("invalid tagID"))

		return nil, err
	}

	tag, err := tag_model.GetTagByID(ctx, tagID)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return nil, err
	}

	if tag.Owner != ctx.Requester.GetUsername() {
		ctx.Error(http.StatusForbidden, wlerrors.New("not the owner of this tag"))

		return nil, wlerrors.New("forbidden")
	}

	return tag, nil
}
