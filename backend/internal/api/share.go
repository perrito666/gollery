// share.go implements OpenGraph meta tag endpoints for social media sharing.
//
// The /share/assets/{id} and /share/albums/{id} routes return minimal HTML
// pages with OpenGraph and Twitter Card meta tags. These pages are designed
// to be consumed by social media crawlers (which don't execute JavaScript)
// and immediately redirect human visitors to the SPA.
//
// ACL checks use a nil principal (anonymous) so that only publicly visible
// content gets full OG metadata. Restricted content shows placeholder text
// ("Private photo" / "Private album") without leaking titles or thumbnails.
package api

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/perrito666/gollery/backend/internal/access"
)

// ogData holds the data used to render the OpenGraph HTML page.
type ogData struct {
	Title       string
	Description string
	ImageURL    string
	PageURL     string
	RedirectURL string
	SiteName    string
	Type        string
}

// ogTmpl is the parsed HTML template for OpenGraph pages.
var ogTmpl = template.Must(template.New("og").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta http-equiv="refresh" content="0;url={{.RedirectURL}}">
<meta property="og:title" content="{{.Title}}">
<meta property="og:description" content="{{.Description}}">
<meta property="og:type" content="{{.Type}}">
<meta property="og:url" content="{{.PageURL}}">
<meta property="og:site_name" content="{{.SiteName}}">
{{if .ImageURL}}<meta property="og:image" content="{{.ImageURL}}">
{{end}}<meta name="twitter:card" content="{{if .ImageURL}}summary_large_image{{else}}summary{{end}}">
<meta name="twitter:title" content="{{.Title}}">
<meta name="twitter:description" content="{{.Description}}">
{{if .ImageURL}}<meta name="twitter:image" content="{{.ImageURL}}">
{{end}}<title>{{.Title}}</title>
</head>
<body>
<p>Redirecting&hellip;</p>
<script>window.location.replace({{.RedirectURL}});</script>
</body>
</html>
`))

// handleShareAsset serves an OG meta tag page for a single asset.
func (s *Server) handleShareAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s.mu.RLock()
	defer s.mu.RUnlock()

	asset, ok := s.assetsByID[id]
	if !ok {
		renderNotFoundOG(w, r)
		return
	}

	album, ok := s.snapshot.Albums[asset.AlbumPath]
	if !ok {
		renderNotFoundOG(w, r)
		return
	}

	baseURL := resolvePublicBaseURL(r)
	redirectURL := baseURL + "/#/assets/" + id

	// Check ACL against nil principal (anonymous).
	var albumACL = effectiveAlbumACL(s.configs, album.Path)
	effectiveACL := access.EffectiveAssetACL(albumACL, asset.Access)

	if access.CheckView(effectiveACL, nil) == access.Deny {
		// Restricted asset — show placeholder without leaking info.
		data := ogData{
			Title:       "Private photo",
			Description: "This photo is not publicly available.",
			PageURL:     baseURL + "/share/assets/" + id,
			RedirectURL: redirectURL,
			SiteName:    "gollery",
			Type:        "article",
		}
		renderOGPage(w, data, http.StatusOK)
		return
	}

	displayTitle := asset.Title
	if displayTitle == "" {
		displayTitle = asset.Filename
	}
	description := asset.Description
	if description == "" {
		description = "Photo from " + album.Title
	}

	imageURL := baseURL + "/api/v1/assets/" + id + "/preview?size=1200"

	data := ogData{
		Title:       displayTitle,
		Description: description,
		ImageURL:    imageURL,
		PageURL:     baseURL + "/share/assets/" + id,
		RedirectURL: redirectURL,
		SiteName:    "gollery",
		Type:        "article",
	}
	renderOGPage(w, data, http.StatusOK)
}

// handleShareAlbum serves an OG meta tag page for an album.
func (s *Server) handleShareAlbum(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s.mu.RLock()
	defer s.mu.RUnlock()

	album, ok := s.albumsByID[id]
	if !ok {
		renderNotFoundOG(w, r)
		return
	}

	baseURL := resolvePublicBaseURL(r)
	redirectURL := baseURL + "/#/albums/" + id

	// Check album ACL against nil principal (anonymous).
	albumACL := effectiveAlbumACL(s.configs, album.Path)
	if access.CheckView(albumACL, nil) == access.Deny {
		data := ogData{
			Title:       "Private album",
			Description: "This album is not publicly available.",
			PageURL:     baseURL + "/share/albums/" + id,
			RedirectURL: redirectURL,
			SiteName:    "gollery",
			Type:        "website",
		}
		renderOGPage(w, data, http.StatusOK)
		return
	}

	description := album.Description
	if description == "" {
		description = "Photo album"
	}

	// Find first public asset for og:image.
	var imageURL string
	for _, ast := range album.Assets {
		astACL := access.EffectiveAssetACL(albumACL, ast.Access)
		if access.CheckView(astACL, nil) == access.Allow {
			imageURL = baseURL + "/api/v1/assets/" + ast.ID + "/thumbnail?size=1200"
			break
		}
	}

	data := ogData{
		Title:       album.Title,
		Description: description,
		ImageURL:    imageURL,
		PageURL:     baseURL + "/share/albums/" + id,
		RedirectURL: redirectURL,
		SiteName:    "gollery",
		Type:        "website",
	}
	renderOGPage(w, data, http.StatusOK)
}

// resolvePublicBaseURL derives the public base URL from the request headers.
// It uses X-Forwarded-Proto for the scheme (defaulting to "http") and r.Host.
// Only "http" and "https" are accepted; anything else defaults to "http".
func resolvePublicBaseURL(r *http.Request) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme != "https" {
		scheme = "http"
	}
	return scheme + "://" + r.Host
}

// renderOGPage writes an HTML page with OG meta tags.
func renderOGPage(w http.ResponseWriter, data ogData, statusCode int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := ogTmpl.Execute(w, data); err != nil {
		slog.Error("rendering OG page", "error", err)
	}
}

// renderNotFoundOG writes a 404 OG page.
func renderNotFoundOG(w http.ResponseWriter, r *http.Request) {
	baseURL := resolvePublicBaseURL(r)
	data := ogData{
		Title:       "Page not found",
		Description: "The requested page does not exist.",
		PageURL:     baseURL + r.URL.Path,
		RedirectURL: baseURL + "/",
		SiteName:    "gollery",
		Type:        "website",
	}
	renderOGPage(w, data, http.StatusNotFound)
}
