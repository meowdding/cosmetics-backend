package routes

import (
	"cosmetics/internal"
	"cosmetics/utils"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

const cosmeticRequest = `
	select data from cosmetics
`

const playerRequest = `
	select player_id as player, player_data as data, cosmetics FROM players_with_cosmetics
`

type PlayerType struct {
	Player    string                 `json:"uuid"`
	Data      map[string]interface{} `json:"extra_data"`
	Cosmetics []string               `json:"cosmetics"`
}

type Response struct {
	Players   []PlayerType  `json:"players"`
	Cosmetics []interface{} `json:"cosmetics"`
}

var cache = ""
var lastCreated time.Time

func GetEntries(ctx internal.RouteContext, res http.ResponseWriter, _ *http.Request) {
	if len(cache) != 0 && time.Now().Sub(lastCreated) < time.Second*5 {
		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("Cache-Control", "max-age=300")
		res.Header().Set("Age", strconv.Itoa(int(time.Now().Sub(lastCreated)/time.Second)))
		_, _ = io.WriteString(res, cache)
		return
	}

	cosmeticResult, err := ctx.Pool.Query(ctx.Context, cosmeticRequest)
	if err != nil {
		utils.PrintData(err)
		utils.PrintData(cosmeticResult)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer cosmeticResult.Close()

	playerResult, err := ctx.Pool.Query(ctx.Context, playerRequest)
	if err != nil {
		utils.PrintData(err)
		utils.PrintData(playerResult)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer playerResult.Close()

	var result = Response{}
	for cosmeticResult.Next() {
		var cosmetic = make(map[string]interface{})
		err := cosmeticResult.Scan(&cosmetic)
		if err != nil {
			continue
		}
		result.Cosmetics = append(result.Cosmetics, cosmetic)
	}
	if result.Cosmetics == nil {
		result.Cosmetics = make([]interface{}, 0)
	}

	list, err := pgx.CollectRows(playerResult, pgx.RowToStructByPos[PlayerType])

	result.Players = list

	tempCache, err := json.Marshal(result)
	if err != nil {
		utils.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	cache = string(tempCache)
	lastCreated = time.Now()
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", "max-age=300")
	_, _ = io.WriteString(res, cache)
}
