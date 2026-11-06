package steam

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Rethunk-Tech/go-httpkit"
)

const (
	storeBrowse = "https://api.steampowered.com/IStoreBrowseService/GetItems/v1"
	storeAssets = "https://shared.steamstatic.com/store_item_assets/"

	storeTimeout  = 15 * time.Second
	maxStoreReply = 1 << 20
)

// StoreHero is the URL of the library hero (else the header) image Steam's store names for appID, or "" when the
// store names none or cannot be reached. Newer apps keep their art under hashed paths that only the store lists, so
// it complements HeroArt, which reads Steam's local cache. hc nil means http.DefaultClient.
func StoreHero(ctx context.Context, hc *http.Client, appID string) string {
	return storeHero(ctx, hc, storeBrowse, storeAssets, appID)
}

func storeHero(ctx context.Context, hc *http.Client, api, assets, appID string) string {
	id, err := strconv.Atoi(appID)
	if err != nil {
		return ""
	}
	input := fmt.Sprintf(`{"ids":[{"appid":%d}],"context":{"country_code":"US"},"data_request":{"include_assets":true}}`, id)
	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api+"?input_json="+url.QueryEscape(input), nil)
	if err != nil {
		return ""
	}
	raw, err := httpkit.DoCapped(hc, req, maxStoreReply, nil)
	if err != nil {
		return ""
	}
	var body struct {
		Response struct {
			StoreItems []struct {
				Assets struct {
					Format      string `json:"asset_url_format"`
					LibraryHero string `json:"library_hero"`
					Header      string `json:"header"`
				} `json:"assets"`
			} `json:"store_items"`
		} `json:"response"`
	}
	if json.Unmarshal(raw, &body) != nil || len(body.Response.StoreItems) == 0 {
		return ""
	}
	a := body.Response.StoreItems[0].Assets
	file := cmp.Or(a.LibraryHero, a.Header)
	if file == "" || !strings.Contains(a.Format, "${FILENAME}") {
		return ""
	}
	return assets + strings.Replace(a.Format, "${FILENAME}", file, 1)
}
