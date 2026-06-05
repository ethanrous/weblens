package file

import (
	"maps"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/ethanrous/weblens/models/embedding"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/media"
	share_model "github.com/ethanrous/weblens/models/share"
	tag_model "github.com/ethanrous/weblens/models/tag"
	"github.com/ethanrous/weblens/modules/option"
	"github.com/ethanrous/weblens/modules/set"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlfs"
	"github.com/ethanrous/weblens/modules/wlslices"
	"github.com/ethanrous/weblens/modules/wlstructs"
	"github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	embed_service "github.com/ethanrous/weblens/services/embed"
	"github.com/ethanrous/weblens/services/reshape"
	"golang.org/x/sync/errgroup"
)

const (
	// minContentScore is the raw-cosine floor below which a semantic hit is dropped.
	minContentScore = 0.30

	// imageDeferMargin is how far a file_chunk hit must out-score an image hit to suppress it.
	imageDeferMargin = 0.10

	// minRelevanceSpread vetoes a kind's set when its top hit barely exceeds the mean.
	minRelevanceSpread = 0.03
)

// fuzzyMatch pairs a file ID with its fuzzy-rank distance (lower = better match).
type fuzzyMatch struct {
	FileID string
	Rank   int
}

// contentHit holds the best semantic score, snippet, and source page for a file.
type contentHit struct {
	Score   float64
	Snippet string
	Page    int
}

// collectCandidates returns parallel fileID/filename slices for baseFolder, filtered by tagFilterFileIDs.
func collectCandidates(baseFolder *file_model.WeblensFileImpl, recursive bool, tagFilterFileIDs set.Set[string]) (fileIDs []string, filenames []string) {
	if recursive {
		_ = baseFolder.RecursiveMap(
			func(f *file_model.WeblensFileImpl) error {
				if tagFilterFileIDs.Len() != 0 && !tagFilterFileIDs.Has(f.ID()) {
					return nil
				}

				fileIDs = append(fileIDs, f.ID())
				filenames = append(filenames, f.GetPortablePath().Filename())

				return nil
			},
		)
	} else {
		for _, child := range baseFolder.GetChildren() {
			if tagFilterFileIDs.Len() != 0 && !tagFilterFileIDs.Has(child.ID()) {
				continue
			}

			fileIDs = append(fileIDs, child.ID())
			filenames = append(filenames, child.GetPortablePath().Filename())
		}
	}

	return fileIDs, filenames
}

// runFilenameMatch matches search against filenames (regex match, or case-insensitive substring ranked by offset).
func runFilenameMatch(search string, useRegex bool, fileIDs []string, filenames []string) ([]fuzzyMatch, error) {
	if useRegex {
		re, err := regexp.Compile(search)
		if err != nil {
			return nil, wlerrors.Statusf(http.StatusBadRequest, "invalid regex pattern: %s", search)
		}

		var out []fuzzyMatch

		for i, filename := range filenames {
			if re.MatchString(filename) {
				out = append(out, fuzzyMatch{FileID: fileIDs[i], Rank: 0})
			}
		}

		return out, nil
	}

	needle := strings.ToLower(search)

	out := make([]fuzzyMatch, 0)

	for i, filename := range filenames {
		idx := strings.Index(strings.ToLower(filename), needle)
		if idx < 0 {
			continue
		}

		out = append(out, fuzzyMatch{FileID: fileIDs[i], Rank: idx})
	}

	wlslices.SortFunc(out, func(a, b fuzzyMatch) int {
		return a.Rank - b.Rank
	})

	return out, nil
}

// mergeContentHits demuxes filtered []embedding.Hit onto file IDs (file_chunk by fileID, image via byMedia).
func mergeContentHits(hits []embedding.Hit, byMedia map[string][]*file_model.WeblensFileImpl) map[string]contentHit {
	filtered := filterContentHits(hits)

	out := map[string]contentHit{}

	for _, h := range filtered {
		var (
			fileIDs []string
			snippet string
		)

		switch h.Kind {
		case embedding.KindFileChunk:
			fileIDs = []string{h.SourceID}
			snippet = h.Snippet
		case embedding.KindImage:
			for _, f := range byMedia[h.SourceID] {
				fileIDs = append(fileIDs, f.ID())
			}
		}

		for _, fid := range fileIDs {
			cur, ok := out[fid]

			switch {
			case !ok || h.Score > cur.Score:
				merged := contentHit{Score: h.Score, Snippet: snippet, Page: h.Page}
				// Image hits carry no snippet/page; keep the prior text hit's so the result stays linkable.
				if merged.Snippet == "" && ok {
					merged.Snippet = cur.Snippet
					merged.Page = cur.Page
				}

				out[fid] = merged
			case cur.Snippet == "" && snippet != "":
				cur.Snippet = snippet
				cur.Page = h.Page
				out[fid] = cur
			}
		}
	}

	return out
}

// filterContentHits gates raw vector hits via per-kind spread veto, absolute floor, and cross-modal defer.
func filterContentHits(hits []embedding.Hit) []embedding.Hit {
	byKind := map[embedding.Kind][]embedding.Hit{}

	for _, h := range hits {
		switch h.Kind {
		case embedding.KindFileChunk, embedding.KindImage:
			byKind[h.Kind] = append(byKind[h.Kind], h)
		}
	}

	surviving := map[embedding.Kind][]embedding.Hit{}

	for kind, kindHits := range byKind {
		if !hasMeaningfulSpread(kindHits) {
			continue
		}

		for _, h := range kindHits {
			if h.Score >= minContentScore {
				surviving[kind] = append(surviving[kind], h)
			}
		}
	}

	textHits := surviving[embedding.KindFileChunk]

	bestText := 0.0
	for _, h := range textHits {
		if h.Score > bestText {
			bestText = h.Score
		}
	}

	out := make([]embedding.Hit, 0, len(hits))

	for kind, kindHits := range surviving {
		for _, h := range kindHits {
			// A confident document match suppresses marginal images, when a file_chunk hit survived.
			if kind == embedding.KindImage && len(textHits) > 0 && bestText-h.Score > imageDeferMargin {
				continue
			}

			out = append(out, h)
		}
	}

	return out
}

// hasMeaningfulSpread reports whether the top hit exceeds the set's mean by at least minRelevanceSpread.
func hasMeaningfulSpread(hits []embedding.Hit) bool {
	if len(hits) < 2 {
		return true
	}

	var (
		sum    float64
		maxVal = hits[0].Score
	)

	for _, h := range hits {
		sum += h.Score

		if h.Score > maxVal {
			maxVal = h.Score
		}
	}

	mean := sum / float64(len(hits))

	return maxVal-mean >= minRelevanceSpread
}

// fileMatch is the running per-file aggregation as filename and content matches are merged.
type fileMatch struct {
	Score   float64
	Kinds   []string
	Snippet string
	Page    int
}

// mergeSearchResults blends fuzzy filename ranks with content hits into SearchResult slices.
func mergeSearchResults(
	ctx context_service.RequestContext,
	fnRanks []fuzzyMatch,
	contentHits map[string]contentHit,
	candidates map[string]*file_model.WeblensFileImpl,
	mediaMap map[string]*media.Media,
	tagBrowse bool,
) ([]wlstructs.SearchResult, error) {
	matches := make(map[string]*fileMatch, len(fnRanks)+len(contentHits))

	// Tags-only browse has no relevance query, so every tag-filtered candidate is a neutral result.
	if tagBrowse {
		for fid := range candidates {
			if fid == ctx.Requester.HomeID || fid == ctx.Requester.TrashID {
				continue
			}

			matches[fid] = &fileMatch{}
		}
	}

	maxRank := 1
	for _, m := range fnRanks {
		if m.Rank > maxRank {
			maxRank = m.Rank
		}
	}

	for _, m := range fnRanks {
		matches[m.FileID] = &fileMatch{
			Score: 1 - float64(m.Rank)/float64(maxRank),
			Kinds: []string{wlstructs.MatchKindFilename},
		}
	}

	for fid, h := range contentHits {
		fm, ok := matches[fid]
		if !ok {
			fm = &fileMatch{}
			matches[fid] = fm
		}

		fm.Kinds = append(fm.Kinds, wlstructs.MatchKindContent)
		fm.Snippet = h.Snippet
		fm.Page = h.Page

		if h.Score > fm.Score {
			fm.Score = h.Score
		}
	}

	out := make([]wlstructs.SearchResult, 0, len(matches))

	// Cache parent-folder permissions so each result carries the requester's actual access.
	parentPerms := make(map[string]*share_model.Permissions)

	for fid, fm := range matches {
		f, ok := candidates[fid]
		if !ok {
			continue
		}

		_, hasMedia := mediaMap[f.GetContentID()]
		opts := reshape.FileInfoOptions{HasMedia: hasMedia}

		if parent := f.GetParent(); parent != nil {
			perms, cached := parentPerms[parent.ID()]
			if !cached {
				p, err := auth.CanUserAccessFile(ctx, ctx.Requester, parent, ctx.Share)
				if err != nil {
					ctx.Log().Error().Err(err).Msgf("failed to check permissions for %s", fid)

					continue
				}

				parentPerms[parent.ID()] = p
				perms = p
			}

			opts.Perms = option.Of(*perms)
		}

		info, err := reshape.WeblensFileToFileInfo(&ctx.AppContext, f, opts)
		if err != nil {
			ctx.Log().Warn().Err(err).Msgf("reshape failed for %s", fid)

			continue
		}

		out = append(out, wlstructs.SearchResult{
			File:         info,
			MatchKind:    fm.Kinds,
			MatchSnippet: fm.Snippet,
			MatchPage:    fm.Page,
			Score:        fm.Score,
		})
	}

	return out, nil
}

// SearchFiles godoc
//
//	@ID			SearchFiles
//
//	@Security	SessionAuth
//
//	@Summary	Search for files by filename or content
//	@Tags		Files
//
//	@Param		search			query		string	true	"Filename to search for"
//	@Param		baseFolderID	query		string	false	"The folder to search in, defaults to the user's home folder"
//	@Param		sortProp		query		string	false	"Property to sort by"									Enums(name, size, updatedAt)	default(name)
//	@Param		sortOrder		query		string	false	"Sort order"											Enums(asc, desc)				default(asc)
//	@Param		recursive		query		boolean	false	"Search recursively"									Enums(true, false)				default(false)
//	@Param		regex			query		boolean	false	"Whether to treat the search term as a regex pattern"	Enums(true, false)				default(false)
//	@Param		tags			query		string	false	"Comma-separated list of tags to filter by"
//	@Param		tagJoinLogic	query		string	false	"Logic to combine multiple tags with, either 'and' or 'or'"	Enums(and, or)	default(or)
//	@Param		includeContent	query		bool	false	"Include semantic content matches"						default(true)
//	@Success	200				{array}		SearchResult
//	@Failure	400
//	@Failure	401
//	@Failure	500
//	@Router		/files/search [get]
func SearchFiles(ctx context_service.RequestContext) {
	filenameSearch := ctx.Query("search")
	tagIDsFilter := ctx.QueryArray("tags")

	if filenameSearch == "" && len(tagIDsFilter) == 0 {
		ctx.Error(http.StatusBadRequest, wlerrors.New("at least one of 'search' or 'tags' query parameters is required"))

		return
	}

	ctx.Log().Trace().Msgf("Searching for filename: %s", filenameSearch)

	baseFolderID := ctx.Query("baseFolderID")
	if baseFolderID == "" {
		baseFolderID = ctx.Requester.HomeID
	}

	baseFolder, err := auth.RequireFileAccessOne(ctx, baseFolderID)
	if err != nil {
		return
	}

	if !baseFolder.IsDir() {
		ctx.Error(http.StatusBadRequest, wlerrors.New("the baseFolderID must be a directory"))

		return
	}

	// Compute tags filter
	tagFilterFileIDs := set.New[string]()

	if len(tagIDsFilter) > 0 {
		tagJoinLogic := ctx.Query("tagJoinLogic")
		if tagJoinLogic == "" {
			tagJoinLogic = "or"
		} else if tagJoinLogic != "or" && tagJoinLogic != "and" {
			ctx.Error(http.StatusBadRequest, wlerrors.New("invalid tagJoinLogic, must be 'and' or 'or'"))

			return
		}

		tags, err := tag_model.GetTagsByOwner(ctx, ctx.Requester.GetUsername(), tagIDsFilter...)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return
		} else if len(tags) != len(tagIDsFilter) {
			ctx.Error(http.StatusBadRequest, wlerrors.Errorf("one or more tags not found of the provided tags: %v", tagIDsFilter))

			return
		}

		andInitialized := false

		for _, t := range tags {
			currentTagSet := set.New(t.FileIDs...)
			if tagJoinLogic == "or" {
				tagFilterFileIDs = tagFilterFileIDs.Union(currentTagSet)
			} else if !andInitialized {
				tagFilterFileIDs = currentTagSet
				andInitialized = true
			} else {
				tagFilterFileIDs = tagFilterFileIDs.Intersection(currentTagSet)
			}
		}
	}

	useRegex := ctx.QueryBool("regex")
	includeContent := true

	if v := ctx.Query("includeContent"); v != "" {
		includeContent = ctx.QueryBool("includeContent")
	}

	fileIDs, filenames := collectCandidates(baseFolder, ctx.QueryBool("recursive"), tagFilterFileIDs)

	// Build a candidate map (fileID → file) and media-content-ID index for image embedding demux.
	candidates := make(map[string]*file_model.WeblensFileImpl, len(fileIDs))
	filesByMediaContentID := make(map[string][]*file_model.WeblensFileImpl)

	for _, fid := range fileIDs {
		f, err := ctx.FileService.GetFileByID(ctx, fid)
		if err != nil {
			ctx.Log().Error().Stack().Err(err).Msgf("Failed to get file by ID: %s", fid)

			continue
		}

		candidates[fid] = f

		if cid := f.GetContentID(); cid != "" {
			filesByMediaContentID[cid] = append(filesByMediaContentID[cid], f)
		}
	}

	// Build the source-ID set for the embedding search: candidate fileIDs plus media content IDs.
	sourceIDSet := make([]string, 0, len(candidates)+len(filesByMediaContentID))
	for fid := range candidates {
		sourceIDSet = append(sourceIDSet, fid)
	}

	for cid := range filesByMediaContentID {
		sourceIDSet = append(sourceIDSet, cid)
	}

	var (
		fnMatches []fuzzyMatch
		embedHits map[string]contentHit
	)

	g, gctx := errgroup.WithContext(ctx)

	// Goroutine A: filename fuzzy / regex search.
	g.Go(func() error {
		_ = gctx

		if filenameSearch == "" {
			return nil
		}

		var err error

		fnMatches, err = runFilenameMatch(filenameSearch, useRegex, fileIDs, filenames)

		return err
	})

	// Goroutine B: semantic content search (skipped for regex, disabled content, or unavailable service).
	g.Go(func() error {
		if filenameSearch == "" || useRegex || !includeContent || embed_service.Default().ServiceUnavailable() {
			return nil
		}

		plainVec, imageVec, err := embed_service.Default().EncodeQueryText(gctx, filenameSearch)
		if err != nil {
			ctx.Log().Warn().Err(err).Msg("embed encode text failed, skipping content search")

			return nil
		}

		textHits, err := embedding.Search(gctx, embedding.Query{
			Vector:    plainVec,
			SourceIDs: sourceIDSet,
			Kind:      embedding.KindFileChunk,
			Limit:     100,
		})
		if err != nil {
			ctx.Log().Warn().Err(err).Msg("embed text search failed, skipping content search")

			return nil
		}

		imageHits, err := embedding.Search(gctx, embedding.Query{
			Vector:    imageVec,
			SourceIDs: sourceIDSet,
			Kind:      embedding.KindImage,
			Limit:     100,
		})
		if err != nil {
			ctx.Log().Warn().Err(err).Msg("embed image search failed, skipping content search")

			return nil
		}

		embedHits = mergeContentHits(append(textHits, imageHits...), filesByMediaContentID)

		return nil
	})

	if err := g.Wait(); err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	}

	// Filter out home/trash from filename matches.
	filtered := fnMatches[:0]

	for _, m := range fnMatches {
		if m.FileID == ctx.Requester.HomeID || m.FileID == ctx.Requester.TrashID {
			continue
		}

		filtered = append(filtered, m)
	}

	fnMatches = filtered

	medias, err := getChildMedias(ctx, slices.Collect(maps.Values(candidates)))
	if err != nil {
		ctx.Error(http.StatusInternalServerError, wlerrors.Wrap(err, "failed to retrieve media information for search results"))

		return
	}

	results, err := mergeSearchResults(ctx, fnMatches, embedHits, candidates, medias, filenameSearch == "")
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	// Results are always sorted by match score descending, with filename as a deterministic tie-break.
	wlslices.SortFunc(results, func(a, b wlstructs.SearchResult) int {
		if slices.Contains(a.MatchKind, wlstructs.MatchKindFilename) && !slices.Contains(b.MatchKind, wlstructs.MatchKindFilename) {
			return -1
		} else if !slices.Contains(a.MatchKind, wlstructs.MatchKindFilename) && slices.Contains(b.MatchKind, wlstructs.MatchKindFilename) {
			return 1
		}

		switch {
		case a.Score > b.Score:
			return -1
		case a.Score < b.Score:
			return 1
		}

		p1, _ := wlfs.ParsePortable(a.File.PortablePath)
		p2, _ := wlfs.ParsePortable(b.File.PortablePath)

		return wlslices.NatSortCompare(p1.Filename(), p2.Filename())
	})

	ctx.JSON(http.StatusOK, results)
}
