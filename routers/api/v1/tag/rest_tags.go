// Package tag provides REST API handlers for file tag management.
package tag

import (
	"net/http"
	"strings"

	"github.com/ethanrous/weblens/models/db"
	tag_model "github.com/ethanrous/weblens/models/tag"
	"github.com/ethanrous/weblens/modules/netwrk"
	"github.com/ethanrous/weblens/modules/wlerrors"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

// GetUserTags returns all tags owned by the authenticated user.
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

// CreateTag creates a new tag for the authenticated user.
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

// GetTag returns a single tag by ID.
func GetTag(ctx context_service.RequestContext) {
	tag, err := getTagFromPath(ctx)
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, tag)
}

// UpdateTag updates a tag's name and/or color.
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

	err = tag_model.UpdateTag(ctx, tag.TagID, strings.TrimSpace(params.Name), params.Color)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.JSON(http.StatusOK, nil)
}

// DeleteTag deletes a tag and removes it from all files.
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

// AddFilesToTag adds file IDs to a tag.
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

// RemoveFilesFromTag removes file IDs from a tag.
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

// GetTagsForFile returns the authenticated user's tags that contain the given file.
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
